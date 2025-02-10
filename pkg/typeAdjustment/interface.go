package typeAdjustment

import (
	"encoding/json"
	"go/ast"
)

type TypeDeterminationFunction interface {
	CouldTypeBeApplied(seenValues map[string]string) bool
	GetType() ast.Expr
	GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error)
	GenerateToTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error)
	GetRequiredImports() []string
	SetFile(file *ast.File)
	GetName() string
	SetState(state json.RawMessage) error
	GetState() (json.RawMessage, error)
}
