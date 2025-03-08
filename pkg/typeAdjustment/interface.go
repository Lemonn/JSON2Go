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
	GetType() (ast.Expr, *fieldData.Import)
	GenerateMarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, []string, error)
	GenerateUnmarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, []string, error)
	GetName() string
	SetState(states []json.RawMessage, currentPath string, fileData fieldData.FileData, codeGenerator codeGenerators.CodeGenerator) error
	GetState() (json.RawMessage, error)
	GetExtraCode() ([]ast.Decl, []string, error)

	GetSubFiles() (map[string]*fieldData.File, error)
	NeedsMarshaller() bool

	TypeExpansion() bool

	ForceSourceType() (*string, bool)
	GetModFileContents() []*fieldData.ModFileContent
	GetVersion() *string
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
