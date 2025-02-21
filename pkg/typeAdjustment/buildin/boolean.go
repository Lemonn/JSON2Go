package buildin

import (
	"encoding/json"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"go/ast"
	"go/token"
)

type Boolean struct {
	trueValues  map[string]struct{}
	falseValues map[string]struct{}
}

func NewBoolean() *Boolean {
	return &Boolean{
		trueValues:  map[string]struct{}{"true": {}},
		falseValues: map[string]struct{}{"false": {}},
	}
}

func (b *Boolean) CouldTypeBeApplied(seenValues map[string]*fieldData.ValueData) (typeAdjustment.State, error) {
	for s, _ := range seenValues {
		if s != "true" && s != "false" {
			return typeAdjustment.StateFailed, nil
		}
	}
	return typeAdjustment.StateApplicable, nil
}

func (b *Boolean) GetType() ast.Expr {
	return &ast.Ident{Name: "bool"}
}

func (b *Boolean) GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error) {
	functionScaffold.Body.List = append(functionScaffold.Body.List, &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X: &ast.BinaryExpr{
				X: &ast.Ident{
					Name: "baseValue",
				},
				Op: token.NEQ,
				Y: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"true\"",
				},
			},
			Op: token.LAND,
			Y: &ast.BinaryExpr{
				X: &ast.Ident{
					Name: "baseValue",
				},
				Op: token.NEQ,
				Y: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"false\"",
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.Ident{
							Name: "nil",
						},
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "errors",
								},
								Sel: &ast.Ident{
									Name: "New",
								},
							},
							Args: []ast.Expr{
								&ast.BinaryExpr{
									X: &ast.BasicLit{
										Kind:  token.STRING,
										Value: "\"only true or false can be represented as bool, given was: \"",
									},
									Op: token.ADD,
									Y: &ast.Ident{
										Name: "baseValue",
									},
								},
							},
						},
					},
				},
			},
		},
		Else: &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X: &ast.Ident{
					Name: "baseValue",
				},
				Op: token.EQL,
				Y: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"true\"",
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.Ident{
								Name: "true",
							},
							&ast.Ident{
								Name: "nil",
							},
						},
					},
				},
			},
			Else: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.Ident{
								Name: "false",
							},
							&ast.Ident{
								Name: "nil",
							},
						},
					},
				},
			},
		},
	})
	return functionScaffold, nil
}

func (b *Boolean) GenerateToTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error) {
	functionScaffold.Body.List = append(functionScaffold.Body.List, &ast.IfStmt{
		Cond: &ast.Ident{
			Name: "baseValue",
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: "\"true\"",
						},
						&ast.Ident{
							Name: "nil",
						},
					},
				},
			},
		},
		Else: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: "\"false\"",
						},
						&ast.Ident{
							Name: "nil",
						},
					},
				},
			},
		},
	})
	return functionScaffold, nil
}

func (b *Boolean) GetRequiredImports() []string {
	return nil
}

func (b *Boolean) SetFile(_ *ast.File) {

}

func (b *Boolean) GetName() string {
	return "json2Go.Boolean"
}

func (b *Boolean) SetState(_ []*json.RawMessage, _ string) error {
	return nil
}

func (b *Boolean) GetState() ([]*json.RawMessage, error) {
	return nil, nil
}
