package marshaller

import (
	"errors"
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"go/token"
	"reflect"
	"unicode"
)

func (g *Generator) structGenerator(str *ast.StructType, path string, name string) ([]ast.Stmt, []string, error) {
	var localStruct *ast.DeclStmt
	var err error
	var stmts []ast.Stmt
	required := false

	localStruct, err = structTypeFromFields(str.Fields.List, path, g.data)
	if err != nil {
		return nil, nil, err
	}
	stmts = append(stmts, &ast.DeclStmt{
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
	})
	stmts = append(stmts, localStruct)
	stmts = append(stmts, utils.GenerateTypeWhitInitializedArrays(localStruct.Decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)))

	for _, field := range str.Fields.List {

		var fData *fieldData.FieldData
		if v, ok := g.data[path+"."+field.Names[0].Name]; !ok {
			return nil, nil, errors.New(fmt.Sprintf("struct field not found, path: %s", path+"."+field.Names[0].Name))
		} else if v.JsonFieldName == nil {
			continue
		} else {
			fData = v
		}

		if fData.ParseFunctions != nil && fData.BaseType != nil {
			required = true
			if levelOfArrays := utils.GetLevelOfArrays(field.Type); levelOfArrays > 0 {
				fieldType, err := utils.WalkArrays(&field.Type)
				if err != nil {
					return nil, nil, err
				}
				g.handleArrayField(&stmts, levelOfArrays, field.Names[0].Name, fData, name, *fieldType)
			} else {
				g.handleField(&stmts, name, field.Names[0].Name, fData)
			}
		} else {
			stmts = append(stmts, &ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.SelectorExpr{
						X: &ast.Ident{
							Name: "lt",
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
							Name: string(unicode.ToLower([]rune(name)[0])),
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

func (g *Generator) handleField(stmts *[]ast.Stmt, structName string, fieldName string, fData *fieldData.FieldData) {
	*stmts = append(*stmts, &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.SelectorExpr{
				X: &ast.Ident{
					Name: "lt",
				},
				Sel: &ast.Ident{
					Name: fieldName,
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
							Name: string(unicode.ToLower([]rune(structName)[0])),
						},
						Sel: &ast.Ident{
							Name: fieldName,
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

func (g *Generator) handleArrayField(stmts *[]ast.Stmt, levelOfArrays int, name string, fData *fieldData.FieldData, structName string, fieldType ast.Expr) {
	innerStmts := []ast.Stmt{
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
						Name: fData.ParseFunctions.ToTypeParseFunction,
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
		utils.GenerateAppendStatement(levelOfArrays-1, 0, &ast.SelectorExpr{X: &ast.Ident{Name: "lt"}, Sel: &ast.Ident{Name: name}}, &ast.Ident{Name: "result"}, "index"),
	}
	*stmts = append(*stmts, utils.GenerateNestedRangeStmt(levelOfArrays, innerStmts, &ast.SelectorExpr{X: &ast.Ident{Name: string(unicode.ToLower([]rune(structName)[0]))}, Sel: &ast.Ident{Name: name}}, &ast.SelectorExpr{X: &ast.Ident{Name: "lt"}, Sel: &ast.Ident{Name: name}}, &ast.Ident{Name: "string"}))
	return
}

func structTypeFromFields(fields []*ast.Field, path string, tags map[string]*fieldData.FieldData) (*ast.DeclStmt, error) {
	var localFields []*ast.Field
	for _, field := range fields {

		if v, ok := tags[path+"."+field.Names[0].Name]; ok && v.BaseType != nil {
			var levelOfArrays int
			currentType := field.Type
			var finished bool
			for !finished {
				switch currentType.(type) {
				case *ast.StarExpr:
					if _, ok := currentType.(*ast.StarExpr).X.(*ast.Ident); !ok {
						return nil, errors.New(fmt.Sprintf("only *ast.StarExpr expression whit an nestet"+
							" *ast.Ident are supported, this one contains an: %s",
							reflect.TypeOf(currentType.(*ast.StarExpr).X)))
					}
					currentType = currentType.(*ast.StarExpr).X
					finished = true
				case *ast.Ident:
					finished = true
				case *ast.ArrayType:
					currentType = currentType.(*ast.ArrayType).Elt
					levelOfArrays++
				case *ast.SelectorExpr:
					finished = true
				case *ast.InterfaceType:
					finished = true
				default:
					return nil, errors.New(fmt.Sprintf("only StarExpr, Ident, SelectorExpr, ArrayType or InterfaceType are"+
						" supported, this expression is of type: %s. Current path: %s",
						reflect.TypeOf(currentType), path+"."+field.Names[0].Name))
				}
			}

			// TODO if we got an *ast.StarExpr we currently replace it with an *ast.Ident, keep it an *ast.StarExpr
			localFields = append(localFields,
				&ast.Field{
					Doc:     field.Doc,
					Names:   field.Names,
					Type:    utils.GeneratedNestedArray(levelOfArrays, utils.GetTypeFromBaseType(*v.BaseType)),
					Tag:     field.Tag,
					Comment: field.Comment,
				})

		} else {
			localFields = append(localFields, field)
		}

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
	}, nil
}
