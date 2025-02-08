package marshaller

import (
	"errors"
	"github.com/Lemonn/AstUtils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/structGenerator"
	"go/ast"
	"go/token"
)

func (g *Generator) structGenerator(str *ast.StructType, path string) ([]ast.Stmt, []string, error) {
	var structFields []*ast.Field
	var stmts []ast.Stmt
	required := false
	for _, field := range str.Fields.List {

		var fData *fieldData.Data
		if v, ok := structGenerator.Tags[path+"."+field.Names[0].Name]; !ok {
			return nil, nil, errors.New("struct field not found")
		} else if v.JsonFieldName == nil {
			continue
		} else {
			fData = v
		}

		if fData.ParseFunctions != nil && fData.BaseType != nil {
			required = true
			structFields = append(structFields, &ast.Field{
				Names: []*ast.Ident{
					&ast.Ident{
						Name: field.Names[0].Name,
					},
				},
				Type: &ast.Ident{
					Name: *fData.BaseType,
				},
				Tag: AstUtils.RemoveTag("json2go", field.Tag),
			})
			stmts = append(stmts, &ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.SelectorExpr{
						X: &ast.Ident{
							Name: "v",
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
							Name: fData.ParseFunctions.ToTypeParseFunction,
						},
						Args: []ast.Expr{
							&ast.SelectorExpr{
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
			stmts = append(stmts, &ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.SelectorExpr{
						X: &ast.Ident{
							Name: "v",
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
							Name: "s",
						},
						Sel: &ast.Ident{
							Name: field.Names[0].Name,
						},
					},
				},
			})
		}
	}
	if len(stmts) == 0 || !required {
		return stmts, []string{}, nil
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
						Name: "v",
					},
				},
			},
		},
	})
	return stmts, []string{"encoding/json"}, nil
}
