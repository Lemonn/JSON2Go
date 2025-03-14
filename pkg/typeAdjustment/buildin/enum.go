package buildin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	"github.com/Lemonn/JSON2Go/pkg/errors/typeChecker"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"go/ast"
	"go/token"
	"maps"
)

type EnumTypeChecker struct {
	seenValues    []string
	currentPath   fieldData.Path
	state         *EnumTypeCheckerState
	settings      *EnumTypeCheckerSettings
	fileData      fieldData.FileData
	codeGenerator codeGenerators.CodeGenerator
}

type EnumTypeCheckerSettings struct {
	MinFieldCount   int
	MaxFieldCount   int
	MinTimesSeen    int
	SeenValuesRatio float64
}

type EnumTypeCheckerState struct {
	FieldOrder map[string]int           `json:"fieldOrder"`
	Settings   *EnumTypeCheckerSettings `json:"settings"`
}

func (s *EnumTypeCheckerState) combine(s1 *EnumTypeCheckerState) (*EnumTypeCheckerState, error) {

	var state EnumTypeCheckerState
	indexes := make(map[int]struct{})
	for fieldName, i := range s.FieldOrder {
		if v, ok := s1.FieldOrder[fieldName]; ok && v != i {
			return nil, &typeChecker.IncompatibleCustomTypeError{
				Err: errors.New(fmt.Sprintf("settings could not be combined, as the filed order is incompatible")),
			}
		}
		indexes[i] = struct{}{}
	}

	for _, i := range s.FieldOrder {
		if _, ok := indexes[i]; !ok {
			return nil, &typeChecker.IncompatibleCustomTypeError{
				Err: errors.New(fmt.Sprintf("settings could not be combined, as a field index occured more than once")),
			}
		}
	}

	state.FieldOrder = make(map[string]int)
	maps.Copy(state.FieldOrder, s.FieldOrder)
	maps.Copy(state.FieldOrder, s1.FieldOrder)

	//TODO combine settings
	return &state, nil
}

func NewEnumTypeChecker(settings *EnumTypeCheckerSettings) *EnumTypeChecker {
	return &EnumTypeChecker{
		settings: settings,
	}
}

func (e *EnumTypeChecker) getEnumName() string {
	return e.currentPath.GetFieldName() + "Enum"
}

func (e *EnumTypeChecker) GenerateMarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, fieldData.Imports, error) {
	functionScaffold.Body.List = append(functionScaffold.Body.List, &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X: &ast.Ident{
				Name: "baseValue",
			},
			Op: token.EQL,
			Y: &ast.Ident{
				Name: e.getEnumName() + "InvalidEnumValue",
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: "\"\"",
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
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "fmt",
										},
										Sel: &ast.Ident{
											Name: "Sprintf",
										},
									},
									Args: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.STRING,
											Value: "\"invalid enum value: %s\"",
										},
										&ast.Ident{
											Name: "baseValue",
										},
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
		},
	})
	//TODO add imports
	return functionScaffold, fieldData.Imports{{Path: "fmt"}}, nil
}

func (e *EnumTypeChecker) GenerateUnmarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, fieldData.Imports, error) {
	functionScaffold.Body.List = append(functionScaffold.Body.List, &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.Ident{
					Name: "New" + e.getEnumName(),
				},
				Args: []ast.Expr{
					&ast.Ident{
						Name: "baseValue",
					},
				},
			},
		},
	})
	//TODO add imports
	return functionScaffold, nil, nil
}

func (e *EnumTypeChecker) SetState(states []json.RawMessage, currentPath fieldData.Path, fileData fieldData.FileData, codeGenerator codeGenerators.CodeGenerator) error {
	e.currentPath = currentPath
	e.fileData = fileData
	e.codeGenerator = codeGenerator
	if states == nil || len(states) == 0 || states[0] == nil {
		e.state = &EnumTypeCheckerState{
			FieldOrder: make(map[string]int),
			Settings:   e.settings,
		}
		return nil
	}

	var state0, state1 *EnumTypeCheckerState
	err := json.Unmarshal(states[0], &state0)
	if err != nil {
		return err
	}
	for {
		if len(states) == 1 {
			break
		}
		err = json.Unmarshal(states[1], &state1)
		if err != nil {
			return err
		}
		states = states[2:]
		state0, err = state0.combine(state1)
		if err != nil {
			return err
		}
	}
	e.state = state0
	return nil

}

