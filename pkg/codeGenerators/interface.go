package codeGenerators

import (
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/google/uuid"
	"go/ast"
)

type CodeGenerator interface {
	GetName() string
	GetVersion() string
	SetActiveTypeFile(fileData fieldData.FileData)
	Generate() (map[string]*fieldData.File, error)
	StoreState() (uuid.UUID, error)
	ResetState(stateID uuid.UUID) error
	DeleteState(stateID uuid.UUID) error
	//GetRegisteredTypeCheckers() typeAdjustment.TypeDeterminationFunctions

	//Clone could replace StoreState and ResetState by simply cloning the CodeGenerator
	Clone() CodeGenerator

	GetFieldType(path string, withoutArray bool) (expr ast.Expr)
	IsStruct(path string) bool
	IsPointer(path string) bool
	GetType(path string) fieldData.Type
	GetLevelOfArrays(path string) int
	CheckType(path string) error
	IsBasicType(path string) bool
	IsBasicTypeWhitDetails(path string) (bool, int, fieldData.Type)
}
