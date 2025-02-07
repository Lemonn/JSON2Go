package pkg

import (
	"fmt"
	"github.com/Lemonn/AstUtils"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
)

func x(levelOfArrays int, recursionCounter int, innerStmts []ast.Stmt, name string) *[]ast.Stmt {
	var localStmts []ast.Stmt

	varDecl := &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: "level" + strconv.Itoa(levelOfArrays),
						},
					},
					Type: func() ast.Expr {
						base := &ast.ArrayType{}
						ident := &ast.Ident{Name: "localDatatype"}
						if levelOfArrays == 0 {
							return ident
						}
						last := base
						for range levelOfArrays - 1 {
							last.Elt = &ast.ArrayType{}
							last = last.Elt.(*ast.ArrayType)
						}
						last.Elt = ident
						return base
					}(),
				},
			},
		},
	}

	localStmts = append(localStmts, varDecl)

	if recursionCounter == 0 && levelOfArrays == 0 {
		localStmts = append(localStmts, &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{
					Name: name + "1",
				},
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.Ident{
					Name: "s",
				},
			},
		})
	}

	if levelOfArrays == 0 {
		localStmts = append(localStmts, innerStmts...)
	}

	if levelOfArrays > 0 {
		localStmts = append(localStmts, &ast.RangeStmt{
			Key: &ast.Ident{
				Name: "_",
			},
			Value: &ast.Ident{
				Name: name + strconv.Itoa(levelOfArrays),
			},
			Tok: token.DEFINE,
			X: func() ast.Expr {
				if recursionCounter == 0 {
					return &ast.StarExpr{X: &ast.Ident{Name: name}}
				}
				return &ast.Ident{Name: name + strconv.Itoa(levelOfArrays+1)}
			}(),
			Body: &ast.BlockStmt{List: *x(levelOfArrays-1, recursionCounter+1, innerStmts, name)},
		})
	}

	if recursionCounter > 0 {
		localStmts = append(localStmts, &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{
					Name: "level" + strconv.Itoa(levelOfArrays+1),
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
							Name: "level" + strconv.Itoa(levelOfArrays+1),
						},
						&ast.Ident{
							Name: "level" + strconv.Itoa(levelOfArrays),
						},
					},
				},
			},
		})
	}
	if recursionCounter > 1 {
		localStmts = append(localStmts, &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X: &ast.Ident{
					Name: "level" + strconv.Itoa(levelOfArrays+1),
				},
				Op: token.EQL,
				Y: &ast.Ident{
					Name: "nil",
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{
							&ast.Ident{
								Name: "level" + strconv.Itoa(levelOfArrays+1),
							},
						},
						Tok: token.ASSIGN,
						Rhs: []ast.Expr{
							&ast.CompositeLit{
								Type: func() *ast.ArrayType {
									init := &ast.ArrayType{}
									last := init
									for range levelOfArrays {
										last.Elt = &ast.ArrayType{}
										last = last.Elt.(*ast.ArrayType)
									}
									last.Elt = &ast.Ident{
										Name: "localDatatype",
									}
									return init
								}(),
							},
						},
					},
				},
			},
		})
	}
	return &localStmts
}

