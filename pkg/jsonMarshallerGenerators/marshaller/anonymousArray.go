package marshaller

import (
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
	"unicode"
)

func (g *Generator) arrayGenerator(path string, levelOfArrays int, name string) ([]ast.Stmt, []string, error) {
	var stmts []ast.Stmt
	var fData *fieldData.FieldData
	if v, ok := g.data[path]; !ok || v.BaseType == nil || levelOfArrays <= 0 {
		return nil, nil, nil
	} else {
		fData = v
	}
	stmts = append(stmts, &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						{
							Name: "lt",
						},
					},
					Type: func() ast.Expr {
						ident := &ast.Ident{Name: *fData.BaseType}
						if levelOfArrays == 0 {
							return ident
						}
						var oe ast.Expr
						var ie *ast.Expr
						ie, oe = utils.GeneratedNestedArray(levelOfArrays, ie, oe)
						(*ie).(*ast.ArrayType).Elt = ident
						return oe
					}(),
				},
			},
		},
	})

	var os ast.Stmt
	var is *ast.Stmt
	for i := range levelOfArrays {
		if is != nil && reflect.TypeOf(*is) == reflect.TypeOf(&ast.RangeStmt{}) {
			(*is).(*ast.RangeStmt).Body.List = append((*is).(*ast.RangeStmt).Body.List, &ast.AssignStmt{
				Lhs: []ast.Expr{
					g.generateIndexing(i),
				},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.Ident{
							Name: "append",
						},
						Args: []ast.Expr{
							g.generateIndexing(i),
							&ast.CompositeLit{
								Type: func() ast.Expr {
									ident := &ast.Ident{Name: *fData.BaseType}
									if levelOfArrays-i == 0 {
										return ident
									}
									var oe ast.Expr
									var ie *ast.Expr
									ie, oe = utils.GeneratedNestedArray(levelOfArrays-i, ie, oe)
									(*ie).(*ast.ArrayType).Elt = ident
									return oe
								}(),
							},
						},
					},
				},
			})
			(*is).(*ast.RangeStmt).Body.List = append((*is).(*ast.RangeStmt).Body.List, &ast.RangeStmt{
				Key: &ast.Ident{
					Name: func() string {
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
			*is = (*is).(*ast.RangeStmt).Body.List[len((*is).(*ast.RangeStmt).Body.List)-1]
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
				X:   &ast.StarExpr{X: &ast.Ident{Name: string(unicode.ToLower([]rune(name)[0]))}},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{},
				},
			}
			os = k
			is = &k
		}
	}

	(*is).(*ast.RangeStmt).Body.List = append((*is).(*ast.RangeStmt).Body.List, &ast.AssignStmt{
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
					Name: "toRRR",
				},
				Args: []ast.Expr{
					&ast.Ident{
						Name: "level" + strconv.Itoa(levelOfArrays-1),
					},
				},
			},
		},
	})
	(*is).(*ast.RangeStmt).Body.List = append((*is).(*ast.RangeStmt).Body.List, &ast.IfStmt{
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
	(*is).(*ast.RangeStmt).Body.List = append((*is).(*ast.RangeStmt).Body.List, &ast.AssignStmt{
		Lhs: []ast.Expr{
			g.generateIndexing(levelOfArrays),
		},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.Ident{
					Name: "append",
				},
				Args: []ast.Expr{
					g.generateIndexing(levelOfArrays),
					&ast.Ident{Name: "value"},
				},
			},
		},
	})

	stmts = append(stmts, os)
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
						Name: "lt",
					},
				},
			},
		},
	})
	return stmts, []string{}, nil
}

func (g *Generator) generateIndexing(levelOfArrays int) ast.Expr {
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
		return &ast.Ident{
			Name: "lt",
		}
	}
	(*ie).(*ast.IndexExpr).X = &ast.Ident{
		Name: "lt",
	}
	(*ie).(*ast.IndexExpr).Index = &ast.Ident{
		Name: "i" + strconv.Itoa(0),
	}
	return oe
}
