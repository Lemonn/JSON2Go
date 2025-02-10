package marshaller

import (
	"errors"
	"fmt"
	"github.com/Lemonn/AstUtils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"reflect"
	"unicode"
)

type Generator struct {
	data map[string]*fieldData.FieldData
}

func NewGenerator(data map[string]*fieldData.FieldData) *Generator {
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
		var name string

		var path string
		for _, parent := range node.Parents {
			if v, ok := (*parent).(*ast.TypeSpec); ok {
				path += v.Name.Name
				name = v.Name.Name
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
			stmts, imports, err = g.structGenerator((*node.Node).(*ast.StructType), path, name)
			if err != nil {
				return err
			}

			AstUtils.AddMissingImports(file, imports)
		case *ast.Ident:
			stmts, imports, err = g.arrayGenerator(path, levelOfArrays, name)
			if err != nil {
				return err
			}
			AstUtils.AddMissingImports(file, imports)
		case *ast.SelectorExpr:
			stmts, imports, err = g.arrayGenerator(path, levelOfArrays, name)
			if err != nil {
				return err
			}
			AstUtils.AddMissingImports(file, imports)
		default:
			return errors.New(fmt.Sprintf("unkown type: %s", reflect.TypeOf(*node.Node).String()))
		}

		if stmts == nil || len(stmts) == 0 {
			continue
		}

		//Add Marshall function to file
		file.Decls = append(file.Decls, &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{
								Name: string(unicode.ToLower([]rune(name)[0])),
							},
						},
						Type: &ast.StarExpr{
							X: &ast.Ident{
								Name: name,
							},
						},
					},
				},
			},
			Name: &ast.Ident{
				Name: "MarshalJSON",
			},
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
				Results: &ast.FieldList{
					List: []*ast.Field{
						{
							Type: &ast.ArrayType{
								Elt: &ast.Ident{
									Name: "byte",
								},
							},
						},
						{
							Type: &ast.Ident{
								Name: "error",
							},
						},
					},
				},
			},
			Body: &ast.BlockStmt{List: stmts},
		})
	}
	return nil
}
