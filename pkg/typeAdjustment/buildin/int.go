package buildin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/codeGenerators"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"go/ast"
	"go/token"
	"math"
	"strconv"
)

type IntTypeChecker struct {
	currentPath fieldData.Path
	fileData    fieldData.FileData
	fieldType   intTypeCheckerFieldType
}

type intTypeCheckerFieldType int

const (
	intTypeCheckerFieldTypeFloat64 intTypeCheckerFieldType = iota
	intTypeCheckerFieldTypeString
	intTypeCheckerFieldTypeMixed
)

func (i *IntTypeChecker) CouldTypeBeApplied() (typeAdjustment.State, error) {
	var f, s bool
	for Type, levels := range i.fileData[i.currentPath].Types {
		if len(levels) > 1 {
			return typeAdjustment.StateFailed, nil
		} else if !(Type == fieldData.String || Type == fieldData.Float64) {
			return typeAdjustment.StateFailed, nil
		} else if Type == fieldData.Float64 {
			f = true
		} else {
			s = true
		}
	}

	if f {
		var levelOfArrays int
		for levelOfArrays, _ = range i.fileData[i.currentPath].Types[fieldData.Float64] {
			break
		}
		for value, _ := range i.fileData[i.currentPath].Types[fieldData.Float64][levelOfArrays] {
			t, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return typeAdjustment.StateFailed, nil
			}
			if t > math.MaxInt {
				return typeAdjustment.StateFailed, nil
			}
			const epsilon = 1e-9 // Margin of error
			if _, frac := math.Modf(math.Abs(t)); frac < epsilon || frac > 1.0-epsilon {
				continue
			} else {
				return typeAdjustment.StateFailed, nil
			}
		}
	}

	if s {
		var levelOfArrays int
		for levelOfArrays, _ = range i.fileData[i.currentPath].Types[fieldData.String] {
			break
		}
		for value, _ := range i.fileData[i.currentPath].Types[fieldData.String][levelOfArrays] {
			_, err := strconv.Atoi(value)
			if err != nil {
				return typeAdjustment.StateFailed, nil
			}
		}
	}

	if f && s {
		i.fieldType = intTypeCheckerFieldTypeMixed
	} else if f {
		i.fieldType = intTypeCheckerFieldTypeFloat64
	} else {
		i.fieldType = intTypeCheckerFieldTypeString
	}
	return typeAdjustment.StateApplicable, nil
}

func (i *IntTypeChecker) GenerateMarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, fieldData.Imports, error) {
	if i.fieldType == intTypeCheckerFieldTypeFloat64 {
		functionScaffold.Body.List = []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.Ident{
							Name: "float64",
						},
						Args: []ast.Expr{
							&ast.Ident{
								Name: "baseValue",
							},
						},
					},
					&ast.Ident{
						Name: "nil",
					},
				},
			},
		}
		return functionScaffold, nil, nil
	} else if i.fieldType == intTypeCheckerFieldTypeString {
		functionScaffold.Body.List = []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "strconv",
							},
							Sel: &ast.Ident{
								Name: "Itoa",
							},
						},
						Args: []ast.Expr{
							&ast.Ident{
								Name: "baseValue",
							},
						},
					},
					&ast.Ident{
						Name: "nil",
					},
				},
			},
		}
		return functionScaffold, fieldData.Imports{&fieldData.Import{Path: fieldData.Path("strconv")}}, nil
	} else {
		//TODO right now this prefers float64 whit mixed types, make this a setting
		functionScaffold.Body.List = []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.Ident{
							Name: "float64",
						},
						Args: []ast.Expr{
							&ast.Ident{
								Name: "baseValue",
							},
						},
					},
					&ast.Ident{
						Name: "nil",
					},
				},
			},
		}
		return functionScaffold, nil, nil
	}
}

