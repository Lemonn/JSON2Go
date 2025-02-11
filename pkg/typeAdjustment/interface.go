package typeAdjustment

import (
	"encoding/json"
	"go/ast"
)

type TypeDeterminationFunction interface {
	CouldTypeBeApplied(seenValues map[string]string) State
	GetType() ast.Expr
	GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error)
	GenerateToTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error)
	GetRequiredImports() []string
	SetFile(file *ast.File)
	GetName() string
	SetState(state json.RawMessage, currentPath string) error
	GetState() (json.RawMessage, error)
}

type State int

const (
	StateFailed State = iota
	StateUndecided
	StateApplicable
)
