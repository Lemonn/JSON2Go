package buildin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	j2gErrors "github.com/Lemonn/JSON2Go/pkg/errors/typeChecker"
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
	fileData              fieldData.FileData
	codeGenerator         codeGenerators.CodeGenerator
	version               string
	currentPath           fieldData.Path
}

func (t *TimeTypeChecker) GetVersion() *string {
	return &t.version
}

func (t *TimeTypeChecker) NeedsMarshaller() bool {
	return true
}

func NewTimeTypeChecker(ignoreYearOnlyStrings bool) *TimeTypeChecker {
	return &TimeTypeChecker{
		ignoreYearOnlyStrings: ignoreYearOnlyStrings,
	}
}

func (t *TimeTypeChecker) CouldTypeBeApplied() (typeAdjustment.State, error) {
	var err error
	pathData := t.fileData[t.currentPath]
	basicType, levelOfArray, Type := t.codeGenerator.IsBasicTypeWhitDetails(t.currentPath)
	if !basicType {
		return typeAdjustment.StateFailed, nil
	}

	layoutStrings := make(map[string]struct{})
	var newLayoutString string
	for value := range pathData.Types[Type][levelOfArray] {
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

func (t *TimeTypeChecker) GenerateMarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, fieldData.Imports, error) {
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
	return functionScaffold, fieldData.Imports{&fieldData.Import{
		Path:            "time",
		Alias:           nil,
		NeedsAdjustment: false,
		IsGlobal:        false,
	}}, nil
}

func (t *TimeTypeChecker) GenerateUnmarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, fieldData.Imports, error) {
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
	return functionScaffold, fieldData.Imports{&fieldData.Import{
		Path:            "time",
		Alias:           nil,
		NeedsAdjustment: false,
		IsGlobal:        false,
	}}, nil
}

func (t *TimeTypeChecker) GetSubFiles() (map[fieldData.Path][]*fieldData.File, error) {
	return nil, nil
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

func (t *TimeTypeChecker) SetState(states []json.RawMessage, currentPath fieldData.Path, fileData fieldData.FileData, codeGenerator codeGenerators.CodeGenerator) error {
	t.fileData = fileData
	t.codeGenerator = codeGenerator
	if states == nil || len(states) == 0 || states[0] == nil {
		t.state = &timeTypeCheckerState{
			LayoutString: nil,
		}
	} else {
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
	}
	t.currentPath = currentPath
	return nil
}

func (t *TimeTypeChecker) GetState() (json.RawMessage, error) {
	b, err := json.Marshal(t.state)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (t *TimeTypeChecker) GetType() (ast.Expr, fieldData.Imports) {
	return &ast.SelectorExpr{
		X: &ast.Ident{
			Name: "time",
		},
		Sel: &ast.Ident{
			Name: "Time",
		},
	}, fieldData.Imports{{Path: "time"}}
}

func (t *TimeTypeChecker) GetName() string {
	return "json2go.TimeTypeChecker"
}

func (t *TimeTypeChecker) Clone() typeAdjustment.TypeDeterminationFunction {
	return nil
}
