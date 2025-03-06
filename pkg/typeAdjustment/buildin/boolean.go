package buildin

import (
	"encoding/json"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"go/ast"
	"go/token"
	"maps"
)

type Boolean struct {
	state         *BooleanState
	fieldData     fieldData.FileData
	codeGenerator codeGenerators.CodeGenerator
}

type BooleanState struct {
	TrueStrings  map[string]struct{} `json:"trueValues,omitempty"`
	FalseStrings map[string]struct{} `json:"falseValues,omitempty"`
}

func NewBoolean(trueStrings map[string]struct{}, falseStrings map[string]struct{}) *Boolean {
	if trueStrings == nil || len(trueStrings) == 0 {
		trueStrings = map[string]struct{}{"true": {}}
	}
	if falseStrings == nil || len(falseStrings) == 0 {
		falseStrings = map[string]struct{}{"false": {}}
	}
	return &Boolean{
		state: &BooleanState{
			TrueStrings:  trueStrings,
			FalseStrings: falseStrings,
		},
	}
}

func (b *Boolean) CouldTypeBeApplied(path string) (typeAdjustment.State, error) {
	basicType, Level, Type := b.codeGenerator.IsBasicTypeWhitDetails(path)
	if !basicType {
		return typeAdjustment.StateFailed, nil
	}
	for s, _ := range b.fieldData[path].Types[Type][Level] {
		if _, ok := b.state.TrueStrings[s]; ok {
			return typeAdjustment.StateApplicable, nil
		} else if _, ok := b.state.FalseStrings[s]; ok {
			return typeAdjustment.StateApplicable, nil
		}
	}
	return typeAdjustment.StateFailed, nil
}

func (b *Boolean) GenerateMarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, []string, error) {
	if len(b.state.TrueStrings) > 1 || len(b.state.FalseStrings) > 1 {
		var trueMapTypesExpr []ast.Expr
		for v, _ := range b.state.TrueStrings {
			trueMapTypesExpr = append(trueMapTypesExpr, &ast.KeyValueExpr{
				Key: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"" + v + "\"",
				},
				Value: &ast.CompositeLit{},
			})
		}

		var falseMapTypesExpr []ast.Expr
		for v, _ := range b.state.FalseStrings {
			falseMapTypesExpr = append(falseMapTypesExpr, &ast.KeyValueExpr{
				Key: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"" + v + "\"",
				},
				Value: &ast.CompositeLit{},
			})
		}

		functionScaffold.Body.List = append(functionScaffold.Body.List, &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{
					Name: "trueStrings",
				},
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CompositeLit{
					Type: &ast.MapType{
						Key: &ast.Ident{
							Name: "string",
						},
						Value: &ast.StructType{
							Fields: &ast.FieldList{},
						},
					},
					Elts: trueMapTypesExpr,
				},
			},
		})
		functionScaffold.Body.List = append(functionScaffold.Body.List, &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{
					Name: "falseStrings",
				},
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.CompositeLit{
					Type: &ast.MapType{
						Key: &ast.Ident{
							Name: "string",
						},
						Value: &ast.StructType{
							Fields: &ast.FieldList{},
						},
					},
					Elts: falseMapTypesExpr,
				},
			},
		})
		functionScaffold.Body.List = append(functionScaffold.Body.List, &ast.IfStmt{
			Cond: &ast.Ident{
				Name: "baseValue",
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.RangeStmt{
						Key: &ast.Ident{
							Name: "str",
						},
						Value: &ast.Ident{
							Name: "_",
						},
						Tok: token.DEFINE,
						X: &ast.Ident{
							Name: "trueStrings",
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.Ident{
											Name: "str",
										},
										&ast.Ident{
											Name: "nil",
										},
									},
								},
							},
						},
					},
				},
			},
			Else: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.RangeStmt{
						Key: &ast.Ident{
							Name: "str",
						},
						Value: &ast.Ident{
							Name: "_",
						},
						Tok: token.DEFINE,
						X: &ast.Ident{
							Name: "falseStrings",
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.Ident{
											Name: "str",
										},
										&ast.Ident{
											Name: "nil",
										},
									},
								},
							},
						},
					},
				},
			},
		})
	} else {
		functionScaffold.Body.List = append(functionScaffold.Body.List, &ast.IfStmt{
			Cond: &ast.Ident{
				Name: "baseValue",
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.BasicLit{
								Kind: token.STRING,
								Value: "\"" + func() string {
									for s, _ := range b.state.TrueStrings {
										return s
									}
									return "true"
								}() + "\"",
							},
							&ast.Ident{
								Name: "nil",
							},
						},
					},
				},
			},
			Else: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.BasicLit{
								Kind: token.STRING,
								Value: "\"" + func() string {
									for s, _ := range b.state.FalseStrings {
										return s
									}
									return "false"
								}() + "\"",
							},
							&ast.Ident{
								Name: "nil",
							},
						},
					},
				},
			},
		})
	}
	return functionScaffold, []string{}, nil
}

