package unmarshaller

import (
	"errors"
	"fmt"
	"github.com/Lemonn/AstUtils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"go/token"
	"reflect"
	"unicode"
)

type Generator struct {
	data  map[string]*fieldData.FieldData
	added bool
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
		if _, ok := g.data[path]; !ok || nested {
			fmt.Println(path + "dsfasdfsdaf")
			continue
		}

		switch (*node.Node).(type) {
		case *ast.StructType:
			fmt.Println("StructType")
			stmts, imports, err = g.structGenerator((*node.Node).(*ast.StructType), path, path)
			if err != nil {
				return err
			}

			//TODO move back to struct generator
			if !g.added {
				g.added = true
				// Add AdditionalElementError + support methods
				file.Decls = append(file.Decls, &ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: "AdditionalElementsError",
							},
							Type: &ast.StructType{
								Fields: &ast.FieldList{
									List: []*ast.Field{
										{
											Names: []*ast.Ident{
												{
													Name: "ParsedObj",
												},
											},
											Type: &ast.Ident{
												Name: "string",
											},
										},
										{
											Names: []*ast.Ident{
												{
													Name: "Elements",
												},
											},
											Type: &ast.MapType{
												Key: &ast.Ident{
													Name: "string",
												},
												Value: &ast.SelectorExpr{
													X: &ast.Ident{
														Name: "json",
													},
													Sel: &ast.Ident{
														Name: "RawMessage",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				})
				file.Decls = append(file.Decls, &ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{
										Name: "j",
									},
								},
								Type: &ast.StarExpr{
									X: &ast.Ident{
										Name: "AdditionalElementsError",
									},
								},
							},
						},
					},
					Name: &ast.Ident{
						Name: "String",
					},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.Ident{
										Name: "string",
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X: &ast.Ident{
												Name: "j",
											},
											Sel: &ast.Ident{
												Name: "Error",
											},
										},
									},
								},
							},
						},
					},
				})
				file.Decls = append(file.Decls, &ast.FuncDecl{
					Recv: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{
										Name: "j",
									},
								},
								Type: &ast.StarExpr{
									X: &ast.Ident{
										Name: "AdditionalElementsError",
									},
								},
							},
						},
					},
					Name: &ast.Ident{
						Name: "Error",
					},
					Type: &ast.FuncType{
						Params: &ast.FieldList{},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: &ast.Ident{
										Name: "string",
									},
								},
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.Ident{
										Name: "m",
									},
								},
								Tok: token.DEFINE,
								Rhs: []ast.Expr{
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: "\"the following unexpected additional elements were found: \"",
									},
								},
							},
							&ast.RangeStmt{
								Key: &ast.Ident{
									Name: "s",
								},
								Value: &ast.Ident{
									Name: "e",
								},
								Tok: token.DEFINE,
								X: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "j",
									},
									Sel: &ast.Ident{
										Name: "Elements",
									},
								},
								Body: &ast.BlockStmt{
									List: []ast.Stmt{
										&ast.AssignStmt{
											Lhs: []ast.Expr{
												&ast.Ident{
													Name: "m",
												},
											},
											Tok: token.ADD_ASSIGN,
											Rhs: []ast.Expr{
												&ast.CallExpr{
													Fun: &ast.SelectorExpr{
														X: &ast.Ident{
															Name: "fmt",
														},
														Sel: &ast.Ident{
															Name: "Sprintf",
														},
													},
													Args: []ast.Expr{
														&ast.BasicLit{
															Kind:  token.STRING,
															Value: "\"[(%s) RawJsonString {\\\"%s\\\": %s}]\"",
														},
														&ast.Ident{
															Name: "s",
														},
														&ast.Ident{
															Name: "s",
														},
														&ast.Ident{
															Name: "e",
														},
													},
												},
											},
										},
									},
								},
							},
							&ast.AssignStmt{
								Lhs: []ast.Expr{
									&ast.Ident{
										Name: "m",
									},
								},
								Tok: token.ADD_ASSIGN,
								Rhs: []ast.Expr{
									&ast.BinaryExpr{
										X: &ast.BasicLit{
											Kind:  token.STRING,
											Value: "\" whilst parsing \"",
										},
										Op: token.ADD,
										Y: &ast.SelectorExpr{
											X: &ast.Ident{
												Name: "j",
											},
											Sel: &ast.Ident{
												Name: "ParsedObj",
											},
										},
									},
								},
							},
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.Ident{
										Name: "m",
									},
								},
							},
						},
					},
				})
			}

			AstUtils.AddMissingImports(file, imports)
		case *ast.Ident:
			fmt.Println("Ident")
			stmts, imports = g.arrayGenerator(path, levelOfArrays, (*node.Node).(*ast.Ident).Name, path)
			AstUtils.AddMissingImports(file, imports)
		case *ast.SelectorExpr:
			fmt.Println("SelectorExpr")
			stmts, imports = g.arrayGenerator(path, levelOfArrays, (*node.Node).(*ast.SelectorExpr).X.(*ast.Ident).Name+"."+(*node.Node).(*ast.SelectorExpr).Sel.Name, path)
			AstUtils.AddMissingImports(file, imports)
		default:
			return errors.New(fmt.Sprintf("unkown type: %s", reflect.TypeOf(*node.Node).String()))
		}
		if len(stmts) == 0 {
			continue
		}
		f1 := &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{
								Name: string(unicode.ToLower([]rune(path)[0])),
							},
						},
						Type: &ast.StarExpr{
							X: &ast.Ident{
								Name: path,
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
						{
							Names: []*ast.Ident{
								{
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
						{
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
