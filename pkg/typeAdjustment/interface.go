package typeAdjustment

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
)

type TypeDeterminationFunction interface {
	CouldTypeBeApplied() (State, error)
	GetType() (ast.Expr, fieldData.Imports)
	GenerateMarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, fieldData.Imports, error)
	GenerateUnmarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, fieldData.Imports, error)
	SetState(states []json.RawMessage, currentPath fieldData.Path, fileData fieldData.FileData, codeGenerator codeGenerators.CodeGenerator) error
	GetState() (json.RawMessage, error)
	GetName() string

	GetSubFiles() (map[fieldData.Path][]*fieldData.File, error)
	NeedsMarshaller() bool

	TypeExpansion() bool
	GetVersion() *string

	Clone() TypeDeterminationFunction
}

type State int

const (
	StateFailed State = iota
	StateUndecided
	StateApplicable
)

type TypeDeterminationFunctions []TypeDeterminationFunction

func (t TypeDeterminationFunctions) GetByName(name string) (TypeDeterminationFunction, error) {
	for _, typeDeterminationFunction := range t {
		if typeDeterminationFunction.GetName() == name {
			return typeDeterminationFunction, nil
		}
	}
	return nil, errors.New("type adjustment function not found")
}

func (t TypeDeterminationFunctions) GetNames() []string {
	var names []string
	for _, checker := range t {
		names = append(names, checker.GetName())
	}
	return names
}

func (t TypeDeterminationFunctions) GenerateHeader() []string {
	if len(t) == 0 {
		return []string{}
	}
	var s []string
	s = append(s, "// Used TypeAdjusters")
	for _, adjuster := range t {
		s = append(s, fmt.Sprintf("// Name: %s Version: %s", adjuster.GetName(), *adjuster.GetVersion()))
	}
	return s
}
