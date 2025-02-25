package marshaller

import (
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"go/token"
	"unicode"
)

func (g *Generator) generateShadowStruct(path string) *ast.DeclStmt {
	var localFields []*ast.Field
	var Level int
	for Level, _ = range g.seenTypes[path].Types[fieldData.Field] {
		break
	}
	for fieldPath, _ := range g.seenTypes[path].Types[fieldData.Field][Level] {
		localFields = append(localFields, &ast.Field{
			Names: []*ast.Ident{{Name: g.GetFieldName(fieldPath)}},
			Type:  g.GetFieldType(fieldPath),
			Tag:   &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("`json:\"%s,omitempty\"`", g.seenTypes[fieldPath].JsonFieldName)},
		})
	}

	return &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{
						Name: "localType",
					},
					Type: &ast.StructType{
						Fields: &ast.FieldList{
							List: localFields,
						},
					},
				},
			},
		},
	}
}

func (g *Generator) structGenerator(path string) ([]ast.Stmt, []string, error) {
	var stmts []ast.Stmt
	required := false

	localStruct := g.generateShadowStruct(path)

	stmts = append(stmts, &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						{
							Name: "err",
						},
					},
					Type: &ast.Ident{
						Name: "error",
					},
				},
			},
		},
	})
	stmts = append(stmts, localStruct)
	stmts = append(stmts, utils.GenerateTypeWhitInitializedArrays(localStruct.Decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)))

	var Level int
	for Level, _ = range g.seenTypes[path].Types[fieldData.Field] {
		break
	}
	for fieldPath, _ := range g.seenTypes[path].Types[fieldData.Field][Level] {
		levelOfArrays := g.GetLevelOfArrays(fieldPath)
		if g.seenTypes[fieldPath].TypeAdjusterData != nil && g.seenTypes[fieldPath].TypeAdjusterData.ActiveType != nil && !g.IsStruct(fieldPath) {
			required = true
			if levelOfArrays > 0 {
				g.handleArrayField(&stmts, fieldPath)
			} else {
				g.handleField(&stmts, fieldPath)
			}
		} else {
			stmts = append(stmts, &ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.SelectorExpr{
						X: &ast.Ident{
							Name: "lt",
						},
						Sel: &ast.Ident{
							Name: g.GetFieldName(fieldPath),
						},
					},
				},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.SelectorExpr{
						X: &ast.Ident{
							Name: string(unicode.ToLower([]rune(g.GetFieldName(path))[0])),
						},
						Sel: &ast.Ident{
							Name: g.GetFieldName(fieldPath),
						},
					},
				},
			})
		}

	}

	if len(stmts) == 0 || !required {
		return []ast.Stmt{}, []string{}, nil
	}

	stmts = append(stmts)
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
	return stmts, []string{"encoding/json"}, nil
}

func (g *Generator) handleField(stmts *[]ast.Stmt, path string) {
	*stmts = append(*stmts, &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.SelectorExpr{
				X: &ast.Ident{
					Name: "lt",
				},
				Sel: &ast.Ident{
					Name: g.GetFieldName(path),
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
					Name: "Marshall" + g.GetFieldName(path),
				},
				Args: []ast.Expr{
					&ast.SelectorExpr{
						X: &ast.Ident{
							Name: string(unicode.ToLower([]rune(g.GetParentFieldName(path))[0])),
						},
						Sel: &ast.Ident{
							Name: g.GetFieldName(path),
						},
					},
				},
			},
		},
	})
	*stmts = append(*stmts, &ast.IfStmt{
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
}

func (g *Generator) handleArrayField(stmts *[]ast.Stmt, path string) {
	var fieldNameExpr ast.Expr
	var structFieldNameIndexExpr ast.Expr
	levelOfArrays := g.GetLevelOfArrays(path)

	fieldNameExpr = &ast.SelectorExpr{X: &ast.Ident{Name: "lt"}, Sel: &ast.Ident{Name: g.GetFieldName(path)}}
	structFieldNameIndexExpr = &ast.SelectorExpr{X: &ast.Ident{Name: string(unicode.ToLower([]rune(g.GetParentFieldName(path))[0]))}, Sel: &ast.Ident{Name: g.GetFieldName(path)}}

	innerStmts := []ast.Stmt{
		&ast.DeclStmt{
			Decl: &ast.GenDecl{
				Tok: token.VAR,
				Specs: []ast.Spec{
					&ast.ValueSpec{
						Names: []*ast.Ident{
							{
								Name: "result",
							},
						},
						Type: g.GetBaseType(path),
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
						Name: "Marshall" + g.GetFieldName(path),
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
								Name: "nil",
							},
							&ast.Ident{
								Name: "err",
							},
						},
					},
				},
			},
		},
		utils.GenerateAppendStatement(levelOfArrays-1, 0, fieldNameExpr, &ast.Ident{Name: "result"}, "index"),
	}
	*stmts = append(*stmts, utils.GenerateNestedRangeStmt(levelOfArrays, innerStmts, structFieldNameIndexExpr, fieldNameExpr, g.GetBaseType(path)))
	return
}
