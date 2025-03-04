package marshaller

import (
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"unicode"
)

type Generator struct {
	fileData      fieldData.FileData
	codeGenerator codeGenerators.CodeGenerator
}

func NewGenerator(fileData fieldData.FileData, codeGenerator codeGenerators.CodeGenerator) *Generator {
	return &Generator{
		fileData:      fileData,
		codeGenerator: codeGenerator,
	}
}

func (g *Generator) GetGlobalFunctions() []ast.Decl {
	var decls []ast.Decl
	return decls
}

func (g *Generator) Generate(path string) ([]ast.Decl, []string, error) {
	var err error
	var decls []ast.Decl
	var imports []string
	var stmts []ast.Stmt

	if g.fileData[path].ForceSourceType != nil {
		if !g.fileData[path].DirectToForceSourceType {
			//TODO struct that is replaces as a whole. This case is not yet implemented
		}
	} else if g.codeGenerator.IsStruct(path) {
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
						Type: &ast.StarExpr{
							X: &ast.Ident{
								Name: utils.GetFieldName(path),
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
