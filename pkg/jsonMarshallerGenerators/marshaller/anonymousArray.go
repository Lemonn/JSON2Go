package marshaller

import (
	"go/ast"
)

func (g *Generator) arrayGenerator(path string) ([]ast.Stmt, []string, error) {
	return []ast.Stmt{}, []string{}, nil
	/*
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
						Type: utils.GeneratedNestedArray(levelOfArrays, g.GetBaseType(path)),
					},
				},
			},
		})
		//TODO replace whit own function
		g.handleArrayField(&stmts, levelOfArrays, "", fData, name, &ast.StarExpr{X: &ast.Ident{Name: string(unicode.ToLower([]rune(name)[0]))}})
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

		return stmts, nil, nil

	*/
}
