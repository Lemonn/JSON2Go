package buildin

import (
	"encoding/json"
	"errors"
	"fmt"
	j2gErrors "github.com/Lemonn/JSON2Go/pkg/errors"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"github.com/araddon/dateparse"
	"go/ast"
)

type TimeTypeChecker struct {
	// IgnoreYearOnlyStrings Set to ignore strings that consist only of a year such as 3294. Most often, they're
	// integers not years!
	ignoreYearOnlyStrings bool
	state                 *timeTypeCheckerState
}

func NewTimeTypeChecker(ignoreYearOnlyStrings bool) *TimeTypeChecker {
	return &TimeTypeChecker{
		ignoreYearOnlyStrings: ignoreYearOnlyStrings,
	}
}

type timeTypeCheckerState struct {
	LayoutString string `json:"layoutString,omitempty"`
}

func (t *TimeTypeChecker) SetState(state json.RawMessage, currentPath string) error {
	var s timeTypeCheckerState
	if state != nil {
		err := json.Unmarshal(state, &s)
		if err != nil {
			return err
		}
	}
	t.state = &s
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

func (t *TimeTypeChecker) CouldTypeBeApplied(seenValues map[string]*fieldData.ValueData) (typeAdjustment.State, error) {
	var err error
	layoutStrings := make(map[string]struct{})
	var newLayoutString string
	for value := range seenValues {
		newLayoutString, err = dateparse.ParseFormat(value)
		if err != nil {
			return typeAdjustment.StateFailed, nil
		} else if t.ignoreYearOnlyStrings && t.state.LayoutString == "2006" {
			return typeAdjustment.StateFailed, nil
		} else {
			layoutStrings[newLayoutString] = struct{}{}
		}
	}
	if len(layoutStrings) > 1 {
		return typeAdjustment.StateFailed, nil
	} else if t.state == nil || t.state.LayoutString == "" {
		t.state.LayoutString = newLayoutString
	} else if _, ok := layoutStrings[t.state.LayoutString]; !ok {
		return typeAdjustment.StateApplicable, &j2gErrors.IncompatibleCustomTypeError{
			Timestamp: 0,
			Err: errors.New(fmt.Sprintf("incompatible time strings. Old string: %s, New string: %s",
				t.state.LayoutString, newLayoutString)),
		}
	}
	return typeAdjustment.StateApplicable, nil
}

func (t *TimeTypeChecker) GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error) {
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
								Name: "\"" + t.state.LayoutString + "\"",
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
	return functionScaffold, nil
}

func (t *TimeTypeChecker) GenerateToTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error) {
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
								Name: "\"" + t.state.LayoutString + "\"",
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
	return functionScaffold, nil
}

func (t *TimeTypeChecker) GetRequiredImports() []string {
	return []string{"time"}
}

func (t *TimeTypeChecker) SetFile(_ *ast.File) {}

func (t *TimeTypeChecker) GetName() string {
	return "json2go.TimeTypeChecker"
}
