package codeGenerators

import (
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"golang.org/x/mod/modfile"
	"time"
)

type CodeGenerator interface {
	GetName() string
	GetVersion() string
	SetActiveTypeFile(fileData fieldData.FileData)
	Generate() (map[fieldData.Path][]*fieldData.File, error)
	GenerateGoMod(name string) (*modfile.File, error)
	//StoreState() (uuid.UUID, error)
	//ResetState(stateID uuid.UUID) error
	//DeleteState(stateID uuid.UUID) error
	//GetRegisteredTypeCheckers() typeAdjustment.TypeDeterminationFunctions
	GetStartTime() time.Time

	//Clone could replace StoreState and ResetState by simply cloning the CodeGenerator
	Clone(fileData fieldData.FileData) CodeGenerator

	GetOriginalFieldType(path fieldData.Path) (expr ast.Expr, imports fieldData.Imports, err error)
	IsStruct(path fieldData.Path) bool
	IsPointer(path fieldData.Path) bool
	GetType(path fieldData.Path) fieldData.Type
	GetLevelOfArrays(path fieldData.Path) int
	CheckType(path fieldData.Path) error
	IsBasicType(path fieldData.Path) bool
	IsBasicTypeWhitDetails(path fieldData.Path) (bool, int, fieldData.Type)
}
