package unmarshaller

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
)

// Handles the case where we have an array of non struct type
func (g *Generator) arrayGenerator(path string, levelOfArrays int, fieldType string) ([]ast.Stmt, []string) {
	var stmts []ast.Stmt

	lit := g.data[path]
	if lit == nil || lit.BaseType == nil {
		return stmts, []string{}
	}

	var OuterExpr ast.Expr
	var InnerExpr *ast.Expr

	//Generate nested array for local type
	InnerExpr, OuterExpr = GeneratedNestedArray(levelOfArrays, InnerExpr, OuterExpr)

	//Set type of local type
	(*InnerExpr).(*ast.ArrayType).Elt = &ast.Ident{
		Name: *lit.BaseType,
	}

	//Append local type to file
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

	//Append var lt localType to file
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

	//Append json unmarshall call to file
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

	//Append json error handler to file
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

	//Generate nested for loops
	var OuterStmt ast.Stmt
	var InnerStmt *ast.Stmt
	for i := range levelOfArrays {
		if InnerStmt != nil && reflect.TypeOf(*InnerStmt) == reflect.TypeOf(&ast.RangeStmt{}) {
			//if i != levelOfArrays-1 {
			(*InnerStmt).(*ast.RangeStmt).Body.List = append((*InnerStmt).(*ast.RangeStmt).Body.List, &ast.AssignStmt{
				Lhs: []ast.Expr{
					func() ast.Expr {
						var oe ast.Expr
						var ie *ast.Expr
						fmt.Println(fmt.Sprintf("S %d", i))
						li := 0
						for range i - 1 {
							if ie != nil && reflect.TypeOf(*ie) == reflect.TypeOf(&ast.IndexExpr{}) {
								(*ie).(*ast.IndexExpr).X = &ast.IndexExpr{}
								(*ie).(*ast.IndexExpr).Index = &ast.Ident{Name: "i" + strconv.Itoa(i-li-1)}
								*ie = (*ie).(*ast.IndexExpr).X
							} else {
								var k ast.Expr
								k = &ast.IndexExpr{}
								ie = &k
								oe = k
							}
							li++

						}
						if i > 1 {
							(*ie).(*ast.IndexExpr).Index = &ast.Ident{
								Name: "i" + strconv.Itoa(i-li-1),
							}
							(*ie).(*ast.IndexExpr).X = &ast.ParenExpr{
								X: &ast.StarExpr{
									X: &ast.Ident{
										Name: "R",
									},
								},
							}
						} else {
							return &ast.StarExpr{X: &ast.Ident{Name: "R"}}
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
								fmt.Println(fmt.Sprintf("S %d", i))
								li := 0
								for range i - 1 {
									if ie != nil && reflect.TypeOf(*ie) == reflect.TypeOf(&ast.IndexExpr{}) {
										(*ie).(*ast.IndexExpr).X = &ast.IndexExpr{}
										(*ie).(*ast.IndexExpr).Index = &ast.Ident{Name: "i" + strconv.Itoa(i-li-1)}
										*ie = (*ie).(*ast.IndexExpr).X
									} else {
										var k ast.Expr
										k = &ast.IndexExpr{}
										ie = &k
										oe = k
									}
									li++

								}
								if i > 1 {
									(*ie).(*ast.IndexExpr).Index = &ast.Ident{
										Name: "i" + strconv.Itoa(i-li-1),
									}
									(*ie).(*ast.IndexExpr).X = &ast.ParenExpr{
										X: &ast.StarExpr{
											X: &ast.Ident{
												Name: "R",
											},
										},
									}
								} else {
									return &ast.StarExpr{X: &ast.Ident{Name: "R"}}
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
									ie, oe = GeneratedNestedArray(levelOfArrays-i, ie, oe)
									(*ie).(*ast.ArrayType).Elt = ident
									return oe
								}(),
							},
						},
					},
				},
			})
			//}
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

	if levelOfArrays > 0 {
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
					li := levelOfArrays - 1
					for range levelOfArrays - 1 {
						if ie != nil && reflect.TypeOf(*ie) == reflect.TypeOf(&ast.IndexExpr{}) {
							(*ie).(*ast.IndexExpr).X = &ast.IndexExpr{}
							(*ie).(*ast.IndexExpr).Index = &ast.Ident{
								Name: "i" + strconv.Itoa(li),
							}
							*ie = (*ie).(*ast.IndexExpr).X
						} else {
							var k ast.Expr
							k = &ast.IndexExpr{}
							ie = &k
							oe = k
						}
						li--
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
							li := levelOfArrays - 1
							for range levelOfArrays - 1 {
								if ie != nil && reflect.TypeOf(*ie) == reflect.TypeOf(&ast.IndexExpr{}) {
									(*ie).(*ast.IndexExpr).X = &ast.IndexExpr{}
									(*ie).(*ast.IndexExpr).Index = &ast.Ident{
										Name: "i" + strconv.Itoa(li),
									}
									*ie = (*ie).(*ast.IndexExpr).X
								} else {
									var k ast.Expr
									k = &ast.IndexExpr{}
									ie = &k
									oe = k
								}
								li--
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
								Name: "i" + strconv.Itoa(0),
							}
							return oe
						}(),
						&ast.Ident{Name: "value"},
					},
				},
			},
		})
		stmts = append(stmts, OuterStmt)
	}
	stmts = append(stmts, &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.Ident{
				Name: "nil",
			},
		},
	})
	return stmts, []string{"encoding/json"}
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