func GenerateJsonMarshall(file *ast.File) error {
	var foundNodes []*AstUtils.FoundNodes
	var completed bool
	AstUtils.SearchNodes(file, &foundNodes, []*ast.Node{}, func(n *ast.Node, parents []*ast.Node, completed *bool) bool {
		if _, ok := (*n).(*ast.StructType); ok && len(parents) > 0 {
			return true
		} else if _, ok := (*n).(*ast.Ident); ok && len(parents) > 0 {
			if _, ok := (*parents[0]).(*ast.ArrayType); ok {
				return true
			}
		} else if _, ok := (*n).(*ast.SelectorExpr); ok && len(parents) > 0 {
			if _, ok := (*parents[0]).(*ast.ArrayType); ok {
				return true
			}
		}
		return false
	}, &completed)

	for _, node := range foundNodes {
		var structFields []*ast.Field
		var stmts []ast.Stmt
		var innerStmts []ast.Stmt
		var structName string
		var levelOfArrays int
		required := false
		var nested bool

		var path string
		for _, parent := range node.Parents {
			if v, ok := (*parent).(*ast.TypeSpec); ok {
				path += v.Name.Name
				structName = v.Name.Name
			} else if _, ok := (*parent).(*ast.StructType); ok {
				nested = true
				//Ignore nested structs
				break
			} else if _, ok := (*parent).(*ast.FuncDecl); ok {
				nested = true
				//Ignore structs inside functions
				break
			} else if _, ok := (*parent).(*ast.ArrayType); ok {
				levelOfArrays++
			}
		}
		if nested {
			continue
		}

		fmt.Println(levelOfArrays)

		switch (*node.Node).(type) {
		case *ast.StructType:
			fmt.Println("StructType")
		case *ast.Ident:

			stmts = append(stmts, handleIndent(path, levelOfArrays, (*node.Node).(*ast.Ident).Name)...)
		case *ast.SelectorExpr:
			fmt.Println("SelectorExpr")
			stmts = append(stmts, handleIndent(path, levelOfArrays, (*node.Node).(*ast.SelectorExpr).X.(*ast.Ident).Name+"."+(*node.Node).(*ast.SelectorExpr).Sel.Name)...)
		default:
			fmt.Println(reflect.TypeOf(*node.Node))

		}

		var _ = &ast.File{
			Package: 1,
			Name: &ast.Ident{
				Name: "main",
			},
			Decls: []ast.Decl{
				&ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							&ast.Field{
								Names: []*ast.Ident{
									&ast.Ident{
										Name: "R",
									},
								},
								Type: &ast.StarExpr{
									X: &ast.Ident{
										Name: "RRR",
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
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.StarExpr{
										X: &ast.Ident{
											Name: "R",
										},
									},
								},
								Tok: token.ASSIGN,
								Rhs: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.Ident{
											Name: "append",
										},
										Args: []ast.Expr{
											&ast.StarExpr{
												X: &ast.Ident{
													Name: "R",
												},
											},
											&ast.Ident{Name: "level0"},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		f1 := &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "R",
							},
						},
						Type: &ast.StarExpr{
							X: &ast.Ident{
								Name: "RRR",
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
			Body: &ast.BlockStmt{List: stmts},
		}
		file.Decls = append(file.Decls, f1)

		return nil

		for _, field := range (*node.Node).(*ast.StructType).Fields.List {
			lit := Tags[path+"."+field.Names[0].Name]
			//lit, err := GetJson2GoTagFromBasicLit(field.Tag)
			//if err != nil {
			//	return err
			//}
			fmt.Println(lit)
			fmt.Println(path + "." + field.Names[0].Name)
			if lit != nil && lit.ParseFunctions != nil && lit.BaseType != nil {
				required = true
				structFields = append(structFields, &ast.Field{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: field.Names[0].Name,
						},
					},
					Type: &ast.Ident{
						Name: *lit.BaseType,
					},
					Tag: AstUtils.RemoveTag("json2go", field.Tag),
				})
				innerStmts = append(innerStmts, &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.SelectorExpr{
							X: &ast.Ident{
								Name: "level0",
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
								Name: lit.ParseFunctions.ToTypeParseFunction,
							},
							Args: []ast.Expr{
								&ast.SelectorExpr{
									X: &ast.Ident{
										Name: "s1",
									},
									Sel: &ast.Ident{
										Name: field.Names[0].Name,
									},
								},
							},
						},
					},
				})
				innerStmts = append(innerStmts, &ast.IfStmt{
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
										Name: "nil",
									},
									&ast.Ident{
										Name: "err",
									},
								},
							},
						},
					},
				})
			} else {
				structFields = append(structFields, &ast.Field{
					Doc:     field.Doc,
					Names:   field.Names,
					Type:    field.Type,
					Tag:     AstUtils.RemoveTag("json2go", field.Tag),
					Comment: field.Comment,
				})
				innerStmts = append(innerStmts, &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.SelectorExpr{
							X: &ast.Ident{
								Name: "level0",
							},
							Sel: &ast.Ident{
								Name: field.Names[0].Name,
							},
						},
					},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.SelectorExpr{
							X: &ast.Ident{
								Name: "s1",
							},
							Sel: &ast.Ident{
								Name: field.Names[0].Name,
							},
						},
					},
				})
			}
		}

		stmt := x(levelOfArrays, 0, innerStmts, "s")

		stmts = append(stmts, *stmt...)

		if len(stmts) == 0 || !required {
			continue
		}
		stmts = append(stmts, &ast.ReturnStmt{
			Results: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: "json",
						},
						Sel: &ast.Ident{
							Name: "Marshal",
						},
					},
					Args: []ast.Expr{
						&ast.Ident{
							Name: "level" + strconv.Itoa(levelOfArrays),
						},
					},
				},
			},
		})
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
								Name: structName,
							},
						},
					},
				},
			},
			Name: &ast.Ident{
				Name: "MarshalJSON",
			},
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
				Results: &ast.FieldList{
					List: []*ast.Field{
						&ast.Field{
							Type: &ast.ArrayType{
								Elt: &ast.Ident{
									Name: "byte",
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
											Name: "err",
										},
									},
									Type: &ast.Ident{
										Name: "error",
									},
								},
							},
						},
					},
					&ast.DeclStmt{
						Decl: &ast.GenDecl{
							Tok: token.TYPE,
							Specs: []ast.Spec{
								&ast.TypeSpec{
									Name: &ast.Ident{
										Name: "localDatatype",
									},
									Type: &ast.StructType{
										Fields: &ast.FieldList{
											List: structFields,
										},
									},
								},
							},
						},
					},
				},
			},
		}
		f.Body.List = append(f.Body.List, stmts...)
		file.Decls = append(file.Decls, f)
	}
	AstUtils.AddMissingImports(file, []string{"encoding/json"})
	return nil
}

