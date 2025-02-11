package buildin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"github.com/iancoleman/strcase"
	"go/ast"
	"go/token"
	"strings"
)

type EnumTypeChecker struct {
	totalFilesSeen int
	minFilesSeen   int
	maxFieldCount  int
	minFieldCount  int
	seenValues     []string
	fieldType      string
	file           *ast.File
	enumType       string
	decls          []ast.Decl
	currentPath    string
	state          *EnumTypeCheckerState
}

func NewEnumTypeChecker(minFilesSeen int, minFieldCount int, maxFieldCount int) *EnumTypeChecker {
	return &EnumTypeChecker{
		minFilesSeen:   minFilesSeen,
		maxFieldCount:  maxFieldCount,
		totalFilesSeen: 100,
		minFieldCount:  2,
	}
}

type EnumTypeCheckerState struct {
	FieldOrder map[string]int
}

func (e *EnumTypeChecker) CouldTypeBeApplied(seenValues map[string]string) typeAdjustment.State {
	if e.totalFilesSeen < e.minFilesSeen {
		return typeAdjustment.StateUndecided
	} else if len(seenValues) > e.maxFieldCount || len(seenValues) < e.minFieldCount {
		return typeAdjustment.StateFailed
	}
	var fieldType string
	for _, Type := range seenValues {
		if Type != "string" {
			return typeAdjustment.StateFailed
		} else if fieldType == "" {
			fieldType = Type
		}
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
	e.fieldType = fieldType
	if e.state != nil && e.state.FieldOrder != nil {
		for fieldValue, _ := range seenValues {
			if _, ok := e.state.FieldOrder[fieldValue]; !ok {
				e.state.FieldOrder[fieldValue] = len(e.state.FieldOrder)
			}
		}
	}
	e.seenValues = make([]string, len(e.state.FieldOrder))
	for s, i := range e.state.FieldOrder {
		e.seenValues[i] = s
	}
	return typeAdjustment.StateApplicable
}

func (e *EnumTypeChecker) GetType() ast.Expr {
	return &ast.Ident{Name: "Enum" + strings.ReplaceAll(e.currentPath, ".", "")}
}

func (e *EnumTypeChecker) GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error) {
	functionScaffold.Body.List = append(functionScaffold.Body.List, &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.Ident{
					Name: "NewEnum" + strings.ReplaceAll(e.currentPath, ".", ""),
				},
				Args: []ast.Expr{
					&ast.Ident{
						Name: "baseValue",
					},
				},
			},
		},
	})

	err := e.generateType("Enum" + strings.ReplaceAll(e.currentPath, ".", ""))
	if err != nil {
		return nil, err
	}
	return functionScaffold, nil
}

func (e *EnumTypeChecker) GenerateToTypeFunction(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, error) {
	functionScaffold.Body.List = append(functionScaffold.Body.List, &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X: &ast.Ident{
				Name: "baseValue",
			},
			Op: token.EQL,
			Y: &ast.Ident{
				Name: "Enum" + strings.ReplaceAll(e.currentPath, ".", "") + "InvalidEnumValue",
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
	return functionScaffold, nil
}

func (e *EnumTypeChecker) GetRequiredImports() []string {
	return []string{"errors"}
}

func (e *EnumTypeChecker) SetFile(file *ast.File) {
	e.file = file
}

func (e *EnumTypeChecker) GetName() string {
	return "json2go.EnumTypeChecker"
}

func (e *EnumTypeChecker) SetState(state json.RawMessage, currentPath string) error {
	e.currentPath = currentPath
	e.state = nil
	if state != nil {
		err := json.Unmarshal(state, e.state)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *EnumTypeChecker) GetState() (json.RawMessage, error) {
	if e.state == nil {
		return nil, nil
	}
	return json.Marshal(e.state)
}

func (e *EnumTypeChecker) generateType(enumName string) error {
	//TODO check if type is existent, if so delete first
	var err error
	var decls []ast.Decl

	//Add enum type
	decls = append(decls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{
					Name: enumName,
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
								Name: enumName + strcase.ToCamel(value),
							},
						},
						Type: &ast.Ident{
							Name: enumName,
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
								Name: enumName + strcase.ToCamel(value),
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
			Name: "New" + enumName,
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
							Name: enumName,
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
										func() *ast.BasicLit {
											fmt.Println(value)
											switch e.fieldType {
											case "string":
												return &ast.BasicLit{
													Kind:  token.STRING,
													Value: "\"" + value + "\"",
												}
											case "int":
												return &ast.BasicLit{
													Kind:  token.INT,
													Value: value,
												}
											case "float":
												return &ast.BasicLit{
													Kind:  token.FLOAT,
													Value: value,
												}
											default:
												err = errors.New(fmt.Sprintf("encountered an unsupported type: %s", e.fieldType))
												return nil
											}
										}(),
									},
									Body: []ast.Stmt{
										&ast.ReturnStmt{
											Results: []ast.Expr{
												&ast.Ident{
													Name: enumName + strcase.ToCamel(value),
												},
												&ast.Ident{
													Name: "nil",
												},
											},
										},
									},
								})
								if err != nil {
									break
								}
							}
							stmts = append(stmts, &ast.CaseClause{
								Body: []ast.Stmt{
									&ast.ReturnStmt{
										Results: []ast.Expr{
											&ast.Ident{
												Name: enumName + "InvalidEnumValue",
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
						Name: enumName,
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
											Name: enumName + strcase.ToCamel(value),
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

	if err != nil {
		return err
	}

	//If no error happened, attach all generated code to the file
	e.file.Decls = append(e.file.Decls, decls...)

	return nil
}
