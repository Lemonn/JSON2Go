package unmarshaller

import (
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"go/parser"
	"strings"
	"unicode"
)

type Generator struct {
	seenTypes      map[string]*fieldData.PathData
	added          bool
	structPrefixes map[string]string
	*utils.SeenTypeUtils
}

func NewGenerator(seenTypes map[string]*fieldData.PathData) *Generator {
	return &Generator{
		seenTypes:     seenTypes,
		SeenTypeUtils: utils.NewSeenTypeUtils(seenTypes),
	}
}

type FieldDetails struct {
	Path          string
	LevelOfArrays int
}

func (g *Generator) getBaseType(path string) (expr ast.Expr) {
	//TODO error on path not found

	if g.seenTypes[path].ForceSourceType != nil {
		parseExpr, err := parser.ParseExpr(*g.seenTypes[path].ForceSourceType)
		if err != nil {
			return nil
		}
		return parseExpr
	}

	if len(g.seenTypes[path].Types) == 1 {
		var Type fieldData.Type
		for Type, _ = range g.seenTypes[path].Types {
			break
		}
		if len(g.seenTypes[path].Types[Type]) == 1 {
			pathElements := strings.Split(path, ".")
			fmt.Println(Type)
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
			//TODO replace whit interface{}
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
		//TODO replace whit interface{}
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
	var err error
	var decls []ast.Decl
	var imports []string

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
								Name: string(unicode.ToLower([]rune(utils.GetParentFieldName(path))[0])),
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