func handleIndent(path string, levelOfArrays int, fieldType string) []ast.Stmt {
	var stmts []ast.Stmt
	fmt.Println("Ident")
	lit := Tags[path]
	if lit == nil || lit.BaseType == nil {
		return stmts
	}

	var OuterExpr ast.Expr
	var InnerExpr *ast.Expr
	InnerExpr, OuterExpr = GeneratedNestedArray(levelOfArrays, InnerExpr, OuterExpr)

	(*InnerExpr).(*ast.ArrayType).Elt = &ast.Ident{
		Name: *lit.BaseType,
	}

	stmts = append(stmts, &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{
						Name: "localType",
					},
					Type: OuterExpr,
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
							Name: "lt",
						},
					},
					Type: &ast.Ident{
						Name: "localType",
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
							Name: "lt",
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

	if levelOfArrays-1 > 0 {
		stmts = append(stmts, &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.StarExpr{
					X: &ast.Ident{
						Name: "R",
					},
				},
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.Ident{
						Name: "append",
					},
					Args: []ast.Expr{
						&ast.StarExpr{
							X: &ast.Ident{
								Name: "R",
							},
						},
						&ast.CompositeLit{
							Type: func() ast.Expr {
								ident := &ast.Ident{Name: fieldType}
								var oe ast.Expr
								var ie *ast.Expr
								ie, oe = GeneratedNestedArray(levelOfArrays-1, ie, oe)
								(*ie).(*ast.ArrayType).Elt = ident
								return oe
							}(),
						},
					},
				},
			},
		})
	}

	var OuterStmt ast.Stmt
	var InnerStmt *ast.Stmt
	for i := range levelOfArrays {
		if InnerStmt != nil && reflect.TypeOf(*InnerStmt) == reflect.TypeOf(&ast.RangeStmt{}) {
			if i != levelOfArrays-1 {
				(*InnerStmt).(*ast.RangeStmt).Body.List = append((*InnerStmt).(*ast.RangeStmt).Body.List, &ast.AssignStmt{
					Lhs: []ast.Expr{
						func() ast.Expr {
							var oe ast.Expr
							var ie *ast.Expr
							for range i {
								if ie != nil && reflect.TypeOf(*ie) == reflect.TypeOf(&ast.IndexExpr{}) {
									(*ie).(*ast.IndexExpr).X = &ast.IndexExpr{}
									*ie = (*ie).(*ast.IndexExpr).X
								} else {
									var k ast.Expr
									k = &ast.IndexExpr{
										X: &ast.ParenExpr{
											X: &ast.StarExpr{
												X: &ast.Ident{
													Name: "R",
												},
											},
										},
										Index: &ast.Ident{
											//TODO fasdf
											Name: func() string {

												return "i" + strconv.Itoa(i-1)
											}(),
										},
									}
									ie = &k
									oe = k
								}
							}
							(*ie).(*ast.IndexExpr).Index = &ast.Ident{
								Name: "i" + strconv.Itoa(i-1),
							}
							(*ie).(*ast.IndexExpr).X = &ast.ParenExpr{
								X: &ast.StarExpr{
									X: &ast.Ident{
										Name: "R",
									},
								},
							}
							return oe

						}(),
					},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.Ident{
								Name: "append",
							},
							Args: []ast.Expr{
								func() ast.Expr {
									var oe ast.Expr
									var ie *ast.Expr
									for range i {
										if ie != nil && reflect.TypeOf(*ie) == reflect.TypeOf(&ast.IndexExpr{}) {
											(*ie).(*ast.IndexExpr).X = &ast.IndexExpr{}
											*ie = (*ie).(*ast.IndexExpr).X
										} else {
											var k ast.Expr
											k = &ast.IndexExpr{
												X: &ast.ParenExpr{
													X: &ast.StarExpr{
														X: &ast.Ident{
															Name: "R",
														},
													},
												},
												Index: &ast.Ident{
													Name: "i" + strconv.Itoa(i-1),
												},
											}
											ie = &k
											oe = k
										}
									}
									(*ie).(*ast.IndexExpr).Index = &ast.Ident{
										Name: "i" + strconv.Itoa(i-1),
									}
									(*ie).(*ast.IndexExpr).X = &ast.ParenExpr{
										X: &ast.StarExpr{
											X: &ast.Ident{
												Name: "R",
											},
										},
									}
									return oe

								}(),
								&ast.CompositeLit{
									Type: func() ast.Expr {
										ident := &ast.Ident{Name: fieldType}
										if levelOfArrays-i == 0 {
											return ident
										}
										var oe ast.Expr
										var ie *ast.Expr
										ie, oe = GeneratedNestedArray(levelOfArrays-i-1, ie, oe)
										(*ie).(*ast.ArrayType).Elt = ident
										return oe
									}(),
								},
							},
						},
					},
				})
			}
			(*InnerStmt).(*ast.RangeStmt).Body.List = append((*InnerStmt).(*ast.RangeStmt).Body.List, &ast.RangeStmt{
				Key: &ast.Ident{
					Name: func() string {
						fmt.Println(levelOfArrays - 1)
						if i == levelOfArrays-1 {
							return "_"
						}
						return "i" + strconv.Itoa(i)
					}(),
				},
				Value: &ast.Ident{
					Name: "level" + strconv.Itoa(i),
				},
				Tok: token.DEFINE,
				X: &ast.Ident{
					Name: "level" + strconv.Itoa(i-1),
				},
				Body: &ast.BlockStmt{},
			})
			*InnerStmt = (*InnerStmt).(*ast.RangeStmt).Body.List[len((*InnerStmt).(*ast.RangeStmt).Body.List)-1]
		} else {
			var k ast.Stmt
			k = &ast.RangeStmt{
				Key: &ast.Ident{
					Name: func() string {
						if levelOfArrays == 1 {
							return "_"
						}
						return "i" + strconv.Itoa(i)
					}(),
				},
				Value: &ast.Ident{
					Name: "level" + strconv.Itoa(i),
				},
				Tok: token.DEFINE,
				X: &ast.Ident{
					Name: "lt",
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{},
				},
			}
			InnerStmt = &k
			OuterStmt = k
		}
	}
	(*InnerStmt).(*ast.RangeStmt).Body.List = append((*InnerStmt).(*ast.RangeStmt).Body.List, &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "value",
			},
			&ast.Ident{
				Name: "err",
			},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.Ident{
					Name: "fromRRR",
				},
				Args: []ast.Expr{
					&ast.Ident{
						Name: "level" + strconv.Itoa(levelOfArrays-1),
					},
				},
			},
		},
	})
	(*InnerStmt).(*ast.RangeStmt).Body.List = append((*InnerStmt).(*ast.RangeStmt).Body.List, &ast.IfStmt{
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
	(*InnerStmt).(*ast.RangeStmt).Body.List = append((*InnerStmt).(*ast.RangeStmt).Body.List, &ast.AssignStmt{
		Lhs: []ast.Expr{
			func() ast.Expr {
				var oe ast.Expr
				var ie *ast.Expr
				for range levelOfArrays - 1 {
					if ie != nil && reflect.TypeOf(*ie) == reflect.TypeOf(&ast.IndexExpr{}) {
						(*ie).(*ast.IndexExpr).X = &ast.IndexExpr{}
						(*ie).(*ast.IndexExpr).Index = &ast.Ident{
							Name: "i0",
						}
						*ie = (*ie).(*ast.IndexExpr).X
					} else {
						var k ast.Expr
						k = &ast.IndexExpr{}
						ie = &k
						oe = k
					}
				}
				if ie == nil {
					return &ast.StarExpr{
						X: &ast.Ident{
							Name: "R",
						},
					}
				}
				(*ie).(*ast.IndexExpr).X = &ast.ParenExpr{
					X: &ast.StarExpr{
						X: &ast.Ident{
							Name: "R",
						},
					},
				}
				(*ie).(*ast.IndexExpr).Index = &ast.Ident{
					Name: "i0",
				}
				return oe
			}(),
		},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.Ident{
					Name: "append",
				},
				Args: []ast.Expr{

					func() ast.Expr {
						var oe ast.Expr
						var ie *ast.Expr
						for range levelOfArrays - 1 {
							if ie != nil && reflect.TypeOf(*ie) == reflect.TypeOf(&ast.IndexExpr{}) {
								(*ie).(*ast.IndexExpr).X = &ast.IndexExpr{}
								(*ie).(*ast.IndexExpr).Index = &ast.Ident{
									Name: "i0",
								}
								*ie = (*ie).(*ast.IndexExpr).X
							} else {
								var k ast.Expr
								k = &ast.IndexExpr{}
								ie = &k
								oe = k
							}
						}
						if ie == nil {
							return &ast.StarExpr{
								X: &ast.Ident{
									Name: "R",
								},
							}
						}
						(*ie).(*ast.IndexExpr).X = &ast.ParenExpr{
							X: &ast.StarExpr{
								X: &ast.Ident{
									Name: "R",
								},
							},
						}
						(*ie).(*ast.IndexExpr).Index = &ast.Ident{
							Name: "i0",
						}
						return oe
					}(),
					&ast.Ident{Name: "value"},
				},
			},
		},
	})
	stmts = append(stmts, OuterStmt)
	return stmts
}

