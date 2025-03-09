package codeGenerators

import (
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"golang.org/x/mod/modfile"
)

type CodeGenerator interface {
	GetName() string
	GetVersion() string
	SetActiveTypeFile(fileData fieldData.FileData)
	Generate() (map[fieldData.Path][]*fieldData.File, []*fieldData.File, error)
	GenerateGoMod(name string) (*modfile.File, error)
	//StoreState() (uuid.UUID, error)
	//ResetState(stateID uuid.UUID) error
	//DeleteState(stateID uuid.UUID) error
	//GetRegisteredTypeCheckers() typeAdjustment.TypeDeterminationFunctions

	//Clone could replace StoreState and ResetState by simply cloning the CodeGenerator
	Clone() CodeGenerator

	GetFieldType(path fieldData.Path, withoutArray bool) (expr ast.Expr)
	IsStruct(path fieldData.Path) bool
	IsPointer(path fieldData.Path) bool
	GetType(path fieldData.Path) fieldData.Type
	GetLevelOfArrays(path fieldData.Path) int
	CheckType(path fieldData.Path) error
	IsBasicType(path fieldData.Path) bool
	IsBasicTypeWhitDetails(path fieldData.Path) (bool, int, fieldData.Type)
}
