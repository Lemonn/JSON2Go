package buildin

import (
	"encoding/json"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"github.com/google/uuid"
	"go/ast"
)

type UUIDTypeChecker struct{}

func (u *UUIDTypeChecker) SetState(state json.RawMessage, currentPath string) error {
	return nil
}

func (u *UUIDTypeChecker) GetState() (json.RawMessage, error) {
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

func (u *UUIDTypeChecker) CouldTypeBeApplied(seenValues map[string]*fieldData.ValueData) typeAdjustment.State {
	var err error
	for value := range seenValues {
		_, err = uuid.Parse(value)
		if err != nil {
			return typeAdjustment.StateFailed
		}
	}
	return typeAdjustment.StateApplicable
}

func (u *UUIDTypeChecker) GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error) {
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

func (u *UUIDTypeChecker) GenerateToTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error) {
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

func (u *UUIDTypeChecker) GetRequiredImports() []string {
	return []string{"github.com/google/uuid"}
}

func (u *UUIDTypeChecker) SetFile(_ *ast.File) {}

func (u *UUIDTypeChecker) GetName() string {
	return "json2go.UUIDTypeChecker"
}