func (e *EnumTypeChecker) GetState() (json.RawMessage, error) {
	return json.Marshal(&e.state)
}

func (e *EnumTypeChecker) GetExtraCode() ([]ast.Decl, []string, error) {
	var decls []ast.Decl

	//Add enum type
	decls = append(decls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{
					Name: e.getEnumName(),
				},
				Type: &ast.Ident{
					Name: "int",
				},
			},
		},
	})

	//Add const's for enum type
	decls = append(decls, &ast.GenDecl{
		Tok: token.CONST,
		Specs: func() []ast.Spec {
			var specs []ast.Spec
			for i, value := range e.seenValues {
				if i == 0 {
					specs = append(specs, &ast.ValueSpec{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: e.getEnumName() + utils.JsonNameToGoName(value),
							},
						},
						Type: &ast.Ident{
							Name: e.getEnumName(),
						},
						Values: []ast.Expr{
							&ast.Ident{
								Name: "iota",
							},
						},
					})
				} else {
					specs = append(specs, &ast.ValueSpec{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: e.getEnumName() + utils.JsonNameToGoName(value),
							},
						},
					})
				}
			}
			return specs
		}(),
	})

	//Add new enum type function
	decls = append(decls, &ast.FuncDecl{
		Name: &ast.Ident{
			Name: "New" + e.getEnumName(),
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{
								Name: "baseValue",
							},
						},
						Type: &ast.Ident{
							Name: "string",
						},
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: &ast.Ident{
							Name: e.getEnumName(),
						},
					},
					&ast.Field{
						Type: &ast.Ident{
							Name: "error",
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.SwitchStmt{
					Tag: &ast.Ident{
						Name: "baseValue",
					},
					Body: &ast.BlockStmt{
						List: func() []ast.Stmt {
							var stmts []ast.Stmt
							for _, value := range e.seenValues {
								stmts = append(stmts, &ast.CaseClause{
									List: []ast.Expr{
										&ast.BasicLit{Value: "\"" + value + "\""},
									},
									Body: []ast.Stmt{
										&ast.ReturnStmt{
											Results: []ast.Expr{
												&ast.Ident{
													Name: e.getEnumName() + utils.JsonNameToGoName(value),
												},
												&ast.Ident{
													Name: "nil",
												},
											},
										},
									},
								})
							}
							stmts = append(stmts, &ast.CaseClause{
								Body: []ast.Stmt{
									&ast.ReturnStmt{
										Results: []ast.Expr{
											&ast.Ident{
												Name: e.getEnumName() + "InvalidEnumValue",
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
													&ast.CallExpr{
														Fun: &ast.SelectorExpr{
															X: &ast.Ident{
																Name: "fmt",
															},
															Sel: &ast.Ident{
																Name: "Sprintf",
															},
														},
														Args: []ast.Expr{
															&ast.BasicLit{
																Kind:  token.STRING,
																Value: "\"unsupported enum value: %s\"",
															},
															&ast.Ident{
																Name: "baseValue",
															},
														},
													},
												},
											},
										},
									},
								},
							})
							return stmts
						}(),
					},
				},
			},
		},
	})

	//Add to string method
	decls = append(decls, &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				&ast.Field{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: "e",
						},
					},
					Type: &ast.Ident{
						Name: e.getEnumName(),
					},
				},
			},
		},
		Name: &ast.Ident{
			Name: "String",
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Type: &ast.Ident{
							Name: "string",
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.SwitchStmt{
					Tag: &ast.Ident{
						Name: "e",
					},
					Body: &ast.BlockStmt{
						List: func() []ast.Stmt {
							var stmts []ast.Stmt
							for _, value := range e.seenValues {
								stmts = append(stmts, &ast.CaseClause{
									List: []ast.Expr{
										&ast.Ident{
											Name: e.getEnumName() + utils.JsonNameToGoName(value),
										},
									},
									Body: []ast.Stmt{
										&ast.ReturnStmt{
											Results: []ast.Expr{
												&ast.BasicLit{
													Kind:  token.STRING,
													Value: "\"" + value + "\"",
												},
											},
										},
									},
								})
							}
							stmts = append(stmts, &ast.CaseClause{
								Body: []ast.Stmt{
									&ast.ReturnStmt{
										Results: []ast.Expr{
											&ast.BasicLit{
												Kind:  token.STRING,
												Value: "\"InvalidEnumValue\"",
											},
										},
									},
								},
							})
							return stmts
						}(),
					},
				},
			},
		},
	})

	return decls, []string{"fmt"}, nil
}

