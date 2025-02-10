package buildin

import (
	"encoding/json"
	"errors"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"go/ast"
	"go/token"
	"strconv"
)

type IntTypeChecker struct{}

func (i *IntTypeChecker) SetState(state *json.RawMessage) error {
	//TODO implement me
	panic("implement me")
}

func (i *IntTypeChecker) GetState() (*json.RawMessage, error) {
	//TODO implement me
	panic("implement me")
}

func (i *IntTypeChecker) CouldTypeBeApplied(seenValues map[string]string) bool {
	for value := range seenValues {
		_, err := strconv.Atoi(value)
		if err != nil {
			return false
		}
	}
	return true
}

func (i *IntTypeChecker) GetType() ast.Expr {
	return &ast.Ident{Name: "int"}
}

func (i *IntTypeChecker) GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error) {
	inputType, err := getInputType(functionScaffold)
	if err != nil {
		return nil, err
	}
	if inputType == "float64" {
		functionScaffold.Body = &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.Ident{
								Name: "int",
							},
							Args: []ast.Expr{
								&ast.Ident{
									Name: "baseValue",
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
	} else {
		functionScaffold.Body = &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: "i",
						},
						&ast.Ident{
							Name: "err",
						},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "strconv",
								},
								Sel: &ast.Ident{
									Name: "Atoi",
								},
							},
							Args: []ast.Expr{
								&ast.Ident{
									Name: "baseValue",
								},
							},
						},
					},
				},
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{
						X: &ast.Ident{
							Name: "err",
						},
						Op: token.NEQ,
						Y: &ast.Ident{
							Name: "nil",
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.BasicLit{
										Kind:  token.INT,
										Value: "0",
									},
									&ast.Ident{
										Name: "err",
									},
								},
							},
						},
					},
				},
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.Ident{
							Name: "i",
						},
						&ast.Ident{
							Name: "nil",
						},
					},
				},
			},
		}
	}
	return functionScaffold, nil
}

func (i *IntTypeChecker) GenerateToTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error) {
	returnType, err := getReturnType(functionScaffold)
	if err != nil {
		return nil, err
	}
	if returnType == "float64" {
		functionScaffold.Body = &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.Ident{
								Name: "float64",
							},
							Args: []ast.Expr{
								&ast.Ident{
									Name: "baseValue",
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
	} else {
		functionScaffold.Body = &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: "i",
						},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "strconv",
								},
								Sel: &ast.Ident{
									Name: "Itoa",
								},
							},
							Args: []ast.Expr{
								&ast.Ident{
									Name: "baseValue",
								},
							},
						},
					},
				},
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.Ident{
							Name: "i",
						},
						&ast.Ident{
							Name: "nil",
						},
					},
				},
			},
		}
	}
	return functionScaffold, nil
}

func (i *IntTypeChecker) GetRequiredImports() []string {
	return []string{"strconv"}
}

func (i *IntTypeChecker) SetFile(_ *ast.File) {}

func (i *IntTypeChecker) GetName() string {
	return "json2go.IntTypeChecker"
}

func getInputType(functionScaffold *ast.FuncDecl) (string, error) {
	for _, expr := range functionScaffold.Type.Params.List {
		n, err := utils.WalkExpressions(&expr.Type)
		if err != nil {
			return "", err
		}
		switch e := (*n).(type) {
		case *ast.SelectorExpr:
			return e.Sel.Name + "." + e.X.(*ast.Ident).Name, nil
		case *ast.Ident:
			return e.Name, nil
		case *ast.InterfaceType:
			return "interface{}", nil
		}
	}
	return "", errors.New("no valid input type")
}

func getReturnType(functionScaffold *ast.FuncDecl) (string, error) {
	for _, expr := range functionScaffold.Type.Results.List {
		n, err := utils.WalkExpressions(&expr.Type)
		if err != nil {
			return "", err
		}
		switch e := (*n).(type) {
		case *ast.SelectorExpr:
			return e.Sel.Name + "." + e.X.(*ast.Ident).Name, nil
		case *ast.Ident:
			return e.Name, nil
		case *ast.InterfaceType:
			return "interface{}", nil
		}
	}
	return "", errors.New("no valid result type")
}
