package JSON2Go

import (
	"github.com/araddon/dateparse"
	"go/ast"
)

type TypeDeterminationFunction interface {
	CouldTypeBeApplied(seenValues []string) bool
	GetType() string
	GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl
	GenerateToTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl
}

type TimeTypeChecker struct {
	layoutString string
}

func (t *TimeTypeChecker) GetType() string {
	return "time.Time"
}

func (t *TimeTypeChecker) CouldTypeBeApplied(seenValues []string) bool {
	var err error
	for _, value := range seenValues {
		t.layoutString, err = dateparse.ParseFormat(value)
		if err != nil {
			return false
		}
	}
	return true
}

func (t *TimeTypeChecker) GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl {
	for i, result := range functionScaffold.Type.Results.List {
		if result.Type.(*ast.Ident).Name == "TODO" {
			functionScaffold.Type.Results.List[i].Type.(*ast.Ident).Name = "time.Time"
		}
	}
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
	return functionScaffold
}

func (t *TimeTypeChecker) GenerateToTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl {
	for i, param := range functionScaffold.Type.Params.List {
		if param.Type.(*ast.Ident).Name == "TODO" {
			functionScaffold.Type.Params.List[i].Type.(*ast.Ident).Name = "time.Time"
		}
	}
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
	return functionScaffold
}
