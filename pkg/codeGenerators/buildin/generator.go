package buildin

import (
	"errors"
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	j2gErrors "github.com/Lemonn/JSON2Go/pkg/errors/combiner"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/jsonMarshallerGenerators"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"go/ast"
	"go/token"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
	"time"
)

type Generator struct {
	fileData          fieldData.FileData
	files             map[fieldData.Path][]*fieldData.File
	stackedMarshaller map[fieldData.Path][]*fieldData.File
	typeAdjuster      *typeAdjustment.TypeAdjuster
	globalFiles       []*fieldData.File
	*codeGenerators.Common
	startTime         time.Time
	version           string
	name              string
	marshallGenerator jsonMarshallerGenerators.Generator
	completed         bool
	goVersion         string
}

func NewGenerator(fileData fieldData.FileData, adjuster typeAdjustment.TypeDeterminationFunctions) (*Generator, error) {
	s := &Generator{
		fileData:          fileData,
		files:             make(map[fieldData.Path][]*fieldData.File),
		stackedMarshaller: make(map[fieldData.Path][]*fieldData.File),
		Common:            codeGenerators.NewCommon(fileData),
		startTime:         time.Now(),
		version:           "v0.0.1",
		name:              "Test",
	}

	s.typeAdjuster = typeAdjustment.NewTypeAdjuster(fileData, s, adjuster, s.startTime)
	return s, nil
}

func (g *Generator) GetName() string {
	return g.name
}

func (g *Generator) GetVersion() string {
	return g.version
}

func (g *Generator) Clone() codeGenerators.CodeGenerator {
	return &Generator{
		fileData:          make(fieldData.FileData),
		files:             make(map[fieldData.Path][]*fieldData.File),
		stackedMarshaller: make(map[fieldData.Path][]*fieldData.File),
		typeAdjuster:      g.typeAdjuster,
		globalFiles:       []*fieldData.File{},
		Common:            g.Common,
		startTime:         g.startTime,
		version:           g.version,
		name:              g.name,
		marshallGenerator: g.marshallGenerator,
		completed:         false,
		goVersion:         g.goVersion,
	}
}

func (g *Generator) appendFileAtPath(path fieldData.Path, file *fieldData.File) {
	if _, ok := g.files[path]; ok {
		g.files[path] = append(g.files[path], file)
	} else {
		g.files[path] = []*fieldData.File{file}
	}
}

func (g *Generator) appendFilesAtPath(path fieldData.Path, files []*fieldData.File) {
	for _, file := range files {
		g.appendFileAtPath(path, file)
	}
}

func (g *Generator) createFileAtPath(path fieldData.Path) *fieldData.File {
	file := fieldData.GetEmptyGoFile()
	if _, ok := g.files[path]; ok {
		g.files[path] = append(g.files[path], file)
	} else {
		g.files[path] = []*fieldData.File{file}
	}
	return file
}

func (g *Generator) Generate() (map[fieldData.Path][]*fieldData.File, []*fieldData.File, error) {
	var currentPath fieldData.Path
	startPath, err := g.getStartPath()
	if err != nil {
		return nil, nil, err
	}
	pathsToProcess := []fieldData.Path{startPath}
	for {
		if len(pathsToProcess) == 0 {
			break
		}
		path := pathsToProcess[0]

		g.fileData[path].Active = true
		if g.IsStruct(path) {
			var fields []*ast.Field
			var levelOfArrays int
			for levelOfArrays, _ = range g.fileData[path].Types["field"] {
				break
			}

			structFile := g.createFileAtPath(path)
			for fieldPath, _ := range g.fileData[path].Types["field"][levelOfArrays] {
				fp, err := fieldData.NewPath(fieldPath)
				//Adjust Types
				subFiles, replacementType, replacementImport, err := g.typeAdjuster.AdjustType(fp)
				if err != nil {
					return nil, nil, err
				}
				g.stackedMarshaller[path] = g.files[path]

				for filePath, f := range subFiles {
					g.appendFileAtPath(currentPath.Append(filePath), f)
				}

				var expr ast.Expr
				if replacementType != nil {
					expr = replacementType
				} else {
					expr, err = g.GetAdjustedFieldType(fp)
					if err != nil {
						return nil, nil, err
					}
				}

				//If it's a struct and has not been replaced by a custom type, we need to import the type.
				if g.IsStruct(fp) && replacementType == nil {
					pathsToProcess = append(pathsToProcess, fp)
					structFile.Imports = append(structFile.Imports, &fieldData.Import{
						Path:            path,
						Alias:           nil,
						NeedsAdjustment: true,
					})
				} else if replacementType != nil {
					structFile.Imports = append(structFile.Imports, replacementImport)
				}

				g.fileData[fp].Active = true
				field := &ast.Field{
					Names: []*ast.Ident{{Name: fp.GetFieldName()}},
					Type:  expr,
					Tag:   &ast.BasicLit{Kind: token.STRING, Value: g.GetJsonTag(fp)},
				}
				fields = append(fields, field)
			}

			structFile.File.Decls = append(structFile.File.Decls, &ast.GenDecl{
				Tok: token.TYPE,
				Specs: []ast.Spec{
					&ast.TypeSpec{
						Name: &ast.Ident{
							Name: path.GetFieldName(),
						},
						Type: &ast.StructType{
							Fields: &ast.FieldList{List: fields},
						},
					},
				},
			})
		} else {
			//TODO handle anonymous array
			/*
				file.Decls = append(file.Decls, &ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: path,
							},
							Type: g.GetFieldType(path, false),
						},
					},
				})

			*/
		}
		pathsToProcess = pathsToProcess[1:]
	}

	err = g.addJSONMarshaller()
	if err != nil {
		return nil, nil, err
	}
	return g.files, g.globalFiles, nil
}

