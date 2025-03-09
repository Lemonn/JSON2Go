package buildin

import (
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
)

type Generator struct {
	fileData      fieldData.FileData
	codeGenerator codeGenerators.CodeGenerator
	alias         string
}

func NewGenerator(fileData fieldData.FileData, generator codeGenerators.CodeGenerator) *Generator {
	return &Generator{
		fileData:      fileData,
		codeGenerator: generator,
		alias:         "globalsTest",
	}
}

func (g *Generator) Unmarshall(path fieldData.Path) ([]*fieldData.File, error) {
	var err error
	var decls []ast.Decl
	var imports fieldData.Imports
	var stmts []ast.Stmt

	//TODO g.fileData[path].ActiveType != nil should this not be forced type?
	if g.codeGenerator.IsStruct(path) && g.fileData[path].TypeAdjusterData != nil && g.fileData[path].ActiveType != nil {
		//TODO struct that is replaces as a whole. This case is not yet implemented
	} else if g.codeGenerator.IsStruct(path) {
		stmts, imports, err = g.unmarshallStructGenerator(path)
		if err != nil {
			return nil, err
		}
	} else {
		stmts, imports, err = g.unmarshallArrayGenerator(path)
		if err != nil {
			return nil, err
		}
	}
	if len(stmts) != 0 {
		decls = append(decls, &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{
								Name: path.GetRune(),
							},
						},
						Type: &ast.StarExpr{X: &ast.Ident{Name: path.GetFieldName()}},
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
	if len(decls) == 0 {
		return nil, nil
	}
	return []*fieldData.File{fieldData.GetGoFile(decls, imports, nil, fieldData.FileClassUnMarshaller)}, nil
}

func (g *Generator) Marshall(path fieldData.Path) ([]*fieldData.File, error) {
	var err error
	var decls []ast.Decl
	var imports fieldData.Imports
	var stmts []ast.Stmt

	if g.fileData[path].ForceSourceType != nil {
		if !g.fileData[path].DirectToForceSourceType {
			//TODO struct that is replaces as a whole. This case is not yet implemented
		}
	} else if g.codeGenerator.IsStruct(path) {
		stmts, imports, err = g.marshallStructGenerator(path)
		if err != nil {
			return nil, err
		}
	} else {
		stmts, imports, err = g.marshallArrayGenerator(path)
		if err != nil {
			return nil, err
		}
	}
	if len(stmts) != 0 {
		decls = append(decls, &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{
							{
								Name: path.GetRune(),
							},
						},
						Type: &ast.StarExpr{
							X: &ast.Ident{
								Name: path.GetFieldName(),
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
	if len(decls) == 0 {
		return nil, nil
	}
	return []*fieldData.File{fieldData.GetGoFile(decls, imports, nil, fieldData.FileClassMarshaller)}, nil
}

func (g *Generator) GlobalFiles() []*fieldData.File {
	var decls []ast.Decl
	decls = append(decls, g.addCheckForFirstErrorNotOfTypeTFunction())
	decls = append(decls, g.addGetAllErrorsOfTypeFunction())
	decls = append(decls, g.addAdditionalElementsError()...)
	decls = append(decls, g.addRequiredFieldMissingError()...)
	return []*fieldData.File{fieldData.GetGoFile(decls, g.getGlobalImports(), nil, fieldData.FileClassGlobal)}
}
