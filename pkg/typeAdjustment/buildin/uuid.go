package buildin

import (
	"encoding/json"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"github.com/google/uuid"
	"go/ast"
)

type UUIDTypeChecker struct {
	fileData      fieldData.FileData
	codeGenerator codeGenerators.CodeGenerator
	version       string
	currentPath   fieldData.Path
}

func (u *UUIDTypeChecker) GetModFileContents() []*fieldData.ModFileContent {
	return nil
}

func (u *UUIDTypeChecker) GetVersion() *string {
	return &u.version
}

func (u *UUIDTypeChecker) ForceSourceType() *string {
	return nil
}

func (u *UUIDTypeChecker) GenerateMarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, fieldData.Imports, error) {
	functionScaffold.Body = &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "baseValue",
							},
							Sel: &ast.Ident{
								Name: "String",
							},
						},
					},
					&ast.Ident{
						Name: "nil",
					},
				},
			},
		},
	}
	return functionScaffold, fieldData.Imports{{Path: "github.com/google/uuid"}}, nil
}

func (u *UUIDTypeChecker) GenerateUnmarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, fieldData.Imports, error) {
	functionScaffold.Body = &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "uuid",
							},
							Sel: &ast.Ident{
								Name: "Parse",
							},
						},
						Args: []ast.Expr{
							&ast.Ident{
								Name: "baseValue",
							},
						},
					},
				},
			},
		},
	}
	return functionScaffold, fieldData.Imports{{Path: "github.com/google/uuid"}}, nil
}

func (u *UUIDTypeChecker) TypeExpansion() bool {
	//TODO implement me
	return false
}

func (u *UUIDTypeChecker) CouldTypeBeApplied() (typeAdjustment.State, error) {
	var Level int
	var err error
	pathData := u.fileData[u.currentPath]
	//TODO check if its struct type and ignore
	if len(pathData.Types) > 1 {
		return typeAdjustment.StateFailed, nil
	}
	Type := u.codeGenerator.GetType(u.currentPath)
	if len(pathData.Types[Type]) > 1 {
		return typeAdjustment.StateFailed, nil
	}
	for Level = range pathData.Types[Type] {
		break
	}

	for value := range pathData.Types[Type][Level] {
		_, err = uuid.Parse(value)
		if err != nil {
			return typeAdjustment.StateFailed, nil
		}
	}
	return typeAdjustment.StateApplicable, nil
}

func (u *UUIDTypeChecker) GetExtraCode() ([]ast.Decl, []string, error) {
	return nil, nil, nil
}

func (u *UUIDTypeChecker) SetState(_ []json.RawMessage, path fieldData.Path, fileData fieldData.FileData, codeGenerator codeGenerators.CodeGenerator) error {
	u.fileData = fileData
	u.codeGenerator = codeGenerator
	u.currentPath = path
	return nil
}

func (u *UUIDTypeChecker) GetState() (json.RawMessage, error) {
	return nil, nil
}

func (u *UUIDTypeChecker) GetType() (ast.Expr, fieldData.Imports) {
	return &ast.SelectorExpr{
		X: &ast.Ident{
			Name: "uuid",
		},
		Sel: &ast.Ident{
			Name: "UUID",
		},
	}, fieldData.Imports{{Path: "github.com/google/uuid"}}
}

func (u *UUIDTypeChecker) GetName() string {
	return "json2go.UUIDTypeChecker"
}

func (u *UUIDTypeChecker) GetSubFiles() (map[fieldData.Path][]*fieldData.File, error) {
	return nil, nil
}

func (u *UUIDTypeChecker) NeedsMarshaller() bool {
	return true
}

func (u *UUIDTypeChecker) Clone() typeAdjustment.TypeDeterminationFunction {
	return &UUIDTypeChecker{}
}
