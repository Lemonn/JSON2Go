package JSON2Go

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/Lemonn/AstUtils"
	"go/ast"
	"go/token"
	"io"
)

// AdjustTypes Goes through all fields and looks at the json2go Tag, to determine if there's a better suiting type
// for the seen float and string values.
// Floats which could be represented as an int, are changed to int
// Strings which could be represented as UUID are change into uuid.UUID
// Strings which could be represented as time, are changed into time.Time
// TODO return required imports
func AdjustTypes(file *ast.File, registeredTypeCheckers []TypeDeterminationFunction) error {
	var foundNodes []*AstUtils.FoundNodes
	var completed bool
	AstUtils.SearchNodes(file, &foundNodes, []*ast.Node{}, func(n *ast.Node, parents []*ast.Node, completed *bool) bool {
		if basicLit, ok := (*n).(*ast.BasicLit); ok && basicLit != nil && basicLit.Kind == token.STRING {
			return true
		}
		return false
	}, &completed)
	for _, node := range foundNodes {
		if v := AstUtils.GetTagValue((*node.Node).(*ast.BasicLit).Value, "json2go"); v != "" {
			decoded := base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer([]byte(v)))
			all, err := io.ReadAll(decoded)
			if err != nil {
				return err
			}
			var json2GoTag Tag
			err = json.Unmarshal(all, &json2GoTag)
			if err != nil {
				return err
			}

			//Get base name
			var baseName string
			for _, parent := range node.Parents {
				if v, ok := (*parent).(*ast.Field); ok {
					baseName += AstUtils.SetExported(v.Names[0].Name)
				}
				if v, ok := (*parent).(*ast.TypeSpec); ok {
					baseName += AstUtils.SetExported(v.Name.Name)
				}
			}

			//Get input type
			var inputType string
			for _, parent := range node.Parents {
				if field, ok := (*parent).(*ast.Field); ok {
					if ident, ok := field.Type.(*ast.Ident); ok {
						inputType = ident.Name
					}
				}
			}
			if inputType == "" {
				//TODO handle error
			}
			for _, checker := range registeredTypeCheckers {
				if checker.CouldTypeBeApplied(json2GoTag.SeenValues) {
					json2GoTag.ParseFunctions = &ParseFunctions{
						FromTypeParseFunction: "from" + baseName,
						ToTypeParseFunction:   "to" + baseName,
					}
					json2GoTag.BaseType = &inputType
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
											Name: inputType,
										},
									},
								},
							},
							Results: &ast.FieldList{
								List: []*ast.Field{
									&ast.Field{
										Type: &ast.Ident{
											Name: "TODO",
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
										Type: &ast.Ident{
											Name: "TODO",
										},
									},
								},
							},
							Results: &ast.FieldList{
								List: []*ast.Field{
									&ast.Field{
										Type: &ast.Ident{
											Name: inputType,
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
					(*node.Parents[0]).(*ast.Field).Type.(*ast.Ident).Name = checker.GetType()
					(*node.Parents[0]).(*ast.Field).Tag, err = json2GoTag.AppendToTag((*node.Parents[0]).(*ast.Field).Tag)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
