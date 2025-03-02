package buildin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	j2gErrors "github.com/Lemonn/JSON2Go/pkg/typeAdjustment/errors"
	"github.com/araddon/dateparse"
	"go/ast"
)

type TimeTypeChecker struct {
	// IgnoreYearOnlyStrings Set to ignore strings that consist only of a year such as 3294. Most often, they're
	// integers not years!
	ignoreYearOnlyStrings bool
	state                 *timeTypeCheckerState
	utils                 *utils.SeenTypeUtils
	seenTypes             map[string]*fieldData.PathData
}

func (t *TimeTypeChecker) GetModFileContents() []*fieldData.ModFileContent {
	return nil
}

func (t *TimeTypeChecker) GetVersion() string {
	return "v0.0.1"
}

func NewTimeTypeChecker(ignoreYearOnlyStrings bool) *TimeTypeChecker {
	return &TimeTypeChecker{
		ignoreYearOnlyStrings: ignoreYearOnlyStrings,
	}
}

func (t *TimeTypeChecker) CouldTypeBeApplied(path string) (typeAdjustment.State, error) {
	var Level int
	var err error
	//TODO check if its struct type and ignore
	if len(t.seenTypes[path].Types) > 1 {
		return typeAdjustment.StateFailed, nil
	}
	Type := t.utils.GetType(path)
	if len(t.seenTypes[path].Types[Type]) > 1 {
		return typeAdjustment.StateFailed, nil
	}
	for Level = range t.seenTypes[path].Types[Type] {
		break
	}

	layoutStrings := make(map[string]struct{})
	var newLayoutString string
	for value := range t.seenTypes[path].Types[Type][Level] {
		newLayoutString, err = dateparse.ParseFormat(value)
		if err != nil {
			return typeAdjustment.StateFailed, nil
		} else if t.ignoreYearOnlyStrings && t.state.LayoutString == utils.StringToPointer("2006") {
			return typeAdjustment.StateFailed, nil
		} else {
			layoutStrings[newLayoutString] = struct{}{}
		}
	}
	if len(layoutStrings) > 1 {
		return typeAdjustment.StateFailed, nil
	} else if t.state == nil || t.state.LayoutString == nil {
		t.state.LayoutString = &newLayoutString
	} else if _, ok := layoutStrings[*t.state.LayoutString]; !ok {
		//TODO needs rework
		return typeAdjustment.StateApplicable, &j2gErrors.IncompatibleCustomTypeError{
			Timestamp: 0,
			Err: errors.New(fmt.Sprintf("incompatible time strings. Old string: %s, New string: %s",
				*t.state.LayoutString, newLayoutString)),
		}
	}

	return typeAdjustment.StateApplicable, nil
}

func (t *TimeTypeChecker) GenerateMarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, []string, error) {
	functionScaffold.Body = &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "baseValue",
							},
							Sel: &ast.Ident{
								Name: "Format",
							},
						},
						Args: []ast.Expr{
							&ast.Ident{
								Name: "\"" + *t.state.LayoutString + "\"",
							},
						},
					},
					&ast.Ident{
						Name: "nil",
					},
				},
			},
		},
	}
	return functionScaffold, []string{"time"}, nil
}

func (t *TimeTypeChecker) GenerateUnmarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, []string, error) {
	functionScaffold.Body = &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "time",
							},
							Sel: &ast.Ident{
								Name: "Parse",
							},
						},
						Args: []ast.Expr{
							&ast.Ident{
								Name: "\"" + *t.state.LayoutString + "\"",
							},
							&ast.Ident{
								Name: "baseValue",
							},
						},
					},
				},
			},
		},
	}
	return functionScaffold, []string{"time"}, nil
}

func (t *TimeTypeChecker) GetExtraCode() ([]ast.Decl, []string) {
	return []ast.Decl{}, nil
}

func (t *TimeTypeChecker) TypeExpansion() bool {
	//TODO implement me
	// Should return true, when the from or to function or any other code part changed
	return false
}

func (t *TimeTypeChecker) ForceSourceType() *string {
	return nil
}

type timeTypeCheckerState struct {
	LayoutString *string `json:"layoutString,omitempty"`
}

func (t *timeTypeCheckerState) combiner(t1 *timeTypeCheckerState) (*timeTypeCheckerState, error) {
	var tNew timeTypeCheckerState
	if t.LayoutString == nil {
		tNew.LayoutString = t1.LayoutString
	} else if t1.LayoutString == nil {
		tNew.LayoutString = t.LayoutString
	} else if t.LayoutString != tNew.LayoutString {
		return &tNew, &j2gErrors.IncompatibleCustomTypeError{
			Timestamp: 0,
			Err:       nil,
		}
	}
	return &tNew, nil
}

func (t *TimeTypeChecker) SetState(states []json.RawMessage, currentPath string, activeTypeCheckers []typeAdjustment.TypeDeterminationFunction, seenTypes map[string]*fieldData.PathData) error {
	t.seenTypes = seenTypes
	t.utils = utils.NewSeenTypeUtils(seenTypes)
	if states == nil || len(states) == 0 {
		t.state = &timeTypeCheckerState{
			LayoutString: nil,
		}
		return nil
	}
	var state *timeTypeCheckerState
	err := json.Unmarshal(states[0], &state)
	if err != nil {
		return err
	}
	for i := 1; i < len(states); i++ {
		var currentState *timeTypeCheckerState
		err := json.Unmarshal(states[i], &currentState)
		if err != nil {
			return err
		}
		state, err = state.combiner(currentState)
		if err != nil {
			return err
		}
	}
	t.state = state
	return nil
}

func (t *TimeTypeChecker) GetState() (json.RawMessage, error) {
	b, err := json.Marshal(t.state)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (t *TimeTypeChecker) GetType() ast.Expr {
	return &ast.SelectorExpr{
		X: &ast.Ident{
			Name: "time",
		},
		Sel: &ast.Ident{
			Name: "Time",
		},
	}
}

func (t *TimeTypeChecker) GetRequiredImports() []string {
	return []string{"time"}
}

func (t *TimeTypeChecker) SetFile(_ *ast.File) {}

func (t *TimeTypeChecker) GetName() string {
	return "json2go.TimeTypeChecker"
}
