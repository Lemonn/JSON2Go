package buildin

import (
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"go/token"
)

func (g *Generator) unmarshallStructGenerator(path fieldData.Path) ([]ast.Stmt, fieldData.Imports, error) {
	var stmts []ast.Stmt
	var required bool
	stmts = append(stmts, &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						{
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
						{
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

	var Level int
	for Level, _ = range g.fileData[path].Types[fieldData.Field] {
		break
	}
	for fieldPath, _ := range g.fileData[path].Types[fieldData.Field][Level] {
		levelOfArrays := g.codeGenerator.GetLevelOfArrays(fieldPath)

		if g.fileData[fieldPath].TypeAdjusterData != nil && g.fileData[fieldPath].ActiveType != nil {
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
								Value: "\"" + g.fileData[fieldPath].JsonFieldName + "\"",
							},
						},
					},
				},
				Cond: &ast.Ident{
					Name: "ok",
				},
				Body: &ast.BlockStmt{
					List: func() []ast.Stmt {
						if levelOfArrays == 0 {
							return g.unmarshallHandleField(fieldPath)
						} else {
							return g.unmarshallHandleArrayField(fieldPath)
						}
					}(),
				},
			})
		} else if !g.codeGenerator.IsStruct(fieldPath) {
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
								Value: "\"" + g.fileData[fieldPath].JsonFieldName + "\"",
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
													Name: path.GetRune(),
												},
												Sel: &ast.Ident{
													Name: fieldPath.GetFieldName(),
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
										Value: "\"" + g.fileData[fieldPath].JsonFieldName + "\"",
									},
								},
							},
						},
					},
				},
				Else: func() ast.Stmt {
					if g.fileData[fieldPath].RequiredField {
						return &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ExprStmt{
									X: &ast.CallExpr{
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
												Name: "err",
											},
											&ast.UnaryExpr{
												Op: token.AND,
												X: &ast.CompositeLit{
													Type: &ast.SelectorExpr{
														X: &ast.Ident{
															Name: utils.GetPackageNameFromImportPath(g.globalsImportPath),
														},
														Sel: &ast.Ident{
															Name: "RequiredFieldMissingError",
														},
													},
													Elts: []ast.Expr{
														&ast.KeyValueExpr{
															Key: &ast.Ident{
																Name: "Path",
															},
															Value: &ast.Ident{Name: "\"" + path.String() + "\""},
														},
													},
												},
											},
										},
									},
								},
							},
						}
					} else {
						return nil
					}
				}(),
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
								Value: "\"" + g.fileData[fieldPath].JsonFieldName + "\"",
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
													Name: path.GetRune(),
												},
												Sel: &ast.Ident{
													Name: fieldPath.GetFieldName(),
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
														X: &ast.SelectorExpr{
															X: &ast.Ident{
																Name: utils.GetPackageNameFromImportPath(g.globalsImportPath),
															},
															Sel: &ast.Ident{
																Name: "AdditionalElementsError",
															},
														},
													},
												},
											},
										},
									},
									&ast.DeclStmt{
										Decl: &ast.GenDecl{
											Tok: token.VAR,
											Specs: []ast.Spec{
												&ast.ValueSpec{
													Names: []*ast.Ident{
														&ast.Ident{
															Name: "requiredFieldMissingError",
														},
													},
													Type: &ast.StarExpr{
														X: &ast.SelectorExpr{
															X: &ast.Ident{
																Name: utils.GetPackageNameFromImportPath(g.globalsImportPath),
															},
															Sel: &ast.Ident{
																Name: "RequiredFieldMissingError",
															},
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
										Else: &ast.IfStmt{
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
															Name: "requiredFieldMissingError",
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
																		Name: "requiredFieldMissingError",
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
										Value: "\"" + g.fileData[fieldPath].JsonFieldName + "\"",
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
										Type: &ast.SelectorExpr{X: &ast.Ident{Name: utils.GetPackageNameFromImportPath(g.globalsImportPath)}, Sel: &ast.Ident{Name: "AdditionalElementsError"}},
										Elts: []ast.Expr{
											&ast.KeyValueExpr{
												Key: &ast.Ident{
													Name: "ParsedObj",
												},
												Value: &ast.BasicLit{
													Kind:  token.STRING,
													Value: "\"" + path.GetFieldName() + "\"",
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
	imports := fieldData.Imports{
		&fieldData.Import{
			Path:            "encoding/json",
			Alias:           nil,
			NeedsAdjustment: false,
			IsGlobal:        false,
		},
		&fieldData.Import{
			Path:            "errors",
			Alias:           nil,
			NeedsAdjustment: false,
			IsGlobal:        false,
		},
		&fieldData.Import{
			Path:            "",
			Alias:           nil,
			NeedsAdjustment: false,
			IsGlobal:        true,
		},
	}
	return stmts, imports, nil
}

// TODO function to get struct and field name from path. Also respect the package in case one is given.
func (g *Generator) unmarshallHandleField(path fieldData.Path) []ast.Stmt {
	return []ast.Stmt{
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{
							{
								Name: "unmarshalledValue",
							},
						},
						Type: g.codeGenerator.GetFieldType(path, true),
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
						Name: path.GetParentRune(),
					},
					Sel: &ast.Ident{
						Name: path.GetFieldName(),
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
						Name: "UnMarshall" + path.GetFieldName(),
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
						Value: "\"" + g.fileData[path].JsonFieldName + "\"",
					},
				},
			},
		},
	}
}

func (g *Generator) unmarshallHandleArrayField(path fieldData.Path) []ast.Stmt {
	levelOfArrays := g.codeGenerator.GetLevelOfArrays(path)
	innerStmts := []ast.Stmt{
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
						Type: g.codeGenerator.GetFieldType(path, false),
					},
				},
			},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{
					Name: "result",
				},
				&ast.Ident{
					Name: "err",
				},
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.Ident{
						Name: "UnMarshall" + path.GetFieldName(),
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
							&ast.Ident{
								Name: "err",
							},
						},
					},
				},
			},
		},
		utils.GenerateAppendStatement(levelOfArrays-1, 0, &ast.SelectorExpr{X: &ast.Ident{Name: path.GetParentRune()}, Sel: &ast.Ident{Name: path.GetFieldName()}}, &ast.Ident{Name: "result"}, "index"),
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
						Value: "\"" + g.fileData[path].JsonFieldName + "\"",
					},
				},
			},
		},
	}
	return []ast.Stmt{
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{
							{
								Name: "lt",
							},
						},
						Type: g.codeGenerator.GetFieldType(path, false),
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
								Name: "lt",
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
		utils.GenerateNestedRangeStmt(levelOfArrays, innerStmts, &ast.Ident{Name: "lt"}, &ast.SelectorExpr{X: &ast.Ident{Name: path.GetParentRune()}, Sel: &ast.Ident{Name: path.GetFieldName()}}, g.codeGenerator.GetFieldType(path, true)),
	}
}
