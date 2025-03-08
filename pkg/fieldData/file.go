package fieldData

import (
	"errors"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"go/ast"
	"go/token"
	"strings"
)

type Path string

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

func (p Path) GetParentPath() Path {
	pathElements := strings.Split(string(p), ".")
	if len(pathElements) > 1 {
		return Path(strings.Join(pathElements[:len(pathElements)-1], "."))
	} else {
		return p
	}
}

func (p Path) GetParentFieldName() string {
	pathElements := strings.Split(string(p), ".")
	return pathElements[len(pathElements)-2]
}

func (p Path) GetFieldName() string {
	pathElements := strings.Split(string(p), ".")
	return pathElements[len(pathElements)-1]
}

type File struct {
	File           *ast.File         `json:"file,omitempty"`
	FSet           *token.FileSet    `json:"fSet,omitempty"`
	Imports        Imports           `json:"extraImports,omitempty"`
	FileEnding     string            `json:"fileType,omitempty"`
	ModFileContent []*ModFileContent `json:"modFileContent,omitempty"`
	Name           string
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

func GetGoFile(decls []ast.Decl, imports Imports, mfc []*ModFileContent) *File {
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
}

func (i *Import) AdjustPathIfNeeded(path Path) {
	if !i.NeedsAdjustment {
		return
	}
	i.Path.Prepend(path)
}
