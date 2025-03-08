package fieldData

import (
	"errors"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"go/ast"
	"go/token"
	"strings"
)

type FilePath string

func (f FilePath) IsFile() (bool, error) {
	if len(f) == 0 {
		return false, errors.New("file path is empty")
	}
	if f[0] != '/' {
		return true, nil
	} else {
		return false, nil
	}
}

func (f FilePath) AddInFront(filePath FilePath) (FilePath, error) {
	isFile, err := f.IsFile()
	if err != nil {
		return "", err
	} else if isFile {
		return "", errors.New("file path is a file, not a path")
	}
	return FilePath(strings.ReplaceAll(string(filePath+f), "//", "")), nil
}

func (f FilePath) AddInBack(filePath FilePath) (FilePath, error) {
	isFile, err := filePath.IsFile()
	if err != nil {
		return "", err
	} else if isFile {
		return "", errors.New("file path is a file, not a path")
	}
	return FilePath(strings.ReplaceAll(string(filePath+f), "//", "")), nil
}

type File struct {
	File           *ast.File         `json:"file,omitempty"`
	FSet           *token.FileSet    `json:"fSet,omitempty"`
	Imports        []string          `json:"extraImports,omitempty"`
	FileEnding     *string           `json:"fileType,omitempty"`
	ModFileContent []*ModFileContent `json:"modFileContent,omitempty"`
}

func GetEmptyGoFile() *File {
	f, fs := utils.GetEmptyAstFile("placeholder")
	return &File{
		File:           f,
		FSet:           fs,
		Imports:        nil,
		FileEnding:     utils.StringToPointer("go"),
		ModFileContent: nil,
	}
}

func GetGoFile(decls []ast.Decl, imports []string, mfc []*ModFileContent) *File {
	f, fs := utils.GetEmptyAstFile("placeholder")
	f.Decls = decls
	return &File{
		File:           f,
		FSet:           fs,
		Imports:        imports,
		FileEnding:     utils.StringToPointer("go"),
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

func (i Imports) AdjustPathsIfNeeded(path string) {
	for _, im := range i {
		im.AdjustPathIfNeeded(path)
	}
}

type Import struct {
	Path            string  `json:"name"`
	Alias           *string `json:"alias"`
	NeedsAdjustment bool    `json:"needsAdjustment"`
}

func (i *Import) AdjustPathIfNeeded(path string) {
	if !i.NeedsAdjustment {
		return
	}
	i.Path = path + i.Path
}
