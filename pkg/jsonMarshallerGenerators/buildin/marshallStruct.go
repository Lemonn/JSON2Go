package buildin

import (
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"go/token"
	"unicode"
)

func (g *Generator) generateShadowStruct(path fieldData.Path) *ast.DeclStmt {
	var localFields []*ast.Field
	var Level int
	for Level, _ = range g.fileData[path].Types[fieldData.Field] {
		break
	}
	for fieldPath, _ := range g.fileData[path].Types[fieldData.Field][Level] {
		localFields = append(localFields, &ast.Field{
			Names: []*ast.Ident{{Name: fieldPath.GetFieldName()}},
			Type:  g.codeGenerator.GetFieldType(fieldPath, false),
			Tag:   &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("`json:\"%s,omitempty\"`", g.fileData[fieldPath].JsonFieldName)},
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

func (g *Generator) marshallStructGenerator(path fieldData.Path) ([]ast.Stmt, fieldData.Imports, error) {
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
	for Level, _ = range g.fileData[path].Types[fieldData.Field] {
		break
	}
	for fieldPath, _ := range g.fileData[path].Types[fieldData.Field][Level] {
		levelOfArrays := g.codeGenerator.GetLevelOfArrays(fieldPath)
		if g.fileData[fieldPath].TypeAdjusterData != nil && g.fileData[fieldPath].ActiveType != nil && !g.codeGenerator.IsStruct(fieldPath) {
			required = true
			if levelOfArrays > 0 {
				g.marshallHandleArrayField(&stmts, fieldPath)
			} else {
				g.marshallHandleField(&stmts, fieldPath)
			}
		} else {
			stmts = append(stmts, &ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.SelectorExpr{
						X: &ast.Ident{
							Name: "lt",
						},
						Sel: &ast.Ident{
							Name: fieldPath.GetFieldName(),
						},
					},
				},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.SelectorExpr{
						X: &ast.Ident{
							Name: string(unicode.ToLower([]rune(path.GetFieldName())[0])),
						},
						Sel: &ast.Ident{
							Name: fieldPath.GetFieldName(),
						},
					},
				},
			})
		}

	}

	if len(stmts) == 0 || !required {
		return nil, nil, nil
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
	return stmts, fieldData.Imports{&fieldData.Import{
		Path:            "encoding/json",
		Alias:           nil,
		NeedsAdjustment: false,
		IsGlobal:        false,
	}}, nil
}

func (g *Generator) marshallHandleField(stmts *[]ast.Stmt, path fieldData.Path) {
	*stmts = append(*stmts, &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.SelectorExpr{
				X: &ast.Ident{
					Name: "lt",
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
					Name: "Marshall" + path.GetFieldName(),
				},
				Args: []ast.Expr{
					&ast.SelectorExpr{
						X: &ast.Ident{
							Name: string(unicode.ToLower([]rune(path.GetParentFieldName())[0])),
						},
						Sel: &ast.Ident{
							Name: path.GetFieldName(),
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

func (g *Generator) marshallHandleArrayField(stmts *[]ast.Stmt, path fieldData.Path) {
	var fieldNameExpr ast.Expr
	var structFieldNameIndexExpr ast.Expr
	levelOfArrays := g.codeGenerator.GetLevelOfArrays(path)

	fieldNameExpr = &ast.SelectorExpr{X: &ast.Ident{Name: "lt"}, Sel: &ast.Ident{Name: path.GetFieldName()}}
	structFieldNameIndexExpr = &ast.SelectorExpr{X: &ast.Ident{Name: string(unicode.ToLower([]rune(path.GetParentFieldName())[0]))}, Sel: &ast.Ident{Name: path.GetFieldName()}}

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
						Name: "Marshall" + path.GetFieldName(),
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
	*stmts = append(*stmts, utils.GenerateNestedRangeStmt(levelOfArrays, innerStmts, structFieldNameIndexExpr, fieldNameExpr, g.codeGenerator.GetFieldType(path, false)))
	return
}
