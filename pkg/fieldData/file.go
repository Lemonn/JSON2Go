package fieldData

import (
	"errors"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"go/ast"
	"go/token"
	"strings"
	"unicode"
)

type Path string

func NewPath(s string) (Path, error) {
	return Path(s), nil
}

func (p Path) PrependElement(element string) Path {
	return Path(element + "." + string(p))
}

func (p Path) AppendElement(element string) Path {
	return Path(string(p) + "." + element)
}

func (p Path) Prepend(path Path) Path {
	return path + "." + p
}

func (p Path) Append(path Path) Path {
	return p + "." + path
}

func (p Path) IsRootPath() bool {
	if len(strings.Split(string(p), ".")) == 1 {
		return true
	}
	return false
}

func (p Path) GetRoot() Path {
	return Path(strings.Split(string(p), ".")[0])
}

func (p Path) IsGlobal() bool {
	if len(strings.Split(string(p), ".")) == 1 {
		return true
	}
	return false
}

func (p Path) GetParentPath() Path {
	pathElements := strings.Split(string(p), ".")
	if len(pathElements) > 1 {
		return Path(strings.Join(pathElements[:len(pathElements)-1], "."))
	} else {
		return p
	}
}

func (p Path) GetFilePath() string {
	return strings.ReplaceAll(p.String(), ".", "/")
}

func (p Path) GetParentFieldName() string {
	pathElements := strings.Split(string(p), ".")
	return pathElements[len(pathElements)-2]
}

func (p Path) GetFieldName() string {
	pathElements := strings.Split(string(p), ".")
	return pathElements[len(pathElements)-1]
}

func (p Path) GetRune() string {
	return string(unicode.ToLower([]rune(p.GetFieldName())[0]))
}

func (p Path) GetParentRune() string {
	return string(unicode.ToLower([]rune(p.GetParentPath())[0]))
}

func (p Path) String() string {
	return string(p)
}

type FileClass int

const (
	FileClassLocal FileClass = iota
	FileClassGlobal
	FileClassMarshallFunction
	FileClassUnMarshallFunction
	FileClassMarshaller
	FileClassUnMarshaller
)

type File struct {
	File           *ast.File         `json:"file,omitempty"`
	FSet           *token.FileSet    `json:"fSet,omitempty"`
	Imports        Imports           `json:"extraImports,omitempty"`
	FileEnding     string            `json:"fileType,omitempty"`
	ModFileContent []*ModFileContent `json:"modFileContent,omitempty"`
	Name           *string           `json:"name,omitempty"`
	FileClass      FileClass         `json:"fileClass,omitempty"`
}

func (f *File) WriteImportsToFile(globalPath Path, packageOffset string) error {
	var alias map[string]struct{}
	imports := make(map[string]ast.Spec)

	if packageOffset[len(packageOffset)-1:] != "/" {
		packageOffset += "/"
	}

	for _, i := range f.Imports {
		if i == nil {
			continue
		}
		filePath := i.Path.GetFilePath()
		if i.NeedsAdjustment {
			filePath = packageOffset + filePath
		} else if i.IsGlobal {
			filePath = i.Path.Prepend(globalPath).GetFilePath()
		}
		if i.Alias != nil {
			if _, ok := alias[*i.Alias]; ok {
				return errors.New("duplicate import alias")
			}
		}

		if _, ok := imports[i.Path.String()]; !ok {
			imports[i.Path.String()] = &ast.ImportSpec{
				Name: func() *ast.Ident {
					if i.Alias != nil {
						return &ast.Ident{
							Name: *i.Alias,
						}
					}
					return nil
				}(),
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"" + filePath + "\"",
				},
			}
		}
	}
	if len(imports) > 0 {
		var i []ast.Spec
		for _, spec := range imports {
			i = append(i, spec)
		}
		f.File.Decls = append([]ast.Decl{&ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: i,
		},
		}, f.File.Decls...)
	}
	return nil
}

func (f *File) combine(f1 *File) (*File, error) {

	//TODO correctly combine values and error on incompatible
	var decls []ast.Decl
	var imports Imports
	var mfc []*ModFileContent

	if f.FileEnding != f1.FileEnding {
		return nil, errors.New("file endings do not match")
	}

	decls = append(decls, f.File.Decls...)
	decls = append(decls, f1.File.Decls...)

	imports = append(imports, f.Imports...)
	imports = append(imports, f1.Imports...)

	mfc = append(mfc, f.ModFileContent...)
	mfc = append(mfc, f1.ModFileContent...)

	return GetGoFile(decls, imports, mfc, f.FileClass), nil
}

func GetEmptyGoFile() *File {
	f, fs := utils.GetEmptyAstFile("placeholder")
	return &File{
		File:           f,
		FSet:           fs,
		Imports:        Imports{},
		FileEnding:     "go",
		ModFileContent: []*ModFileContent{},
	}
}

func GetGoFile(decls []ast.Decl, imports Imports, mfc []*ModFileContent, class FileClass) *File {
	f, fs := utils.GetEmptyAstFile("placeholder")
	f.Decls = decls
	return &File{
		File:           f,
		FSet:           fs,
		Imports:        imports,
		FileEnding:     "go",
		ModFileContent: mfc,
	}
}

func (f *File) SetPackage(p string) {
	f.File.Name = &ast.Ident{Name: p}
}

func (f *File) GetPackage() (string, error) {
	if f.File.Name == nil {
		return "", errors.New("no package set")
	}
	return f.File.Name.Name, nil
}

func (f *File) GetDecls() ([]ast.Decl, error) {
	return f.File.Decls, nil
}

func (f *File) EnrichFileWhitImports() {

}

func (f *File) RemoveImportsFromFile() {

}

type ModFileContent struct {
	Path     string `json:"path"`
	Version  string `json:"version"`
	Indirect bool   `json:"indirect"`
}

type Imports []*Import

func (i Imports) AdjustPathsIfNeeded(path Path) {
	for _, im := range i {
		im.AdjustPathIfNeeded(path)
	}
}

type Import struct {
	Path            Path    `json:"name"`
	Alias           *string `json:"alias"`
	NeedsAdjustment bool    `json:"needsAdjustment"`
	IsGlobal        bool    `json:"isGlobal"`
}

func (i *Import) AdjustPathIfNeeded(path Path) {
	if !i.NeedsAdjustment {
		return
	}
	i.Path.Prepend(path)
}
