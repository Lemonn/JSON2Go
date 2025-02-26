package codeGen

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Lemonn/AstUtils"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/jsonMarshallerGenerators/marshaller"
	"github.com/Lemonn/JSON2Go/pkg/jsonMarshallerGenerators/unmarshaller"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
	"time"
)

type FileData struct {
	file *ast.File
	fSet *token.FileSet
}

type Generator struct {
	seenTypes         map[string]*fieldData.PathData
	files             map[string]*FileData
	stackedMarshaller map[string]*ast.File
	basePath          *string
	moduleName        *string
	typeAdjusters     []typeAdjustment.TypeDeterminationFunction
	outputPath        string
	seenTypesUtils    *utils.SeenTypeUtils
	startTime         time.Time
}

func NewCodeGenerator(inputFile map[string]*fieldData.PathData, outputPath string, clearPath bool, moduleName *string, basePath *string, adjuster []typeAdjustment.TypeDeterminationFunction) (*Generator, error) {
	if moduleName == nil && basePath == nil {
		return nil, errors.New("either packageName or basePath must be specified")
	}
	/*
		var seenTypes map[string]*fieldData.PathData
		err := json.Unmarshal(inputFile, &seenTypes)
		if err != nil {
			return nil, err
		}

	*/
	return &Generator{
		seenTypes:         inputFile,
		files:             make(map[string]*FileData),
		stackedMarshaller: map[string]*ast.File{},
		basePath:          basePath,
		moduleName:        moduleName,
		typeAdjusters:     adjuster,
		outputPath:        outputPath,
		seenTypesUtils:    utils.NewSeenTypeUtils(inputFile),
		startTime:         time.Now(),
	}, nil
}

func (s *Generator) Generate() error {
	startPath, err := s.getStartPath()
	if err != nil {
		return err
	}

	pathsToProcess := []string{startPath}

	for {
		for _, path := range pathsToProcess {
			var file *ast.File
			if _, ok := s.files[path]; !ok {
				fset := token.NewFileSet()
				pathElements := strings.Split(path, ".")
				file, err = parser.ParseFile(fset, "", "package "+pathElements[len(pathElements)-1], parser.ParseComments)
				if err != nil {
					return err
				}
				s.files[path] = &FileData{
					file: file,
					fSet: fset,
				}
				file = s.files[path].file
			}

			file = s.files[path].file
			if s.seenTypesUtils.IsStruct(path) {
				var fields []*ast.Field
				var levelOfArrays int
				for levelOfArrays, _ = range s.seenTypes[path].Types["field"] {
					break
				}
				if len(s.typeAdjusters) > 0 {
					s.stackedMarshaller[path] = file
				}
				for fieldPath, _ := range s.seenTypes[path].Types["field"][levelOfArrays] {
					//Adjust Types
					ta := typeAdjustment.NewTypeAdjuster(s.seenTypes, s.typeAdjusters, s.startTime)
					err := ta.AdjustTypesNew(fieldPath)
					if err != nil {
						return err
					}

					expr, err := s.seenTypesUtils.GetAdjustedFieldType(fieldPath)
					if err != nil {
						return err
					}
					if s.seenTypesUtils.IsStruct(fieldPath) {
						pathsToProcess = append(pathsToProcess, fieldPath)
						AstUtils.AddMissingImports(file, []string{strings.ReplaceAll(*s.basePath+"/"+fieldPath, ".", "/")})
					}

					pathElements := strings.Split(fieldPath, ".")
					fields = append(fields, &ast.Field{
						Names: []*ast.Ident{{Name: pathElements[len(pathElements)-1]}},
						Type:  expr,
						//TODO set path correctly. Do not set a path if JsonFieldName == "". Write a function, that checks if omitempty should be set
						Tag: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("`json:\"%s,omitempty\"`", s.seenTypes[fieldPath].JsonFieldName)},
					})
				}
				pathElements := strings.Split(path, ".")
				file.Decls = append(file.Decls, &ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: pathElements[len(pathElements)-1],
							},
							Type: func() ast.Expr {
								return &ast.StructType{
									Fields: &ast.FieldList{List: fields},
								}
							}(),
						},
					},
				})
			} else {
				file.Decls = append(file.Decls, &ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: path,
							},
							Type: s.seenTypesUtils.GetFieldType(path, false),
						},
					},
				})
			}
			pathsToProcess = pathsToProcess[1:]
		}
		if len(pathsToProcess) == 0 {
			break
		}
	}

	err = s.addJSONMarshaller()
	if err != nil {
		return err
	}

	err = s.writeFiles()
	if err != nil {
		return err
	}
	return nil
}

func (s *Generator) addJSONMarshaller() error {
	for path, file := range s.stackedMarshaller {
		uGen := unmarshaller.NewGenerator(s.seenTypes)
		generate, _, err := uGen.Generate(path)
		if err != nil {
			return err
		}
		file.Decls = append(file.Decls, generate...)

		mGen := marshaller.NewGenerator(s.seenTypes)
		gen, _, err := mGen.Generate(path)
		if err != nil {
			return err
		}
		file.Decls = append(file.Decls, gen...)

		err = s.addTypeConverterFunctions(path, file)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Generator) generateGoMod() {

}

func (s *Generator) addTypeConverterFunctions(path string, file *ast.File) error {
	var elements map[string]*fieldData.ValueDetails
	for level, _ := range s.seenTypes[path].Types[s.seenTypesUtils.GetType(path)] {
		elements = s.seenTypes[path].Types[s.seenTypesUtils.GetType(path)][level]
	}

	for elementPath, _ := range elements {
		if s.seenTypes[elementPath].TypeAdjusterData != nil && s.seenTypes[elementPath].TypeAdjusterData.ParseFunctions != nil {
			MarshallExprFile, err := parser.ParseFile(token.NewFileSet(), "", "package main \n"+s.seenTypes[elementPath].TypeAdjusterData.ParseFunctions.Marshall, parser.AllErrors)
			if err != nil {
				return err
			}
			file.Decls = append(file.Decls, MarshallExprFile.Decls[0])

			UnMarshallExprFile, err := parser.ParseFile(token.NewFileSet(), "", "package main \n"+s.seenTypes[elementPath].TypeAdjusterData.ParseFunctions.Unmarshall, parser.AllErrors)
			if err != nil {
				return err
			}
			file.Decls = append(file.Decls, UnMarshallExprFile.Decls[0])
		}
	}
	return nil
}

func (s *Generator) getStartPath() (string, error) {
	for path, _ := range s.seenTypes {
		pathElements := strings.Split(path, ".")
		if len(pathElements) == 1 {
			return pathElements[0], nil
		}
	}
	return "", errors.New("start path not found")
}

func (s *Generator) writeFiles() error {
	for path, file := range s.files {
		err := os.MkdirAll(strings.ReplaceAll(s.outputPath+"/"+path, ".", "/"), os.ModePerm)
		if err != nil {
			return err
		}
		output := bytes.NewBuffer([]byte{})
		if err := printer.Fprint(output, file.fSet, file.file); err != nil {
			return err
		}
		c, err := format.Source(output.Bytes())
		if err != nil {
			return err
		}
		pathElements := strings.Split(path, ".")
		err = os.WriteFile(strings.ReplaceAll(s.outputPath+"/"+path, ".", "/")+"/"+pathElements[len(pathElements)-1]+".go", c, 0666)
		if err != nil {
			return err
		}
	}
	return nil
}
