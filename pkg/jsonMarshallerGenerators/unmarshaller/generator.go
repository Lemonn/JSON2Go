package unmarshaller

import (
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"strings"
	"unicode"
)

type Generator struct {
	seenTypes      map[string]*fieldData.PathData
	added          bool
	structPrefixes map[string]string
}

func NewGenerator(seenTypes map[string]*fieldData.PathData) *Generator {
	return &Generator{
		seenTypes: seenTypes,
	}
}

type FieldDetails struct {
	Path          string
	LevelOfArrays int
}

func (g *Generator) getBaseType(path string) (expr ast.Expr) {
	//TODO error on path not found
	if len(g.seenTypes[path].Types) == 1 {
		var Type fieldData.Type
		for Type, _ = range g.seenTypes[path].Types {
			break
		}
		if len(g.seenTypes[path].Types[Type]) == 1 {
			pathElements := strings.Split(path, ".")
			if Type == fieldData.Field {
				if v, ok := g.structPrefixes[pathElements[len(pathElements)-1]]; ok {
					expr = &ast.SelectorExpr{
						X:   &ast.StarExpr{X: &ast.Ident{Name: v}},
						Sel: &ast.Ident{Name: pathElements[len(pathElements)-1]},
					}
				} else {
					//TODO do not know if this is correct
					expr = &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: pathElements[len(pathElements)-1]}, Sel: &ast.Ident{Name: pathElements[len(pathElements)-1]}}}

				}
			} else if Type == fieldData.EmptyArray {
				expr = &ast.InterfaceType{Methods: &ast.FieldList{}}
			} else if Type == fieldData.EmptyStruct {
				expr = &ast.InterfaceType{Methods: &ast.FieldList{}}
			} else {
				expr = &ast.Ident{Name: string(Type)}
			}
		} else {
			expr = &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "json",
				},
				Sel: &ast.Ident{
					Name: "RawMessage",
				},
			}
		}
	} else {
		expr = &ast.SelectorExpr{
			X: &ast.Ident{
				Name: "json",
			},
			Sel: &ast.Ident{
				Name: "RawMessage",
			},
		}
	}

	return expr
}

func (g *Generator) isStruct(path string) bool {
	if len(g.seenTypes[path].Types) == 1 {
		var Type fieldData.Type
		for Type, _ = range g.seenTypes[path].Types {
			break
		}
		if len(g.seenTypes[path].Types[Type]) == 1 {
			if Type == fieldData.Field {
				return true
			} else {
				return false
			}
		} else {
			return false
		}
	} else {
		return false
	}
}

func (g *Generator) Generate(path string) ([]ast.Decl, []string, error) {
	//rawJsonMsg := false
	var err error
	var decls []ast.Decl
	var imports []string

	if g.seenTypes[path].TypeAdjusterData != nil {
		//TODO only set to rawMsg if the field is of struct type or mixed type, not if it is of base type such as int, string, etc.
		//rawJsonMsg = true
	}

	var stmts []ast.Stmt

	//TODO basic field vs not basic field, is for the fact, that a not basic field could return an additional elements error
	if g.isStruct(path) && g.seenTypes[path].TypeAdjusterData != nil && g.seenTypes[path].TypeAdjusterData.ActiveType != nil {
		//TODO struct that is replaces as a whole. This case is not yet implemented
	} else if g.isStruct(path) {
		stmts, imports, err = g.structGenerator(path)
		if err != nil {
			return nil, nil, err
		}
	} else {
		stmts, imports, err = g.arrayGenerator(path)
		if err != nil {
			return nil, nil, err
		}
	}
	if len(stmts) != 0 {
		decls = append(decls, &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{
								Name: string(unicode.ToLower([]rune(g.getFieldName(path))[0])),
							},
						},
						Type: g.getFieldType(path, 1),
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
		})
	}

	return decls, imports, nil
}

/*
func (g *Generator) Generate() error {
	var foundNodes []*AstUtils.FoundNodes
	var completed bool
	AstUtils.SearchNodes(g.inputFile, &foundNodes, []*ast.Node{}, func(n *ast.Node, parents []*ast.Node, completed *bool) bool {
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
			continue
		}

		switch (*node.Node).(type) {
		case *ast.StructType:
			stmts, imports, err = g.structGenerator((*node.Node).(*ast.StructType), path, path)
			if err != nil {
				return err
			}

			//TODO move back to struct generator
			if !g.added {
				g.added = true
				// Add AdditionalElementError + support methods
				g.outputFile.Decls = append(g.outputFile.Decls, &ast.GenDecl{
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
				g.outputFile.Decls = append(g.outputFile.Decls, &ast.FuncDecl{
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
				g.outputFile.Decls = append(g.outputFile.Decls, &ast.FuncDecl{
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
				addGetAllErrorsOfTypeFunction(g.outputFile)
				addCheckForFirstErrorNotOfTypeTFunction(g.outputFile)
			}

			AstUtils.AddMissingImports(g.outputFile, imports)
		case *ast.Ident:
			stmts, imports = g.arrayGenerator(path, levelOfArrays, (*node.Node).(*ast.Ident), path)
			AstUtils.AddMissingImports(g.outputFile, imports)
		case *ast.SelectorExpr:
			stmts, imports = g.arrayGenerator(path, levelOfArrays, (*node.Node).(*ast.SelectorExpr), path)
			AstUtils.AddMissingImports(g.outputFile, imports)
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
		g.outputFile.Decls = append(g.outputFile.Decls, f1)
	}
	AstUtils.AddMissingImports(g.outputFile, []string{"encoding/json"})
	return nil
}
*/
