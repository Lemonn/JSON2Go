package unmarshaller

import (
	"errors"
	"github.com/Lemonn/AstUtils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/structGenerator"
	"go/ast"
	"go/token"
)

func (g *Generator) structGenerator(str *ast.StructType, path string) ([]ast.Stmt, []string, error) {
	var stmts []ast.Stmt

	for _, field := range str.Fields.List {
		var fData *fieldData.Data
		var jsonName string
		if v, ok := structGenerator.Tags[path+"."+field.Names[0].Name]; !ok {
			return nil, nil, errors.New("struct field not found")
		} else if v.JsonFieldName == nil {
			continue
		} else {
			fData = v
			jsonName = *v.JsonFieldName
		}

		if fData.ParseFunctions != nil {
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
										Name: "s",
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
													Name: "s",
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
													Name: "s",
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

	if len(stmts) == 0 {
		return stmts, nil, nil
	}

	// Scaffold of custom unmarshall function
	f := &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				&ast.Field{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: "s",
						},
					},
					Type: &ast.StarExpr{
						X: &ast.Ident{
							Name: (*node.Parents[0]).(*ast.TypeSpec).Name.Name,
						},
					},
				},
			},
		},
		Name: &ast.Ident{
			Name: "UnmarshalJSON",
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "bytes",
							},
						},
						Type: &ast.ArrayType{
							Elt: &ast.Ident{
								Name: "byte",
							},
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
				&ast.DeclStmt{
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
				},
				&ast.DeclStmt{
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
				},
				&ast.AssignStmt{
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
			},
		},
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
													Value: "\"" + (*node.Parents[0]).(*ast.TypeSpec).Name.Name + "\"",
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

	// Add statements to custom unmarshall function
	for _, stmt := range stmts {
		f.Body.List = append(f.Body.List, stmt)
	}

	//Add custom json unmarshall function to file
	file.Decls = append(file.Decls, f)

	// Add AdditionalElementError + support methods
	file.Decls = append(file.Decls, &ast.GenDecl{
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
	file.Decls = append(file.Decls, &ast.FuncDecl{
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
	file.Decls = append(file.Decls, &ast.FuncDecl{
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