func (e *EnumTypeChecker) ratio(levelOfArrays int) float64 {
	return float64(e.fileData[e.currentPath].SeenCounter) / float64(len(e.fileData[e.currentPath].Types[fieldData.String][levelOfArrays]))
}

func (e *EnumTypeChecker) CouldTypeBeApplied() (typeAdjustment.State, error) {

	isBasicType, levelOfArrays, Type := e.codeGenerator.IsBasicTypeWhitDetails(e.currentPath)
	if !isBasicType || Type != fieldData.String {
		return typeAdjustment.StateFailed, nil
	} else if e.fileData[e.currentPath].SeenCounter < e.settings.MinTimesSeen {
		return typeAdjustment.StateUndecided, nil
	} else if len(e.fileData[e.currentPath].Types[fieldData.String][levelOfArrays]) > e.settings.MaxFieldCount {
		return typeAdjustment.StateFailed, nil
	} else if e.ratio(levelOfArrays) < e.settings.SeenValuesRatio {
		return typeAdjustment.StateFailed, nil
	}

	if e.state == nil {
		e.state = &EnumTypeCheckerState{
			FieldOrder: make(map[string]int),
		}
	} else if e.state.FieldOrder == nil {
		e.state.FieldOrder = make(map[string]int)
	}

	if len(e.state.FieldOrder) == 0 {
		e.state.FieldOrder["InvalidEnumValue"] = 0
	}

	e.seenValues = []string{}
	for fieldValue, _ := range e.fileData[e.currentPath].Types[fieldData.String][levelOfArrays] {
		if _, ok := e.state.FieldOrder[fieldValue]; !ok {
			e.state.FieldOrder[fieldValue] = len(e.state.FieldOrder)
		}
	}
	e.seenValues = make([]string, len(e.state.FieldOrder))
	for s, i := range e.state.FieldOrder {
		e.seenValues[i] = s
	}
	return typeAdjustment.StateApplicable, nil
}

func (e *EnumTypeChecker) TypeExpansion() bool {
	//TODO replace whit implementation
	return false
}

func (e *EnumTypeChecker) ForceSourceType() *string {
	return nil
}

func (e *EnumTypeChecker) GetModFileContents() []*fieldData.ModFileContent {
	return nil
}

func (e *EnumTypeChecker) GetVersion() *string {
	return utils.StringToPointer("v0.0.1")
}

func (e *EnumTypeChecker) GetType() (ast.Expr, fieldData.Imports) {
	return &ast.Ident{Name: e.getEnumName()}, nil
}

func (e *EnumTypeChecker) GetName() string {
	return "json2go.EnumTypeChecker"
}

func (e *EnumTypeChecker) GetSubFiles() (map[fieldData.Path][]*fieldData.File, error) {
	return nil, nil
}

func (e *EnumTypeChecker) NeedsMarshaller() bool {
	return true
}

func (e *EnumTypeChecker) Clone() typeAdjustment.TypeDeterminationFunction {
	return NewEnumTypeChecker(&EnumTypeCheckerSettings{
		MinFieldCount:   e.settings.MinFieldCount,
		MaxFieldCount:   e.settings.MaxFieldCount,
		MinTimesSeen:    e.settings.MinTimesSeen,
		SeenValuesRatio: e.settings.SeenValuesRatio,
	})
}
