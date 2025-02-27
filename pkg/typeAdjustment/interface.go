package typeAdjustment

import (
	"encoding/json"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
)

type TypeDeterminationFunction interface {
	CouldTypeBeApplied(path string) (State, error)
	GetType() ast.Expr
	GenerateMarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error)
	GenerateUnmarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error)
	GetRequiredImports() []string
	GetName() string
	SetState(states []json.RawMessage, currentPath string, activeTypeCheckers []TypeDeterminationFunction, seenTypes map[string]*fieldData.PathData) error
	GetState() ([]json.RawMessage, error)
	GetExtraCode() []ast.Decl
	TypeExpansion() bool
	ForceSourceType() *string
	GetModFileContents() []*fieldData.ModFileContent
	GetVersion() string
}

type State int

const (
	StateFailed State = iota
	StateUndecided
	StateApplicable
)
