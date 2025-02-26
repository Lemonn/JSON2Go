package unmarshaller

import (
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
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

func (g *Generator) Generate(path string) ([]ast.Decl, []string, error) {
	var err error
	var decls []ast.Decl
	var imports []string
	var stmts []ast.Stmt

	if g.IsStruct(path) && g.seenTypes[path].TypeAdjusterData != nil && g.seenTypes[path].TypeAdjusterData.ActiveType != nil {
		//TODO struct that is replaces as a whole. This case is not yet implemented
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
								Name: string(unicode.ToLower([]rune(utils.GetFieldName(path))[0])),
							},
						},
						Type: &ast.StarExpr{X: &ast.Ident{Name: utils.GetFieldName(path)}},
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
