package JSON2Go

import (
	"fmt"
	"github.com/Lemonn/AstUtils"
	"go/ast"
	"go/token"
	"reflect"
)

// AdjustTypes Goes through all fields and looks at the json2go Tag, to determine if there's a better suiting type
// for the seen float and string values.
// Floats which could be represented as an int, are changed to int
// Strings which could be represented as UUID are change into uuid.UUID
// Strings which could be represented as time, are changed into time.Time
func AdjustTypes(file *ast.File, registeredTypeCheckers []TypeDeterminationFunction) error {
	var foundNodes []*AstUtils.FoundNodes
	var completed bool
	var requiredImports []string
	AstUtils.SearchNodes(file, &foundNodes, []*ast.Node{}, func(n *ast.Node, parents []*ast.Node, completed *bool) bool {
		if basicLit, ok := (*n).(*ast.BasicLit); ok && basicLit != nil && basicLit.Kind == token.STRING {
			return true
		}
		return false
	}, &completed)
	for _, node := range foundNodes {
		json2GoTag, err := GetJson2GoTagFromBasicLit((*node.Node).(*ast.BasicLit))
		if err != nil {
			return err
		}
		if json2GoTag != nil && json2GoTag.BaseType == nil {
			//Get base name
			var baseName string
			for _, parent := range node.Parents {
				if v, ok := (*parent).(*ast.Field); ok {
					baseName += AstUtils.SetExported(v.Names[0].Name)
				}
				if v, ok := (*parent).(*ast.TypeSpec); ok {
					baseName += AstUtils.SetExported(v.Name.Name)
					break
				}
			}

			//Get input type
			var originalType string
			var exp *ast.Expr
			for i, _ := range node.Parents {
				if field, ok := (*node.Parents[i]).(*ast.Field); ok {
					expr := walkExpressions(&field.Type)
					fmt.Println(reflect.TypeOf(*expr).String())
					switch e := (*expr).(type) {
					case *ast.SelectorExpr:
						originalType = e.Sel.Name + "." + e.X.(*ast.Ident).Name
						exp = expr
					case *ast.Ident:
						originalType = e.Name
						exp = expr
					case *ast.InterfaceType:
						originalType = "interface{}"
						exp = expr
					}
					break
				}
			}

			for _, checker := range registeredTypeCheckers {
				if checker.CouldTypeBeApplied(json2GoTag.SeenValues) {
					json2GoTag.ParseFunctions = &ParseFunctions{
						FromTypeParseFunction: "from" + baseName,
						ToTypeParseFunction:   "to" + baseName,
					}
					json2GoTag.BaseType = &originalType
					file.Decls = append(file.Decls, checker.GenerateFromTypeFunction(&ast.FuncDecl{
						Name: &ast.Ident{
							Name: json2GoTag.ParseFunctions.FromTypeParseFunction,
						},
						Type: &ast.FuncType{
							Params: &ast.FieldList{
								List: []*ast.Field{
									&ast.Field{
										Names: []*ast.Ident{
											&ast.Ident{
												Name: "baseValue",
											},
										},
										Type: &ast.Ident{
											Name: originalType,
										},
									},
								},
							},
							Results: &ast.FieldList{
								List: []*ast.Field{
									&ast.Field{
										Type: checker.GetType(),
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
							List: []ast.Stmt{},
						},
					}))
					file.Decls = append(file.Decls, checker.GenerateToTypeFunction(&ast.FuncDecl{
						Name: &ast.Ident{
							Name: json2GoTag.ParseFunctions.ToTypeParseFunction,
						},
						Type: &ast.FuncType{
							Params: &ast.FieldList{
								List: []*ast.Field{
									&ast.Field{
										Names: []*ast.Ident{
											&ast.Ident{
												Name: "baseValue",
											},
										},
										Type: checker.GetType(),
									},
								},
							},
							Results: &ast.FieldList{
								List: []*ast.Field{
									&ast.Field{
										Type: &ast.Ident{
											Name: originalType,
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
							List: []ast.Stmt{},
						},
					}))
					*exp = checker.GetType()
					(*node.Parents[0]).(*ast.Field).Tag, err = json2GoTag.AppendToTag((*node.Parents[0]).(*ast.Field).Tag)
					if err != nil {
						return err
					}
					requiredImports = append(requiredImports, checker.GetRequiredImports()...)
					break
				}
			}

		}
	}
	AstUtils.AddMissingImports(file, requiredImports)
	return nil
}

// TODO add error on unsupported Expr
func walkExpressions(expr *ast.Expr) *ast.Expr {
	switch e := (*expr).(type) {
	case *ast.Ident:
		return expr
	case *ast.StarExpr:
		return walkExpressions(&e.X)
	case *ast.ArrayType:
		return walkExpressions(&e.Elt)
	case *ast.InterfaceType:
		return expr
	case *ast.SelectorExpr:
		return expr
	}
	return nil
}
