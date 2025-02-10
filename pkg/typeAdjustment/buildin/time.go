package buildin

import (
	"encoding/json"
	"github.com/araddon/dateparse"
	"go/ast"
)

type TimeTypeChecker struct {
	// IgnoreYearOnlyStrings Set to ignore strings that consist only of a year such as 3294. Most often, they're
	// integers not years!
	IgnoreYearOnlyStrings bool
	layoutString          string
}

func (t *TimeTypeChecker) SetState(state *json.RawMessage) error {
	//TODO implement me
	panic("implement me")
}

func (t *TimeTypeChecker) GetState() (*json.RawMessage, error) {
	//TODO implement me
	panic("implement me")
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

func (t *TimeTypeChecker) CouldTypeBeApplied(seenValues map[string]string) bool {
	var err error
	for value := range seenValues {
		t.layoutString, err = dateparse.ParseFormat(value)
		if t.IgnoreYearOnlyStrings && t.layoutString == "2006" {
			return false
		}
		if err != nil {
			return false
		}
	}
	return true
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
								Name: "\"" + t.layoutString + "\"",
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
								Name: "\"" + t.layoutString + "\"",
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