func (g *Generator) addJSONMarshaller() error {
	for path, _ := range g.stackedMarshaller {
		marshall, err := g.marshallGenerator.Marshall(path)
		if err != nil {
			return err
		}
		g.appendFilesAtPath(path, marshall)

		unmarshall, err := g.marshallGenerator.Unmarshall(path)
		if err != nil {
			return err
		}
		g.appendFilesAtPath(path, unmarshall)
	}
	g.globalFiles = append(g.globalFiles, g.marshallGenerator.GlobalFiles()...)
	return nil
}

func (g *Generator) getStartPath() (fieldData.Path, error) {
	for path, _ := range g.fileData {
		if _, ok := g.fileData[path.GetRoot()]; ok {
			return path.GetRoot(), nil
		}
		break
	}
	return "", errors.New("start path not found")
}

func (g *Generator) CheckType(path fieldData.Path) error {
	pathData := g.fileData[path]
	if pathData.TypeAdjusterData != nil && pathData.TypeAdjusterData.NameOfActiveTypeAdjuster != nil {
		return g.typeAdjuster.CheckActiveChecker(path, true)
	} else if pathData.ActiveType != nil {
		typeString, err := utils.ExprToString(g.GetFieldType(path, true))
		if err != nil {
			return err
		}
		if *pathData.ActiveType != typeString {
			return &j2gErrors.TypeChangeError{
				OldType: *pathData.ActiveType,
				NewType: typeString,
			}
		}
	}
	return nil
}

func (g *Generator) GenerateGoMod(name string) (*modfile.File, error) {
	requiresMap := make(map[string]*fieldData.ModFileContent)
	if !g.completed {
		return nil, errors.New("generateGoMod() called before Generate()")
	}

	modFile, err := modfile.Parse("", []byte("module "+name+"\n\ngo "+g.goVersion+"\n\n"), nil)
	if err != nil {
		return nil, err
	}

	for _, files := range g.files {
		for _, file := range files {
			for _, modFileContent := range file.ModFileContent {
				if !semver.IsValid(modFileContent.Version) {
					return nil, fmt.Errorf("invalid semver version: %s", modFileContent.Version)
				}
				if v, ok := requiresMap[modFileContent.Path]; !ok {
					requiresMap[modFileContent.Path] = modFileContent
				} else {
					if semver.Compare(v.Version, modFileContent.Version) == -1 {
						requiresMap[modFileContent.Path].Version = modFileContent.Version
					}
					if !requiresMap[modFileContent.Path].Indirect || !modFileContent.Indirect {
						requiresMap[modFileContent.Path].Indirect = false
					}
				}
			}
		}
	}

	for _, file := range g.globalFiles {
		for _, modFileContent := range file.ModFileContent {
			if !semver.IsValid(modFileContent.Version) {
				return nil, fmt.Errorf("invalid semver version: %s", modFileContent.Version)
			}
			if v, ok := requiresMap[modFileContent.Path]; !ok {
				requiresMap[modFileContent.Path] = modFileContent
			} else {
				if semver.Compare(v.Version, modFileContent.Version) == -1 {
					requiresMap[modFileContent.Path].Version = modFileContent.Version
				}
				if !requiresMap[modFileContent.Path].Indirect || !modFileContent.Indirect {
					requiresMap[modFileContent.Path].Indirect = false
				}
			}
		}
	}

	var requires []*modfile.Require
	for _, content := range requiresMap {
		requires = append(requires, &modfile.Require{
			Mod: module.Version{
				Path:    content.Path,
				Version: content.Version,
			},
			Indirect: content.Indirect,
		})
	}
	modFile.SetRequireSeparateIndirect(requires)
	return modFile, nil
}
