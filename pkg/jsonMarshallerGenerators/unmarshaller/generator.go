package unmarshaller

import (
	"errors"
	"fmt"
	"github.com/Lemonn/AstUtils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"reflect"
)

type Generator struct {
	data map[string]*fieldData.Data
}

func NewGenerator(data map[string]*fieldData.Data) *Generator {
	return &Generator{data: data}
}

func (g *Generator) Generate(file *ast.File) error {
	var foundNodes []*AstUtils.FoundNodes
	var completed bool
	AstUtils.SearchNodes(file, &foundNodes, []*ast.Node{}, func(n *ast.Node, parents []*ast.Node, completed *bool) bool {
		if _, ok := (*n).(*ast.StructType); ok && len(parents) > 0 {
			return true
		} else if _, ok := (*n).(*ast.Ident); ok && len(parents) > 0 {
			if _, ok := (*parents[0]).(*ast.ArrayType); ok {
				return true
			}
		} else if _, ok := (*n).(*ast.SelectorExpr); ok && len(parents) > 0 {
			if _, ok := (*parents[0]).(*ast.ArrayType); ok {
				return true
			}
		}
		return false
	}, &completed)
	for _, node := range foundNodes {
		var stmts []ast.Stmt
		var levelOfArrays int
		var nested bool
		var err error
		var imports []string

		var path string
		for _, parent := range node.Parents {
			if v, ok := (*parent).(*ast.TypeSpec); ok {
				path += v.Name.Name
			} else if _, ok := (*parent).(*ast.StructType); ok {
				nested = true
				//Ignore nested structs
				break
			} else if _, ok := (*parent).(*ast.FuncDecl); ok {
				nested = true
				//Ignore structs inside functions
				break
			} else if _, ok := (*parent).(*ast.ArrayType); ok {
				levelOfArrays++
			}
		}
		if nested {
			continue
		}

		switch (*node.Node).(type) {
		case *ast.StructType:
			fmt.Println("StructType")
			stmts, imports, err = g.structGenerator((*node.Node).(*ast.StructType), path)
			if err != nil {
				return err
			}
			AstUtils.AddMissingImports(file, imports)
		case *ast.Ident:
			fmt.Println("Ident")
			stmts, imports = g.arrayGenerator(path, levelOfArrays, (*node.Node).(*ast.Ident).Name)
			AstUtils.AddMissingImports(file, imports)
		case *ast.SelectorExpr:
			fmt.Println("SelectorExpr")
			stmts, imports = g.arrayGenerator(path, levelOfArrays, (*node.Node).(*ast.SelectorExpr).X.(*ast.Ident).Name+"."+(*node.Node).(*ast.SelectorExpr).Sel.Name)
			AstUtils.AddMissingImports(file, imports)
		default:
			return errors.New(fmt.Sprintf("unkown type: %s", reflect.TypeOf(*node.Node).String()))
		}
		f1 := &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "R",
							},
						},
						Type: &ast.StarExpr{
							X: &ast.Ident{
								Name: "RRR",
							},
						},
					},
				},
			},
			Name: &ast.Ident{
				Name: "UnmarshalJSON",
			},
			Type: &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{
						&ast.Field{
							Names: []*ast.Ident{
								&ast.Ident{
									Name: "bytes",
								},
							},
							Type: &ast.ArrayType{
								Elt: &ast.Ident{
									Name: "byte",
								},
							},
						},
					},
				},
				Results: &ast.FieldList{
					List: []*ast.Field{
						&ast.Field{
							Type: &ast.Ident{
								Name: "error",
							},
						},
					},
				},
			},
			Body: &ast.BlockStmt{List: stmts},
		}
		file.Decls = append(file.Decls, f1)
	}
	AstUtils.AddMissingImports(file, []string{"encoding/json"})
	return nil
}
