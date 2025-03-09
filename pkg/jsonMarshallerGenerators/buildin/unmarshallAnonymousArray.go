package buildin

import (
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"go/token"
	"unicode"
)

// Handles the case where we have an array of non struct type
func (g *Generator) unmarshallArrayGenerator(path fieldData.Path) ([]ast.Stmt, fieldData.Imports, error) {
	var stmts []ast.Stmt
	levelOfArrays := g.codeGenerator.GetLevelOfArrays(path)
	//Content of the nested range statement
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
						Type: g.codeGenerator.GetFieldType(path, true),
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
		utils.GenerateAppendStatement(levelOfArrays-1, 0, &ast.StarExpr{X: &ast.Ident{Name: string(unicode.ToLower([]rune(path.GetFieldName())[0]))}}, &ast.Ident{Name: "result"}, "index"),
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
					//TODO replace whit original type
					Type: utils.GeneratedNestedArray(levelOfArrays, g.codeGenerator.GetFieldType(path, true)),
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
	stmts = append(stmts, &ast.AssignStmt{
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
	stmts = append(stmts, utils.GenerateNestedRangeStmt(levelOfArrays, innerStmts, &ast.Ident{Name: "lt"}, &ast.StarExpr{X: &ast.Ident{Name: path.GetRune()}}, g.codeGenerator.GetFieldType(path, true)))
	stmts = append(stmts, &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.Ident{
				Name: "nil",
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
