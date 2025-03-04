package unmarshaller

import (
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"unicode"
)

type Generator struct {
	codeGenerator     codeGenerators.CodeGenerator
	fileData          fieldData.FileData
	added             bool
	structPrefixes    map[string]string
	globalsImportPath string
}

func NewGenerator(codeGenerator codeGenerators.CodeGenerator, fileData fieldData.FileData, globalImportsPath string) *Generator {
	return &Generator{
		codeGenerator:     codeGenerator,
		fileData:          fileData,
		globalsImportPath: globalImportsPath,
	}
}

func (g *Generator) GetGlobalFunctions() ([]ast.Decl, []string) {
	var decls []ast.Decl
	decls = append(decls, g.addAdditionalElementsError()...)
	decls = append(decls, g.addCheckForFirstErrorNotOfTypeTFunction())
	decls = append(decls, g.addGetAllErrorsOfTypeFunction())
	decls = append(decls, g.addRequiredFieldMissingError()...)
	return decls, g.getGlobalImports()
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

	if g.codeGenerator.IsStruct(path) && g.fileData[path].TypeAdjusterData != nil && g.fileData[path].ActiveType != nil {
		//TODO struct that is replaces as a whole. This case is not yet implemented
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