func test(baseValue interface{}) (int, error) {
	switch t := baseValue.(type) {
	case float64:
		if t > math.MaxInt {
			return 0, errors.New("int overflow")
		}
		const epsilon = 1e-9 // Margin of error
		if _, frac := math.Modf(math.Abs(t)); frac < epsilon || frac > 1.0-epsilon {
			return int(t), nil
		} else {
			return 0, errors.New(fmt.Sprintf("the given float value: %f cloud not be represented as an intiger.", t))
		}
	case string:
		s, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0, err
		}
		if s > math.MaxInt {
			return 0, errors.New("int overflow")
		}
		const epsilon = 1e-9 // Margin of error
		if _, frac := math.Modf(math.Abs(s)); frac < epsilon || frac > 1.0-epsilon {
			return int(s), nil
		} else {
			return 0, errors.New(fmt.Sprintf("the given string value: %s cloud not be represented as an intiger.", t))
		}
	default:
		return 0, errors.New(fmt.Sprintf("unsupported type %T, only string and float64 are supported", baseValue))
	}
}

func (i *IntTypeChecker) GenerateUnmarshall(functionScaffold *ast.FuncDecl) (*ast.FuncDecl, fieldData.Imports, error) {
	if i.fieldType == intTypeCheckerFieldTypeFloat64 {
		functionScaffold.Body.List = []ast.Stmt{
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X: &ast.Ident{
						Name: "baseValue",
					},
					Op: token.GTR,
					Y: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: "math",
						},
						Sel: &ast.Ident{
							Name: "MaxInt",
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.BasicLit{
									Kind:  token.INT,
									Value: "0",
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
										&ast.BasicLit{
											Kind:  token.STRING,
											Value: "\"int overflow\"",
										},
									},
								},
							},
						},
					},
				},
			},
			&ast.DeclStmt{
				Decl: &ast.GenDecl{
					Tok: token.CONST,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names: []*ast.Ident{
								&ast.Ident{
									Name: "epsilon",
								},
							},
							Values: []ast.Expr{
								&ast.BasicLit{
									Kind:  token.FLOAT,
									Value: "1e-9",
								},
							},
							Comment: &ast.CommentGroup{
								List: []*ast.Comment{
									&ast.Comment{
										Slash: 126,
										Text:  "// Margin of error",
									},
								},
							},
						},
					},
				},
			},
			&ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: "_",
						},
						&ast.Ident{
							Name: "frac",
						},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "math",
								},
								Sel: &ast.Ident{
									Name: "Modf",
								},
							},
							Args: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "math",
										},
										Sel: &ast.Ident{
											Name: "Abs",
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
				},
				Cond: &ast.BinaryExpr{
					X: &ast.BinaryExpr{
						X: &ast.Ident{
							Name: "frac",
						},
						Op: token.LSS,
						Y: &ast.Ident{
							Name: "epsilon",
						},
					},
					Op: token.LOR,
					Y: &ast.BinaryExpr{
						X: &ast.Ident{
							Name: "frac",
						},
						Op: token.GTR,
						Y: &ast.BinaryExpr{
							X: &ast.BasicLit{
								Kind:  token.FLOAT,
								Value: "1.0",
							},
							Op: token.SUB,
							Y: &ast.Ident{
								Name: "epsilon",
							},
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.Ident{
										Name: "int",
									},
									Args: []ast.Expr{
										&ast.Ident{
											Name: "baseValue",
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
				Else: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.BasicLit{
									Kind:  token.INT,
									Value: "0",
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
													Value: "\"the given float value: %f cloud not be represented as an intiger.\"",
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
			},
		}
		return functionScaffold, fieldData.Imports{
			&fieldData.Import{
				Path: "fmt",
			},
			&fieldData.Import{
				Path: "math",
			},
		}, nil
	} else if i.fieldType == intTypeCheckerFieldTypeString {
		functionScaffold.Body.List = []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{
						Name: "s",
					},
					&ast.Ident{
						Name: "err",
					},
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "strconv",
							},
							Sel: &ast.Ident{
								Name: "ParseFloat",
							},
						},
						Args: []ast.Expr{
							&ast.Ident{
								Name: "baseValue",
							},
							&ast.BasicLit{
								Kind:  token.INT,
								Value: "64",
							},
						},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X: &ast.Ident{
						Name: "err",
					},
					Op: token.NEQ,
					Y: &ast.Ident{
						Name: "nil",
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.BasicLit{
									Kind:  token.INT,
									Value: "0",
								},
								&ast.Ident{
									Name: "err",
								},
							},
						},
					},
				},
			},
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X: &ast.Ident{
						Name: "s",
					},
					Op: token.GTR,
					Y: &ast.SelectorExpr{
						X: &ast.Ident{
							Name: "math",
						},
						Sel: &ast.Ident{
							Name: "MaxInt",
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.BasicLit{
									Kind:  token.INT,
									Value: "0",
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
										&ast.BasicLit{
											Kind:  token.STRING,
											Value: "\"int overflow\"",
										},
									},
								},
							},
						},
					},
				},
			},
			&ast.DeclStmt{
				Decl: &ast.GenDecl{
					Tok: token.CONST,
					Specs: []ast.Spec{
						&ast.ValueSpec{
							Names: []*ast.Ident{
								&ast.Ident{
									Name: "epsilon",
								},
							},
							Values: []ast.Expr{
								&ast.BasicLit{
									Kind:  token.FLOAT,
									Value: "1e-9",
								},
							},
							Comment: &ast.CommentGroup{
								List: []*ast.Comment{
									&ast.Comment{
										Slash: 203,
										Text:  "// Margin of error",
									},
								},
							},
						},
					},
				},
			},
			&ast.IfStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: "_",
						},
						&ast.Ident{
							Name: "frac",
						},
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "math",
								},
								Sel: &ast.Ident{
									Name: "Modf",
								},
							},
							Args: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "math",
										},
										Sel: &ast.Ident{
											Name: "Abs",
										},
									},
									Args: []ast.Expr{
										&ast.Ident{
											Name: "s",
										},
									},
								},
							},
						},
					},
				},
				Cond: &ast.BinaryExpr{
					X: &ast.BinaryExpr{
						X: &ast.Ident{
							Name: "frac",
						},
						Op: token.LSS,
						Y: &ast.Ident{
							Name: "epsilon",
						},
					},
					Op: token.LOR,
					Y: &ast.BinaryExpr{
						X: &ast.Ident{
							Name: "frac",
						},
						Op: token.GTR,
						Y: &ast.BinaryExpr{
							X: &ast.BasicLit{
								Kind:  token.FLOAT,
								Value: "1.0",
							},
							Op: token.SUB,
							Y: &ast.Ident{
								Name: "epsilon",
							},
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.Ident{
										Name: "int",
									},
									Args: []ast.Expr{
										&ast.Ident{
											Name: "s",
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
				Else: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ReturnStmt{
							Results: []ast.Expr{
								&ast.BasicLit{
									Kind:  token.INT,
									Value: "0",
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
													Value: "\"the given string value: %s cloud not be represented as an intiger.\"",
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
			},
		}
		return functionScaffold, fieldData.Imports{
			{Path: "fmt"},
			{Path: "math"},
			{Path: "strconv"},
		}, nil
	} else {
		functionScaffold.Body.List = []ast.Stmt{
			&ast.CaseClause{
				List: []ast.Expr{
					&ast.Ident{
						Name: "float64",
					},
				},
				Body: []ast.Stmt{
					&ast.IfStmt{
						Cond: &ast.BinaryExpr{
							X: &ast.Ident{
								Name: "t",
							},
							Op: token.GTR,
							Y: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "math",
								},
								Sel: &ast.Ident{
									Name: "MaxInt",
								},
							},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.INT,
											Value: "0",
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
												&ast.BasicLit{
													Kind:  token.STRING,
													Value: "\"int overflow\"",
												},
											},
										},
									},
								},
							},
						},
					},
					&ast.DeclStmt{
						Decl: &ast.GenDecl{
							Tok: token.CONST,
							Specs: []ast.Spec{
								&ast.ValueSpec{
									Names: []*ast.Ident{
										&ast.Ident{
											Name: "epsilon",
										},
									},
									Values: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.FLOAT,
											Value: "1e-9",
										},
									},
									Comment: &ast.CommentGroup{
										List: []*ast.Comment{
											&ast.Comment{
												Slash: 165,
												Text:  "// Margin of error",
											},
										},
									},
								},
							},
						},
					},
					&ast.IfStmt{
						Init: &ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{
									Name: "_",
								},
								&ast.Ident{
									Name: "frac",
								},
							},
							Tok: token.DEFINE,
							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "math",
										},
										Sel: &ast.Ident{
											Name: "Modf",
										},
									},
									Args: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X: &ast.Ident{
													Name: "math",
												},
												Sel: &ast.Ident{
													Name: "Abs",
												},
											},
											Args: []ast.Expr{
												&ast.Ident{
													Name: "t",
												},
											},
										},
									},
								},
							},
						},
						Cond: &ast.BinaryExpr{
							X: &ast.BinaryExpr{
								X: &ast.Ident{
									Name: "frac",
								},
								Op: token.LSS,
								Y: &ast.Ident{
									Name: "epsilon",
								},
							},
							Op: token.LOR,
							Y: &ast.BinaryExpr{
								X: &ast.Ident{
									Name: "frac",
								},
								Op: token.GTR,
								Y: &ast.BinaryExpr{
									X: &ast.BasicLit{
										Kind:  token.FLOAT,
										Value: "1.0",
									},
									Op: token.SUB,
									Y: &ast.Ident{
										Name: "epsilon",
									},
								},
							},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.Ident{
												Name: "int",
											},
											Args: []ast.Expr{
												&ast.Ident{
													Name: "t",
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
						Else: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.INT,
											Value: "0",
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
															Value: "\"the given float value: %f cloud not be represented as an intiger.\"",
														},
														&ast.Ident{
															Name: "t",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			&ast.CaseClause{
				List: []ast.Expr{
					&ast.Ident{
						Name: "string",
					},
				},
				Body: []ast.Stmt{
					&ast.AssignStmt{
						Lhs: []ast.Expr{
							&ast.Ident{
								Name: "s",
							},
							&ast.Ident{
								Name: "err",
							},
						},
						Tok: token.DEFINE,
						Rhs: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X: &ast.Ident{
										Name: "strconv",
									},
									Sel: &ast.Ident{
										Name: "ParseFloat",
									},
								},
								Args: []ast.Expr{
									&ast.Ident{
										Name: "t",
									},
									&ast.BasicLit{
										Kind:  token.INT,
										Value: "64",
									},
								},
							},
						},
					},
					&ast.IfStmt{
						Cond: &ast.BinaryExpr{
							X: &ast.Ident{
								Name: "err",
							},
							Op: token.NEQ,
							Y: &ast.Ident{
								Name: "nil",
							},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.INT,
											Value: "0",
										},
										&ast.Ident{
											Name: "err",
										},
									},
								},
							},
						},
					},
					&ast.IfStmt{
						Cond: &ast.BinaryExpr{
							X: &ast.Ident{
								Name: "s",
							},
							Op: token.GTR,
							Y: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "math",
								},
								Sel: &ast.Ident{
									Name: "MaxInt",
								},
							},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.INT,
											Value: "0",
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
												&ast.BasicLit{
													Kind:  token.STRING,
													Value: "\"int overflow\"",
												},
											},
										},
									},
								},
							},
						},
					},
					&ast.DeclStmt{
						Decl: &ast.GenDecl{
							Tok: token.CONST,
							Specs: []ast.Spec{
								&ast.ValueSpec{
									Names: []*ast.Ident{
										&ast.Ident{
											Name: "epsilon",
										},
									},
									Values: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.FLOAT,
											Value: "1e-9",
										},
									},
									Comment: &ast.CommentGroup{
										List: []*ast.Comment{
											&ast.Comment{
												Slash: 590,
												Text:  "// Margin of error",
											},
										},
									},
								},
							},
						},
					},
					&ast.IfStmt{
						Init: &ast.AssignStmt{
							Lhs: []ast.Expr{
								&ast.Ident{
									Name: "_",
								},
								&ast.Ident{
									Name: "frac",
								},
							},
							Tok: token.DEFINE,
							Rhs: []ast.Expr{
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "math",
										},
										Sel: &ast.Ident{
											Name: "Modf",
										},
									},
									Args: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X: &ast.Ident{
													Name: "math",
												},
												Sel: &ast.Ident{
													Name: "Abs",
												},
											},
											Args: []ast.Expr{
												&ast.Ident{
													Name: "s",
												},
											},
										},
									},
								},
							},
						},
						Cond: &ast.BinaryExpr{
							X: &ast.BinaryExpr{
								X: &ast.Ident{
									Name: "frac",
								},
								Op: token.LSS,
								Y: &ast.Ident{
									Name: "epsilon",
								},
							},
							Op: token.LOR,
							Y: &ast.BinaryExpr{
								X: &ast.Ident{
									Name: "frac",
								},
								Op: token.GTR,
								Y: &ast.BinaryExpr{
									X: &ast.BasicLit{
										Kind:  token.FLOAT,
										Value: "1.0",
									},
									Op: token.SUB,
									Y: &ast.Ident{
										Name: "epsilon",
									},
								},
							},
						},
						Body: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.Ident{
												Name: "int",
											},
											Args: []ast.Expr{
												&ast.Ident{
													Name: "s",
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
						Else: &ast.BlockStmt{
							List: []ast.Stmt{
								&ast.ReturnStmt{
									Results: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.INT,
											Value: "0",
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
															Value: "\"the given string value: %s cloud not be represented as an intiger.\"",
														},
														&ast.Ident{
															Name: "t",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			&ast.CaseClause{
				Body: []ast.Stmt{
					&ast.ReturnStmt{
						Results: []ast.Expr{
							&ast.BasicLit{
								Kind:  token.INT,
								Value: "0",
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
												Value: "\"unsupported type %T, only string and float64 are supported\"",
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
		}
		return functionScaffold, fieldData.Imports{
			{Path: "fmt"},
			{Path: "math"},
			{Path: "strconv"},
		}, nil
	}
}

func (i *IntTypeChecker) SetState(_ []json.RawMessage, currentPath fieldData.Path, fileData fieldData.FileData, _ codeGenerators.CodeGenerator) error {
	i.currentPath = currentPath
	i.fileData = fileData
	return nil
}

func (i *IntTypeChecker) GetState() (json.RawMessage, error) {
	return nil, nil
}

func (i *IntTypeChecker) GetExtraCode() ([]ast.Decl, []string, error) {
	return nil, nil, nil
}

func (i *IntTypeChecker) TypeExpansion() bool {
	//TODO it's a type expansion, if mixed type is found
	return false
}

func (i *IntTypeChecker) ForceSourceType() *string {
	return nil
}

func (i *IntTypeChecker) GetModFileContents() []*fieldData.ModFileContent {
	return nil
}

func (i *IntTypeChecker) GetVersion() *string {
	return utils.StringToPointer("v0.0.1")
}

func (i *IntTypeChecker) GetType() (ast.Expr, fieldData.Imports) {
	return &ast.Ident{Name: "int"}, nil
}

func (i *IntTypeChecker) GetName() string {
	return "json2go.IntTypeChecker"
}

func (i *IntTypeChecker) GetSubFiles() (map[fieldData.Path][]*fieldData.File, error) {
	return nil, nil
}

func (i *IntTypeChecker) NeedsMarshaller() bool {
	return true
}

func (i *IntTypeChecker) Clone() typeAdjustment.TypeDeterminationFunction {
	return &IntTypeChecker{}
}
