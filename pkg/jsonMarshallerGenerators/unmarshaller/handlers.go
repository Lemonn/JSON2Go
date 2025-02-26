package unmarshaller

import (
	"go/ast"
	"go/token"
)

func (g *Generator) getGlobalImports() []string {
	return []string{
		"encoding/json",
		"errors",
		"fmt",
	}
}

func (g *Generator) addCheckForFirstErrorNotOfTypeTFunction() ast.Decl {
	return &ast.FuncDecl{
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
	}
}

func (g *Generator) addGetAllErrorsOfTypeFunction() ast.Decl {
	return &ast.FuncDecl{
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
	}
}

func (g *Generator) addAdditionalElementsError() []ast.Decl {
	var decls []ast.Decl
	decls = append(decls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{
					Name: "AdditionalElementsError",
				},
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: []*ast.Field{
							&ast.Field{
								Names: []*ast.Ident{
									&ast.Ident{
										Name: "ParsedObj",
									},
								},
								Type: &ast.Ident{
									Name: "string",
								},
							},
							&ast.Field{
								Names: []*ast.Ident{
									&ast.Ident{
										Name: "Elements",
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
				},
			},
		},
	})
	decls = append(decls, &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				&ast.Field{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: "j",
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
		Name: &ast.Ident{
			Name: "String",
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: &ast.Ident{
							Name: "string",
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "j",
								},
								Sel: &ast.Ident{
									Name: "Error",
								},
							},
						},
					},
				},
			},
		},
	})
	decls = append(decls, &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				&ast.Field{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: "j",
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
		Name: &ast.Ident{
			Name: "Error",
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: &ast.Ident{
							Name: "string",
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
							Name: "m",
						},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: "\"the following unexpected additional elements were found: \"",
						},
					},
				},
				&ast.RangeStmt{
					Key: &ast.Ident{
						Name: "s",
					},
					Value: &ast.Ident{
						Name: "e",
					},
					Tok: token.DEFINE,
					X: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: "j",
						},
						Sel: &ast.Ident{
							Name: "Elements",
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.Ident{
										Name: "m",
									},
								},
								Tok: token.ADD_ASSIGN,
								Rhs: []ast.Expr{
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
												Value: "\"[(%s) RawJsonString {\\\"%s\\\": %s}]\"",
											},
											&ast.Ident{
												Name: "s",
											},
											&ast.Ident{
												Name: "s",
											},
											&ast.Ident{
												Name: "e",
											},
										},
									},
								},
							},
						},
					},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: "m",
						},
					},
					Tok: token.ADD_ASSIGN,
					Rhs: []ast.Expr{
						&ast.BinaryExpr{
							X: &ast.BasicLit{
								Kind:  token.STRING,
								Value: "\" whilst parsing \"",
							},
							Op: token.ADD,
							Y: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "j",
								},
								Sel: &ast.Ident{
									Name: "ParsedObj",
								},
							},
						},
					},
				},
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.Ident{
							Name: "m",
						},
					},
				},
			},
		},
	})
	return decls
}
