package marshaller

import (
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/jsonMarshallerGenerators"
	"go/ast"
	"unicode"
)

type Generator struct {
	seenTypes map[string]*fieldData.PathData
	*jsonMarshallerGenerators.WrappingJSONMarshaller
}

func NewGenerator(seenTypes map[string]*fieldData.PathData) *Generator {
	return &Generator{
		seenTypes:              seenTypes,
		WrappingJSONMarshaller: jsonMarshallerGenerators.NewWrappingJSONMarshaller(seenTypes),
	}
}

func (g *Generator) Generate(path string) ([]ast.Decl, []string, error) {
	var err error
	var decls []ast.Decl
	var imports []string
	var stmts []ast.Stmt

	if g.seenTypes[path].ForceSourceType != nil {
		if !g.seenTypes[path].DirectToForceSourceType {
			//TODO struct that is replaces as a whole. This case is not yet implemented
		}
	} else if g.IsStruct(path) {
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
								Name: string(unicode.ToLower([]rune(g.GetFieldName(path))[0])),
							},
						},
						Type: &ast.StarExpr{
							X: &ast.Ident{
								Name: g.GetFieldName(path),
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

	return decls, imports, nil
}
