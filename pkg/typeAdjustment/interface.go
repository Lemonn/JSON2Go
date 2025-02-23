package typeAdjustment

import (
	"encoding/json"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
)

type TypeDeterminationFunction interface {
	CouldTypeBeApplied(seenTypes map[fieldData.Type]map[int]map[string]*fieldData.ValueDetails) (State, error)
	GetType() ast.Expr
	GenerateMarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error)
	GenerateUnmarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error)
	GetRequiredImports() []string
	GetName() string
	SetState(state []*json.RawMessage, currentPath string) error
	GetState() ([]*json.RawMessage, error)
	GetExtraCode() []ast.Decl
	TypeExpansion() bool
}

type State int

const (
	StateFailed State = iota
	StateUndecided
	StateApplicable
)
