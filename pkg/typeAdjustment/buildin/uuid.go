package buildin

import (
	"encoding/json"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"github.com/google/uuid"
	"go/ast"
)

type UUIDTypeChecker struct {
	utils     *utils.SeenTypeUtils
	seenTypes map[string]*fieldData.PathData
}

func (u *UUIDTypeChecker) GetModFileContents() []*fieldData.ModFileContent {
	return nil
}

func (u *UUIDTypeChecker) GetVersion() string {
	return "v0.0.1"
}

func (u *UUIDTypeChecker) ForceSourceType() *string {
	return nil
}

func (u *UUIDTypeChecker) GenerateMarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, []string, error) {
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
	return functionScaffold, []string{"github.com/google/uuid"}, nil
}

func (u *UUIDTypeChecker) GenerateUnmarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, []string, error) {
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
	return functionScaffold, []string{"github.com/google/uuid"}, nil
}

func (u *UUIDTypeChecker) TypeExpansion() bool {
	//TODO implement me
	return false
}

func (u *UUIDTypeChecker) CouldTypeBeApplied(path string) (typeAdjustment.State, error) {
	var Level int
	var err error
	//TODO check if its struct type and ignore
	if len(u.seenTypes[path].Types) > 1 {
		return typeAdjustment.StateFailed, nil
	}
	Type := u.utils.GetType(path)
	if len(u.seenTypes[path].Types[Type]) > 1 {
		return typeAdjustment.StateFailed, nil
	}
	for Level = range u.seenTypes[path].Types[Type] {
		break
	}

	for value := range u.seenTypes[path].Types[Type][Level] {
		_, err = uuid.Parse(value)
		if err != nil {
			return typeAdjustment.StateFailed, nil
		}
	}
	return typeAdjustment.StateApplicable, nil
}

func (u *UUIDTypeChecker) GetExtraCode() ([]ast.Decl, []string) {
	return nil, nil
}

func (u *UUIDTypeChecker) SetState(_ []json.RawMessage, _ string, _ []typeAdjustment.TypeDeterminationFunction, seenTypes map[string]*fieldData.PathData) error {
	u.seenTypes = seenTypes
	u.utils = utils.NewSeenTypeUtils(seenTypes)
	return nil
}

func (u *UUIDTypeChecker) GetState() ([]json.RawMessage, error) {
	return nil, nil
}

func (u *UUIDTypeChecker) GetType() ast.Expr {
	return &ast.SelectorExpr{
		X: &ast.Ident{
			Name: "uuid",
		},
		Sel: &ast.Ident{
			Name: "UUID",
		},
	}
}

func (u *UUIDTypeChecker) GetName() string {
	return "json2go.UUIDTypeChecker"
}
