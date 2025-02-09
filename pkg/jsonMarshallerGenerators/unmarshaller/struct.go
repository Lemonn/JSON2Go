package unmarshaller

import (
	"errors"
	"fmt"
	"github.com/Lemonn/AstUtils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/structGenerator"
	"go/ast"
	"go/token"
	"unicode"
)

func (g *Generator) structGenerator(str *ast.StructType, path string, name string) ([]ast.Stmt, []string, error) {
	var stmts []ast.Stmt
	var required bool

	stmts = append(stmts, &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: "data",
						},
					},
					Type: &ast.MapType{
						Key: &ast.Ident{
							Name: "string",
						},
						Value: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "json",
							},
							Sel: &ast.Ident{
								Name: "RawMessage",
							},
						},
					},
				},
			},
		},
	})
	stmts = append(stmts, &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: "joinedErrors",
						},
					},
					Type: &ast.Ident{
						Name: "error",
					},
				},
			},
		},
	})
	stmts = append(stmts, &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "err",
			},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "json",
					},
					Sel: &ast.Ident{
						Name: "Unmarshal",
					},
				},
				Args: []ast.Expr{
					&ast.Ident{
						Name: "bytes",
					},
					&ast.UnaryExpr{
						Op: token.AND,
						X: &ast.Ident{
							Name: "data",
						},
					},
				},
			},
		},
	})
	stmts = append(stmts, &ast.IfStmt{
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
						&ast.Ident{
							Name: "err",
						},
					},
				},
			},
		},
	})

	for _, field := range str.Fields.List {
		var fData *fieldData.Data
		var jsonName string
		if v, ok := structGenerator.Tags[path+"."+field.Names[0].Name]; !ok {
			return nil, nil, errors.New(fmt.Sprintf("struct field not found, path: %s", path+"."+field.Names[0].Name))
		} else if v.JsonFieldName == nil {
			continue
		} else {
			fData = v
			jsonName = *v.JsonFieldName
		}

		if fData.ParseFunctions != nil {
			required = true
			stmts = append(stmts, &ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: "value",
						},
						&ast.Ident{
							Name: "ok",
						},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.IndexExpr{
							X: &ast.Ident{
								Name: "data",
							},
							Index: &ast.BasicLit{
								Kind:  token.STRING,
								Value: "\"" + jsonName + "\"",
							},
						},
					},
				},
				Cond: &ast.Ident{
					Name: "ok",
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.DeclStmt{
							Decl: &ast.GenDecl{
								Tok: token.VAR,
								Specs: []ast.Spec{
									&ast.ValueSpec{
										Names: []*ast.Ident{
											&ast.Ident{
												Name: "unmarshalledValue",
											},
										},
										Type: &ast.Ident{
											Name: *fData.BaseType,
										},
									},
								},
							},
						},
						&ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{
									Name: "err",
								},
							},
							Tok: token.ASSIGN,
							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "json",
										},
										Sel: &ast.Ident{
											Name: "Unmarshal",
										},
									},
									Args: []ast.Expr{
										&ast.Ident{
											Name: "value",
										},
										&ast.UnaryExpr{
											Op: token.AND,
											X: &ast.Ident{
												Name: "unmarshalledValue",
											},
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
											&ast.Ident{
												Name: "err",
											},
										},
									},
								},
							},
						},
						&ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.SelectorExpr{
									X: &ast.Ident{
										Name: string(unicode.ToLower([]rune(name)[0])),
									},
									Sel: &ast.Ident{
										Name: field.Names[0].Name,
									},
								},
								&ast.Ident{
									Name: "err",
								},
							},
							Tok: token.ASSIGN,
							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.Ident{
										Name: fData.ParseFunctions.FromTypeParseFunction,
									},
									Args: []ast.Expr{
										&ast.Ident{
											Name: "unmarshalledValue",
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
											&ast.CallExpr{
												Fun: &ast.SelectorExpr{
													X: &ast.Ident{
														Name: "errors",
													},
													Sel: &ast.Ident{
														Name: "Join",
													},
												},
												Args: []ast.Expr{
													&ast.Ident{
														Name: "joinedErrors",
													},
													&ast.Ident{
														Name: "err",
													},
												},
											},
										},
									},
								},
							},
						},
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: &ast.Ident{
									Name: "delete",
								},
								Args: []ast.Expr{
									&ast.Ident{
										Name: "data",
									},
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: "\"" + jsonName + "\"",
									},
								},
							},
						},
					},
				},
			})
		} else if !AstUtils.IsBasicField(field) {
			stmts = append(stmts, &ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: "value",
						},
						&ast.Ident{
							Name: "ok",
						},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.IndexExpr{
							X: &ast.Ident{
								Name: "data",
							},
							Index: &ast.BasicLit{
								Kind:  token.STRING,
								Value: "\"" + jsonName + "\"",
							},
						},
					},
				},
				Cond: &ast.Ident{
					Name: "ok",
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{
									Name: "err",
								},
							},
							Tok: token.ASSIGN,
							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "json",
										},
										Sel: &ast.Ident{
											Name: "Unmarshal",
										},
									},
									Args: []ast.Expr{
										&ast.Ident{
											Name: "value",
										},
										&ast.UnaryExpr{
											Op: token.AND,
											X: &ast.SelectorExpr{
												X: &ast.Ident{
													Name: string(unicode.ToLower([]rune(name)[0])),
												},
												Sel: &ast.Ident{
													Name: field.Names[0].Name,
												},
											},
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
									&ast.DeclStmt{
										Decl: &ast.GenDecl{
											Tok: token.VAR,
											Specs: []ast.Spec{
												&ast.ValueSpec{
													Names: []*ast.Ident{
														&ast.Ident{
															Name: "additionalElementsError",
														},
													},
													Type: &ast.StarExpr{
														X: &ast.Ident{
															Name: "AdditionalElementsError",
														},
													},
												},
											},
										},
									},
									&ast.IfStmt{
										Cond: &ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X: &ast.Ident{
													Name: "errors",
												},
												Sel: &ast.Ident{
													Name: "As",
												},
											},
											Args: []ast.Expr{
												&ast.Ident{
													Name: "err",
												},
												&ast.UnaryExpr{
													Op: token.AND,
													X: &ast.Ident{
														Name: "additionalElementsError",
													},
												},
											},
										},
										Body: &ast.BlockStmt{
											List: []ast.Stmt{
												&ast.AssignStmt{
													Lhs: []ast.Expr{
														&ast.Ident{
															Name: "joinedErrors",
														},
													},
													Tok: token.ASSIGN,
													Rhs: []ast.Expr{
														&ast.CallExpr{
															Fun: &ast.SelectorExpr{
																X: &ast.Ident{
																	Name: "errors",
																},
																Sel: &ast.Ident{
																	Name: "Join",
																},
															},
															Args: []ast.Expr{
																&ast.Ident{
																	Name: "additionalElementsError",
																},
															},
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
															Name: "err",
														},
													},
												},
											},
										},
									},
								},
							},
						},
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: &ast.Ident{
									Name: "delete",
								},
								Args: []ast.Expr{
									&ast.Ident{
										Name: "data",
									},
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: "\"" + jsonName + "\"",
									},
								},
							},
						},
					},
				},
			})
		} else {
			stmts = append(stmts, &ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: "value",
						},
						&ast.Ident{
							Name: "ok",
						},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.IndexExpr{
							X: &ast.Ident{
								Name: "data",
							},
							Index: &ast.BasicLit{
								Kind:  token.STRING,
								Value: "\"" + jsonName + "\"",
							},
						},
					},
				},
				Cond: &ast.Ident{
					Name: "ok",
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{
									Name: "err",
								},
							},
							Tok: token.ASSIGN,
							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "json",
										},
										Sel: &ast.Ident{
											Name: "Unmarshal",
										},
									},
									Args: []ast.Expr{
										&ast.Ident{
											Name: "value",
										},
										&ast.UnaryExpr{
											Op: token.AND,
											X: &ast.SelectorExpr{
												X: &ast.Ident{
													Name: string(unicode.ToLower([]rune(name)[0])),
												},
												Sel: &ast.Ident{
													Name: field.Names[0].Name,
												},
											},
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
									&ast.DeclStmt{
										Decl: &ast.GenDecl{
											Tok: token.VAR,
											Specs: []ast.Spec{
												&ast.ValueSpec{
													Names: []*ast.Ident{
														&ast.Ident{
															Name: "additionalElementsError",
														},
													},
													Type: &ast.StarExpr{
														X: &ast.Ident{
															Name: "AdditionalElementsError",
														},
													},
												},
											},
										},
									},
									&ast.IfStmt{
										Cond: &ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X: &ast.Ident{
													Name: "errors",
												},
												Sel: &ast.Ident{
													Name: "As",
												},
											},
											Args: []ast.Expr{
												&ast.Ident{
													Name: "err",
												},
												&ast.UnaryExpr{
													Op: token.AND,
													X: &ast.Ident{
														Name: "additionalElementsError",
													},
												},
											},
										},
										Body: &ast.BlockStmt{
											List: []ast.Stmt{
												&ast.AssignStmt{
													Lhs: []ast.Expr{
														&ast.Ident{
															Name: "joinedErrors",
														},
													},
													Tok: token.ASSIGN,
													Rhs: []ast.Expr{
														&ast.CallExpr{
															Fun: &ast.SelectorExpr{
																X: &ast.Ident{
																	Name: "errors",
																},
																Sel: &ast.Ident{
																	Name: "Join",
																},
															},
															Args: []ast.Expr{
																&ast.Ident{
																	Name: "joinedErrors",
																},
																&ast.Ident{
																	Name: "additionalElementsError",
																},
															},
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
															Name: "err",
														},
													},
												},
											},
										},
									},
								},
							},
						},
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: &ast.Ident{
									Name: "delete",
								},
								Args: []ast.Expr{
									&ast.Ident{
										Name: "data",
									},
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: "\"" + jsonName + "\"",
									},
								},
							},
						},
					},
				},
			})
		}
	}

	if len(stmts) == 0 || !required {
		return []ast.Stmt{}, nil, nil
	}

	// Add If statement to custom unmarshall function that checks for additional elements,
	// and joins an error to the error list if it's the case
	stmts = append(stmts, &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X: &ast.CallExpr{
				Fun: &ast.Ident{
					Name: "len",
				},
				Args: []ast.Expr{
					&ast.Ident{
						Name: "data",
					},
				},
			},
			Op: token.NEQ,
			Y: &ast.BasicLit{
				Kind:  token.INT,
				Value: "0",
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: "joinedErrors",
						},
					},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "errors",
								},
								Sel: &ast.Ident{
									Name: "Join",
								},
							},
							Args: []ast.Expr{
								&ast.Ident{
									Name: "joinedErrors",
								},
								&ast.UnaryExpr{
									Op: token.AND,
									X: &ast.CompositeLit{
										Type: &ast.Ident{
											Name: "AdditionalElementsError",
										},
										Elts: []ast.Expr{
											&ast.KeyValueExpr{
												Key: &ast.Ident{
													Name: "ParsedObj",
												},
												Value: &ast.BasicLit{
													Kind:  token.STRING,
													Value: "\"" + name + "\"",
												},
											},
											&ast.KeyValueExpr{
												Key: &ast.Ident{
													Name: "Elements",
												},
												Value: &ast.Ident{
													Name: "data",
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
	})
	// Add return statement to custom unmarshall function
	stmts = append(stmts, &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.Ident{
				Name: "joinedErrors",
			},
		},
	})
	return stmts, []string{"encoding/json", "errors", "fmt"}, nil
}

func addGetAllErrorsOfTypeFunction(file *ast.File) {
	file.Decls = append(file.Decls, &ast.FuncDecl{
		Name: &ast.Ident{
			Name: "GetAllErrorsOfType",
		},
		Type: &ast.FuncType{
			TypeParams: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "T",
							},
						},
						Type: &ast.Ident{
							Name: "error",
						},
					},
				},
			},
			Params: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "errType",
							},
						},
						Type: &ast.Ident{
							Name: "T",
						},
					},
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "e",
							},
						},
						Type: &ast.Ident{
							Name: "error",
						},
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: &ast.ArrayType{
							Elt: &ast.Ident{
								Name: "T",
							},
						},
					},
					&ast.Field{
						Type: &ast.Ident{
							Name: "error",
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.DeclStmt{
					Decl: &ast.GenDecl{
						Tok: token.VAR,
						Specs: []ast.Spec{
							&ast.ValueSpec{
								Names: []*ast.Ident{
									&ast.Ident{
										Name: "result",
									},
								},
								Type: &ast.ArrayType{
									Elt: &ast.Ident{
										Name: "T",
									},
								},
							},
						},
					},
				},
				&ast.LabeledStmt{
					Label: &ast.Ident{
						Name: "UNWRAP",
					},
					Stmt: &ast.TypeSwitchStmt{
						Assign: &ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{
									Name: "err",
								},
							},
							Tok: token.DEFINE,
							Rhs: []ast.Expr{
								&ast.TypeAssertExpr{
									X: &ast.Ident{
										Name: "e",
									},
								},
							},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.CaseClause{
									List: []ast.Expr{
										&ast.InterfaceType{
											Methods: &ast.FieldList{
												List: []*ast.Field{
													&ast.Field{
														Names: []*ast.Ident{
															&ast.Ident{
																Name: "Unwrap",
															},
														},
														Type: &ast.FuncType{
															Params: &ast.FieldList{},
															Results: &ast.FieldList{
																List: []*ast.Field{
																	&ast.Field{
																		Type: &ast.ArrayType{
																			Elt: &ast.Ident{
																				Name: "error",
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
									Body: []ast.Stmt{
										&ast.IfStmt{
											Cond: &ast.CallExpr{
												Fun: &ast.SelectorExpr{
													X: &ast.Ident{
														Name: "errors",
													},
													Sel: &ast.Ident{
														Name: "As",
													},
												},
												Args: []ast.Expr{
													&ast.IndexExpr{
														X: &ast.CallExpr{
															Fun: &ast.SelectorExpr{
																X: &ast.Ident{
																	Name: "err",
																},
																Sel: &ast.Ident{
																	Name: "Unwrap",
																},
															},
														},
														Index: &ast.BinaryExpr{
															X: &ast.CallExpr{
																Fun: &ast.Ident{
																	Name: "len",
																},
																Args: []ast.Expr{
																	&ast.CallExpr{
																		Fun: &ast.SelectorExpr{
																			X: &ast.Ident{
																				Name: "err",
																			},
																			Sel: &ast.Ident{
																				Name: "Unwrap",
																			},
																		},
																	},
																},
															},
															Op: token.SUB,
															Y: &ast.BasicLit{
																Kind:  token.INT,
																Value: "1",
															},
														},
													},
													&ast.UnaryExpr{
														Op: token.AND,
														X: &ast.Ident{
															Name: "errType",
														},
													},
												},
											},
											Body: &ast.BlockStmt{
												List: []ast.Stmt{
													&ast.AssignStmt{
														Lhs: []ast.Expr{
															&ast.Ident{
																Name: "result",
															},
														},
														Tok: token.ASSIGN,
														Rhs: []ast.Expr{
															&ast.CallExpr{
																Fun: &ast.Ident{
																	Name: "append",
																},
																Args: []ast.Expr{
																	&ast.CompositeLit{
																		Type: &ast.ArrayType{
																			Elt: &ast.Ident{
																				Name: "T",
																			},
																		},
																		Elts: []ast.Expr{
																			&ast.Ident{
																				Name: "errType",
																			},
																		},
																	},
																	&ast.Ident{
																		Name: "result",
																	},
																},
																Ellipsis: 270,
															},
														},
													},
												},
											},
										},
										&ast.IfStmt{
											Cond: &ast.BinaryExpr{
												X: &ast.CallExpr{
													Fun: &ast.Ident{
														Name: "len",
													},
													Args: []ast.Expr{
														&ast.CallExpr{
															Fun: &ast.SelectorExpr{
																X: &ast.Ident{
																	Name: "err",
																},
																Sel: &ast.Ident{
																	Name: "Unwrap",
																},
															},
														},
													},
												},
												Op: token.GTR,
												Y: &ast.BasicLit{
													Kind:  token.INT,
													Value: "0",
												},
											},
											Body: &ast.BlockStmt{
												List: []ast.Stmt{
													&ast.AssignStmt{
														Lhs: []ast.Expr{
															&ast.Ident{
																Name: "e",
															},
														},
														Tok: token.ASSIGN,
														Rhs: []ast.Expr{
															&ast.IndexExpr{
																X: &ast.CallExpr{
																	Fun: &ast.SelectorExpr{
																		X: &ast.Ident{
																			Name: "err",
																		},
																		Sel: &ast.Ident{
																			Name: "Unwrap",
																		},
																	},
																},
																Index: &ast.BasicLit{
																	Kind:  token.INT,
																	Value: "0",
																},
															},
														},
													},
													&ast.BranchStmt{
														Tok: token.GOTO,
														Label: &ast.Ident{
															Name: "UNWRAP",
														},
													},
												},
											},
											Else: &ast.BlockStmt{
												List: []ast.Stmt{
													&ast.ReturnStmt{
														Results: []ast.Expr{
															&ast.Ident{
																Name: "result",
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
								&ast.CaseClause{
									Body: []ast.Stmt{
										&ast.IfStmt{
											Cond: &ast.BinaryExpr{
												X: &ast.CallExpr{
													Fun: &ast.Ident{
														Name: "len",
													},
													Args: []ast.Expr{
														&ast.Ident{
															Name: "result",
														},
													},
												},
												Op: token.GTR,
												Y: &ast.BasicLit{
													Kind:  token.INT,
													Value: "0",
												},
											},
											Body: &ast.BlockStmt{
												List: []ast.Stmt{
													&ast.ReturnStmt{
														Results: []ast.Expr{
															&ast.Ident{
																Name: "result",
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
													X: &ast.BinaryExpr{
														X: &ast.CallExpr{
															Fun: &ast.Ident{
																Name: "len",
															},
															Args: []ast.Expr{
																&ast.Ident{
																	Name: "result",
																},
															},
														},
														Op: token.EQL,
														Y: &ast.BasicLit{
															Kind:  token.INT,
															Value: "0",
														},
													},
													Op: token.LAND,
													Y: &ast.CallExpr{
														Fun: &ast.SelectorExpr{
															X: &ast.Ident{
																Name: "errors",
															},
															Sel: &ast.Ident{
																Name: "As",
															},
														},
														Args: []ast.Expr{
															&ast.Ident{
																Name: "err",
															},
															&ast.UnaryExpr{
																Op: token.AND,
																X: &ast.Ident{
																	Name: "errType",
																},
															},
														},
													},
												},
												Body: &ast.BlockStmt{
													List: []ast.Stmt{
														&ast.AssignStmt{
															Lhs: []ast.Expr{
																&ast.Ident{
																	Name: "result",
																},
															},
															Tok: token.ASSIGN,
															Rhs: []ast.Expr{
																&ast.CallExpr{
																	Fun: &ast.Ident{
																		Name: "append",
																	},
																	Args: []ast.Expr{
																		&ast.Ident{
																			Name: "result",
																		},
																		&ast.Ident{
																			Name: "errType",
																		},
																	},
																},
															},
														},
														&ast.ReturnStmt{
															Results: []ast.Expr{
																&ast.Ident{
																	Name: "result",
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
																		&ast.BasicLit{
																			Kind:  token.STRING,
																			Value: "\"error is not of joinError type and also error is not of searched type\"",
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
							},
						},
					},
				},
			},
		},
	})
}

func addCheckForFirstErrorNotOfTypeTFunction(file *ast.File) {
	file.Decls = append(file.Decls, &ast.FuncDecl{
		Name: &ast.Ident{
			Name: "CheckForFirstErrorNotOfTypeT",
		},
		Type: &ast.FuncType{
			TypeParams: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "T",
							},
						},
						Type: &ast.Ident{
							Name: "error",
						},
					},
				},
			},
			Params: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "errType",
							},
						},
						Type: &ast.Ident{
							Name: "T",
						},
					},
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "e",
							},
						},
						Type: &ast.Ident{
							Name: "error",
						},
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: &ast.Ident{
							Name: "error",
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.LabeledStmt{
					Label: &ast.Ident{
						Name: "UNWRAP",
					},
					Stmt: &ast.TypeSwitchStmt{
						Assign: &ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{
									Name: "err",
								},
							},
							Tok: token.DEFINE,
							Rhs: []ast.Expr{
								&ast.TypeAssertExpr{
									X: &ast.Ident{
										Name: "e",
									},
								},
							},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.CaseClause{
									List: []ast.Expr{
										&ast.InterfaceType{
											Methods: &ast.FieldList{
												List: []*ast.Field{
													&ast.Field{
														Names: []*ast.Ident{
															&ast.Ident{
																Name: "Unwrap",
															},
														},
														Type: &ast.FuncType{
															Params: &ast.FieldList{},
															Results: &ast.FieldList{
																List: []*ast.Field{
																	&ast.Field{
																		Type: &ast.ArrayType{
																			Elt: &ast.Ident{
																				Name: "error",
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
									Body: []ast.Stmt{
										&ast.IfStmt{
											Cond: &ast.UnaryExpr{
												Op: token.NOT,
												X: &ast.CallExpr{
													Fun: &ast.SelectorExpr{
														X: &ast.Ident{
															Name: "errors",
														},
														Sel: &ast.Ident{
															Name: "As",
														},
													},
													Args: []ast.Expr{
														&ast.IndexExpr{
															X: &ast.CallExpr{
																Fun: &ast.SelectorExpr{
																	X: &ast.Ident{
																		Name: "err",
																	},
																	Sel: &ast.Ident{
																		Name: "Unwrap",
																	},
																},
															},
															Index: &ast.BinaryExpr{
																X: &ast.CallExpr{
																	Fun: &ast.Ident{
																		Name: "len",
																	},
																	Args: []ast.Expr{
																		&ast.CallExpr{
																			Fun: &ast.SelectorExpr{
																				X: &ast.Ident{
																					Name: "err",
																				},
																				Sel: &ast.Ident{
																					Name: "Unwrap",
																				},
																			},
																		},
																	},
																},
																Op: token.SUB,
																Y: &ast.BasicLit{
																	Kind:  token.INT,
																	Value: "1",
																},
															},
														},
														&ast.UnaryExpr{
															Op: token.AND,
															X: &ast.Ident{
																Name: "errType",
															},
														},
													},
												},
											},
											Body: &ast.BlockStmt{
												List: []ast.Stmt{
													&ast.ReturnStmt{
														Results: []ast.Expr{
															&ast.IndexExpr{
																X: &ast.CallExpr{
																	Fun: &ast.SelectorExpr{
																		X: &ast.Ident{
																			Name: "err",
																		},
																		Sel: &ast.Ident{
																			Name: "Unwrap",
																		},
																	},
																},
																Index: &ast.BinaryExpr{
																	X: &ast.CallExpr{
																		Fun: &ast.Ident{
																			Name: "len",
																		},
																		Args: []ast.Expr{
																			&ast.CallExpr{
																				Fun: &ast.SelectorExpr{
																					X: &ast.Ident{
																						Name: "err",
																					},
																					Sel: &ast.Ident{
																						Name: "Unwrap",
																					},
																				},
																			},
																		},
																	},
																	Op: token.SUB,
																	Y: &ast.BasicLit{
																		Kind:  token.INT,
																		Value: "1",
																	},
																},
															},
														},
													},
												},
											},
										},
										&ast.IfStmt{
											Cond: &ast.BinaryExpr{
												X: &ast.CallExpr{
													Fun: &ast.Ident{
														Name: "len",
													},
													Args: []ast.Expr{
														&ast.CallExpr{
															Fun: &ast.SelectorExpr{
																X: &ast.Ident{
																	Name: "err",
																},
																Sel: &ast.Ident{
																	Name: "Unwrap",
																},
															},
														},
													},
												},
												Op: token.GTR,
												Y: &ast.BasicLit{
													Kind:  token.INT,
													Value: "0",
												},
											},
											Body: &ast.BlockStmt{
												List: []ast.Stmt{
													&ast.AssignStmt{
														Lhs: []ast.Expr{
															&ast.Ident{
																Name: "e",
															},
														},
														Tok: token.ASSIGN,
														Rhs: []ast.Expr{
															&ast.IndexExpr{
																X: &ast.CallExpr{
																	Fun: &ast.SelectorExpr{
																		X: &ast.Ident{
																			Name: "err",
																		},
																		Sel: &ast.Ident{
																			Name: "Unwrap",
																		},
																	},
																},
																Index: &ast.BasicLit{
																	Kind:  token.INT,
																	Value: "0",
																},
															},
														},
													},
													&ast.BranchStmt{
														Tok: token.GOTO,
														Label: &ast.Ident{
															Name: "UNWRAP",
														},
													},
												},
											},
											Else: &ast.BlockStmt{
												List: []ast.Stmt{
													&ast.ReturnStmt{
														Results: []ast.Expr{
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
								&ast.CaseClause{
									Body: []ast.Stmt{
										&ast.IfStmt{
											Cond: &ast.UnaryExpr{
												Op: token.NOT,
												X: &ast.CallExpr{
													Fun: &ast.SelectorExpr{
														X: &ast.Ident{
															Name: "errors",
														},
														Sel: &ast.Ident{
															Name: "As",
														},
													},
													Args: []ast.Expr{
														&ast.Ident{
															Name: "err",
														},
														&ast.UnaryExpr{
															Op: token.AND,
															X: &ast.Ident{
																Name: "errType",
															},
														},
													},
												},
											},
											Body: &ast.BlockStmt{
												List: []ast.Stmt{
													&ast.ReturnStmt{
														Results: []ast.Expr{
															&ast.Ident{
																Name: "err",
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
				},
			},
		},
	})
}
