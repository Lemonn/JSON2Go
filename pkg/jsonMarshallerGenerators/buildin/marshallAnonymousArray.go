package buildin

import (
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"go/token"
)

func (g *Generator) marshallArrayGenerator(path fieldData.Path) ([]ast.Stmt, fieldData.Imports, error) {
	var stmts []ast.Stmt
	levelOfArrays := g.codeGenerator.GetLevelOfArrays(path)
	originalExpr, imports, err := g.codeGenerator.GetOriginalFieldType(path)
	if err != nil {
		return nil, nil, err
	}
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
					Type: utils.GeneratedNestedArray(levelOfArrays, originalExpr),
				},
			},
		},
	})
	//TODO replace whit own function
	fieldStmts, err := g.marshallHandleArrayField(path)
	if err != nil {
		return nil, nil, err
	}
	stmts = append(stmts, fieldStmts...)
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

	return stmts, imports, nil
}