func (b *Boolean) GenerateUnmarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, []string, error) {
	functionScaffold.Body.List = append(functionScaffold.Body.List, &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X: &ast.BinaryExpr{
				X: &ast.Ident{
					Name: "baseValue",
				},
				Op: token.NEQ,
				Y: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"true\"",
				},
			},
			Op: token.LAND,
			Y: &ast.BinaryExpr{
				X: &ast.Ident{
					Name: "baseValue",
				},
				Op: token.NEQ,
				Y: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"false\"",
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.Ident{
							Name: "nil",
						},
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "errors",
								},
								Sel: &ast.Ident{
									Name: "New",
								},
							},
							Args: []ast.Expr{
								&ast.BinaryExpr{
									X: &ast.BasicLit{
										Kind:  token.STRING,
										Value: "\"only true or false can be represented as bool, given was: \"",
									},
									Op: token.ADD,
									Y: &ast.Ident{
										Name: "baseValue",
									},
								},
							},
						},
					},
				},
			},
		},
		Else: &ast.IfStmt{
			Cond: &ast.BinaryExpr{
				X: &ast.Ident{
					Name: "baseValue",
				},
				Op: token.EQL,
				Y: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"true\"",
				},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.Ident{
								Name: "true",
							},
							&ast.Ident{
								Name: "nil",
							},
						},
					},
				},
			},
			Else: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.Ident{
								Name: "false",
							},
							&ast.Ident{
								Name: "nil",
							},
						},
					},
				},
			},
		},
	})
	return functionScaffold, []string{"errors"}, nil
}

func (b *Boolean) SetState(states []json.RawMessage, _ string, fileData fieldData.FileData, _ typeAdjustment.TypeDeterminationFunctions, codeGenerator codeGenerators.CodeGenerator) error {
	if b.state == nil {
		b.state = &BooleanState{
			TrueStrings:  make(map[string]struct{}),
			FalseStrings: make(map[string]struct{}),
		}
	}
	for _, state := range states {
		var lbs BooleanState
		err := json.Unmarshal(state, &lbs)
		if err != nil {
			return err
		}
		maps.Copy(b.state.FalseStrings, lbs.FalseStrings)
		maps.Copy(b.state.TrueStrings, lbs.TrueStrings)
	}
	b.fieldData = fileData
	b.codeGenerator = codeGenerator
	return nil
}

func (b *Boolean) GetState() (json.RawMessage, error) {
	return json.Marshal(b.state)
}

func (b *Boolean) GetExtraCode() ([]ast.Decl, []string) {
	return nil, nil
}

func (b *Boolean) TypeExpansion() bool {
	return false
}

func (b *Boolean) ForceSourceType() *string {
	return nil
}

func (b *Boolean) GetModFileContents() []*fieldData.ModFileContent {
	return nil
}

func (b *Boolean) GetVersion() *string {
	return utils.StringToPointer("v0.0.1")
}

func (b *Boolean) GetType() ast.Expr {
	var expr ast.Expr
	expr = &ast.Ident{Name: "bool"}
	//TODO should we generate the pointer type here, or should we make each type to an pointer type, if the field is
	// of pointer type?
	/*
		if b.codeGenerator.IsPointer(b.activePath) {
			expr = &ast.StarExpr{X: expr}
		}

	*/
	return expr
}

func (b *Boolean) GetName() string {
	return "json2Go.Boolean"
}