func GeneratedNestedArray(levelOfArrays int, InnerExpr *ast.Expr, OuterExpr ast.Expr) (*ast.Expr, ast.Expr) {
	for range levelOfArrays {
		if InnerExpr != nil && reflect.TypeOf(*InnerExpr) == reflect.TypeOf(&ast.ArrayType{}) {
			(*InnerExpr).(*ast.ArrayType).Elt = &ast.ArrayType{}
			*InnerExpr = (*InnerExpr).(*ast.ArrayType).Elt
		} else {
			var k ast.Expr
			k = &ast.ArrayType{}
			InnerExpr = &k
			OuterExpr = k
		}
	}
	return InnerExpr, OuterExpr
}

func GenerateJsonUnmarshall(file *ast.File) error {
	var foundNodes []*AstUtils.FoundNodes
	var completed bool
	AstUtils.SearchNodes(file, &foundNodes, []*ast.Node{}, func(n *ast.Node, parents []*ast.Node, completed *bool) bool {
		if _, ok := (*n).(*ast.StructType); ok && len(parents) > 0 {
			return true
		}
		return false
	}, &completed)

	for _, node := range foundNodes {

		var structName string
		var levelOfArrays int
		var nested bool
		var path string
		for _, parent := range node.Parents {
			if v, ok := (*parent).(*ast.TypeSpec); ok {
				path += v.Name.Name

				structName = v.Name.Name

			} else if _, ok := (*parent).(*ast.StructType); ok {
				nested = true
				//Ignore nested structs
				break
			} else if _, ok := (*parent).(*ast.FuncDecl); ok {
				nested = true
				//Ignore structs inside functions
				break
			} else if _, ok := (*parent).(*ast.ArrayType); ok {
				levelOfArrays++
			}
		}
		if nested {
			continue
		}

		var stmts []ast.Stmt
		var structFields []*ast.Field
		var innerStmts []ast.Stmt
		for _, field := range (*node.Node).(*ast.StructType).Fields.List {

			structFields = append(structFields, field)
			lit := Tags[path+"."+field.Names[0].Name]
			if lit == nil {
				continue
			}
			if lit.JsonFieldName == nil {
				continue
			}
			jsonName := *lit.JsonFieldName

			//lit, err := GetJson2GoTagFromBasicLit(field.Tag)
			if lit.ParseFunctions != nil {
				innerStmts = append(innerStmts, &ast.IfStmt{
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
									Name: "s1",
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
												Name: *lit.BaseType,
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
											Name: "level0",
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
											Name: lit.ParseFunctions.FromTypeParseFunction,
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
											Name: "s1",
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
				innerStmts = append(innerStmts, &ast.IfStmt{
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
									Name: "s1",
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
														Name: "level0",
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
											Name: "s1",
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
				innerStmts = append(innerStmts, &ast.IfStmt{
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
									Name: "s1",
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
														Name: "level0",
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
											Name: "s1",
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

		if len(innerStmts) == 0 {
			continue
		}

		var _ = &ast.File{
			Package: 1,
			Name: &ast.Ident{
				Name: "main",
			},
			Decls: []ast.Decl{
				&ast.FuncDecl{
					Name: &ast.Ident{
						Name: "Test",
					},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.DeclStmt{
								Decl: &ast.GenDecl{
									Tok: token.TYPE,
									Specs: []ast.Spec{
										&ast.TypeSpec{
											Name: &ast.Ident{
												Name: "localDatatype",
											},
											Type: &ast.StructType{
												Fields: &ast.FieldList{
													List: []*ast.Field{
														&ast.Field{
															Names: []*ast.Ident{
																&ast.Ident{
																	Name: "T",
																},
															},
															Type: &ast.Ident{
																Name: "int",
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

		stmts = append(stmts, &ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.TYPE,
				Specs: []ast.Spec{
					&ast.TypeSpec{
						Name: &ast.Ident{
							Name: "localDatatype",
						},
						Type: &ast.StructType{
							Fields: &ast.FieldList{
								List: structFields,
							},
						},
					},
				},
			},
		})

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
								Name: structName,
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
									Type: func() ast.Expr {
										mapType := &ast.MapType{
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
										}
										if levelOfArrays == 0 {
											return mapType
										}
										init := &ast.ArrayType{}
										last := init
										for range levelOfArrays - 1 {
											last.Elt = &ast.ArrayType{}
											last = last.Elt.(*ast.ArrayType)
										}
										last.Elt = mapType
										return init
									}(),
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
		s := x(levelOfArrays, 0, innerStmts, "s")
		stmts = append(stmts, *s...)

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
														Value: "\"" + structName + "\"",
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
	}
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
	addGetAllErrorsOfTypeFunction(file)
	addCheckForFirstErrorNotOfTypeTFunction(file)
	AstUtils.AddMissingImports(file, []string{"encoding/json", "errors", "fmt"})
	return nil
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
