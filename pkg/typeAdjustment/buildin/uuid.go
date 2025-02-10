package buildin

import (
	"encoding/json"
	"github.com/google/uuid"
	"go/ast"
)

type UUIDTypeChecker struct{}

func (u *UUIDTypeChecker) SetState(state *json.RawMessage) error {
	//TODO implement me
	panic("implement me")
}

func (u *UUIDTypeChecker) GetState() (*json.RawMessage, error) {
	//TODO implement me
	panic("implement me")
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

func (u *UUIDTypeChecker) CouldTypeBeApplied(seenValues map[string]string) bool {
	var err error
	for value := range seenValues {
		_, err = uuid.Parse(value)
		if err != nil {
			return false
		}
	}
	return true
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
