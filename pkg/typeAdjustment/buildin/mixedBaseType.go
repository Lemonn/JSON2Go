package buildin

import (
	"encoding/json"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"go/ast"
	"go/token"
)

type MixedBaseTypeTypeChecker struct {
}

func (m *MixedBaseTypeTypeChecker) CouldTypeBeApplied(seenValues map[string]*fieldData.ValueData) (typeAdjustment.State, error) {
	seenTypes := make(map[string]struct{})
	for _, data := range seenValues {
		if !(data.Type == "string" || data.Type == "float64" || data.Type == "bool") {
			return typeAdjustment.StateFailed, nil
		}
		seenTypes[data.Type] = struct{}{}
	}
	if len(seenTypes) == 1 {
		return typeAdjustment.StateFailed, nil
	}
	return typeAdjustment.StateApplicable, nil
}

func (m *MixedBaseTypeTypeChecker) GetType() ast.Expr {
	return &ast.Ident{Name: "string"}
}

func (m *MixedBaseTypeTypeChecker) GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error) {
	functionScaffold.Body = &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.TypeSwitchStmt{
				Assign: &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: "v",
						},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.TypeAssertExpr{
							X: &ast.Ident{
								Name: "baseValue",
							},
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.CaseClause{
							List: []ast.Expr{
								&ast.Ident{
									Name: "string",
								},
							},
							Body: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.Ident{
											Name: "v",
										},
										&ast.Ident{
											Name: "nil",
										},
									},
								},
							},
						},
						&ast.CaseClause{
							List: []ast.Expr{
								&ast.Ident{
									Name: "float64",
								},
							},
							Body: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X: &ast.Ident{
													Name: "strconv",
												},
												Sel: &ast.Ident{
													Name: "FormatFloat",
												},
											},
											Args: []ast.Expr{
												&ast.Ident{
													Name: "v",
												},
												&ast.BasicLit{
													Kind:  token.CHAR,
													Value: "'f'",
												},
												&ast.UnaryExpr{
													Op: token.SUB,
													X: &ast.BasicLit{
														Kind:  token.INT,
														Value: "1",
													},
												},
												&ast.BasicLit{
													Kind:  token.INT,
													Value: "64",
												},
											},
										},
										&ast.Ident{
											Name: "nil",
										},
									},
								},
							},
						},
						&ast.CaseClause{
							List: []ast.Expr{
								&ast.Ident{
									Name: "bool",
								},
							},
							Body: []ast.Stmt{
								&ast.IfStmt{
									Cond: &ast.Ident{
										Name: "v",
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
								},
							},
						},
					},
				},
			},
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.BasicLit{
						Kind:  token.STRING,
						Value: "\"\"",
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
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "fmt",
									},
									Sel: &ast.Ident{
										Name: "Sprintf",
									},
								},
								Args: []ast.Expr{
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: "\"invalid type: %T\"",
									},
									&ast.Ident{
										Name: "baseValue",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return functionScaffold, nil
}

func (m *MixedBaseTypeTypeChecker) GenerateToTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error) {
	functionScaffold.Body = &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{
						Name: "float",
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
								Name: "ParseFloat",
							},
						},
						Args: []ast.Expr{
							&ast.Ident{
								Name: "baseValue",
							},
							&ast.BasicLit{
								Kind:  token.INT,
								Value: "64",
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
					Op: token.EQL,
					Y: &ast.Ident{
						Name: "nil",
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.Ident{
									Name: "float",
								},
								&ast.Ident{
									Name: "nil",
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
					Else: &ast.IfStmt{
						Cond: &ast.BinaryExpr{
							X: &ast.Ident{
								Name: "baseValue",
							},
							Op: token.EQL,
							Y: &ast.BasicLit{
								Kind:  token.STRING,
								Value: "\"false\"",
							},
						},
						Body: &ast.BlockStmt{
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
						Else: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.Ident{
											Name: "baseValue",
										},
										&ast.Ident{
											Name: "nil",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return functionScaffold, nil
}

func (m *MixedBaseTypeTypeChecker) GetRequiredImports() []string {
	return []string{"strconv"}
}

func (m *MixedBaseTypeTypeChecker) SetFile(_ *ast.File) {}

func (m *MixedBaseTypeTypeChecker) GetName() string {
	return "json2go.MixedBaseTypeTypeChecker"
}

func (m *MixedBaseTypeTypeChecker) SetState(_ json.RawMessage, _ string) error {
	return nil
}

func (m *MixedBaseTypeTypeChecker) GetState() (json.RawMessage, error) {
	return nil, nil
}
