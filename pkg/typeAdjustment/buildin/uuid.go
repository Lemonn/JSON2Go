package buildin

import (
	"encoding/json"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"github.com/google/uuid"
	"go/ast"
)

type UUIDTypeChecker struct{}

func (u *UUIDTypeChecker) GenerateMarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error) {
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
	return functionScaffold, nil
}

func (u *UUIDTypeChecker) GenerateUnmarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error) {
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
	return functionScaffold, nil
}

func (u *UUIDTypeChecker) TypeExpansion() bool {
	//TODO implement me
	panic("implement me")
}

func (u *UUIDTypeChecker) CouldTypeBeApplied(seenTypes map[fieldData.Type]map[int]map[string]*fieldData.ValueDetails) (typeAdjustment.State, error) {
	var Type fieldData.Type
	var Level int
	var err error
	//TODO check if its struct type and ignore
	if len(seenTypes) > 1 {
		return typeAdjustment.StateFailed, nil
	}
	for Type = range seenTypes {
		break
	}
	if len(seenTypes[Type]) > 1 {
		return typeAdjustment.StateFailed, nil
	}
	for Level = range seenTypes[Type] {
		break
	}

	for value := range seenTypes[Type][Level] {
		_, err = uuid.Parse(value)
		if err != nil {
			return typeAdjustment.StateFailed, nil
		}
	}
	return typeAdjustment.StateApplicable, nil
}

func (u *UUIDTypeChecker) GetExtraCode() []ast.Decl {
	return nil
}

func (u *UUIDTypeChecker) SetState(state []*json.RawMessage, currentPath string) error {
	return nil
}

func (u *UUIDTypeChecker) GetState() ([]*json.RawMessage, error) {
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

func (u *UUIDTypeChecker) GetRequiredImports() []string {
	return []string{"github.com/google/uuid"}
}

func (u *UUIDTypeChecker) GetName() string {
	return "json2go.UUIDTypeChecker"
}
