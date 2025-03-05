package codeGenerators

import (
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
)

type CodeGenerator interface {
	GetName() string
	GetVersion() string
	SetActiveTypeFile(fileData fieldData.FileData)

	GetFieldType(path string, withoutArray bool) (expr ast.Expr)
	IsStruct(path string) bool
	IsPointer(path string) bool
	GetType(path string) fieldData.Type
	GetLevelOfArrays(path string) int
	CheckType(path string) error
	IsBasicType(path string) bool
	IsBasicTypeWhitDetails(path string) (bool, int, fieldData.Type)
}
