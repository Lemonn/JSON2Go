package pkg

import (
	"github.com/araddon/dateparse"
	"github.com/google/uuid"
	"go/ast"
	"go/token"
	"strconv"
)

type TypeDeterminationFunction interface {
	CouldTypeBeApplied(seenValues map[string]string) bool
	GetType() ast.Expr
	GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl
	GenerateToTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl
	GetRequiredImports() []string
	SetFile(file *ast.File)
	GetName() string
}

type TimeTypeChecker struct {
	// IgnoreYearOnlyStrings Set to ignore strings that consist only of a year such as 3294. Most often, they're
	// integers not years!
	IgnoreYearOnlyStrings bool
	layoutString          string
}

func (t *TimeTypeChecker) GetType() ast.Expr {
	return &ast.SelectorExpr{
		X: &ast.Ident{
			Name: "time",
		},
		Sel: &ast.Ident{
			Name: "Time",
		},
	}
}

func (t *TimeTypeChecker) CouldTypeBeApplied(seenValues map[string]string) bool {
	var err error
	for value, _ := range seenValues {
		t.layoutString, err = dateparse.ParseFormat(value)
		if t.IgnoreYearOnlyStrings && t.layoutString == "2006" {
			return false
		}
		if err != nil {
			return false
		}
	}
	return true
}

func (t *TimeTypeChecker) GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl {
	functionScaffold.Body = &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "time",
							},
							Sel: &ast.Ident{
								Name: "Parse",
							},
						},
						Args: []ast.Expr{
							&ast.Ident{
								Name: "\"" + t.layoutString + "\"",
							},
							&ast.Ident{
								Name: "baseValue",
							},
						},
					},
				},
			},
		},
	}
	return functionScaffold
}

func (t *TimeTypeChecker) GenerateToTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl {
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
								Name: "Format",
							},
						},
						Args: []ast.Expr{
							&ast.Ident{
								Name: "\"" + t.layoutString + "\"",
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
	return functionScaffold
}

func (t *TimeTypeChecker) GetRequiredImports() []string {
	return []string{"time"}
}

func (t *TimeTypeChecker) SetFile(_ *ast.File) {}

func (t *TimeTypeChecker) GetName() string {
	return "json2go.TimeTypeChecker"
}

type UUIDTypeChecker struct{}

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
	for value, _ := range seenValues {
		_, err = uuid.Parse(value)
		if err != nil {
			return false
		}
	}
	return true
}

func (u *UUIDTypeChecker) GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl {
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
	return functionScaffold
}

func (u *UUIDTypeChecker) GenerateToTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl {
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
	return functionScaffold
}

func (u *UUIDTypeChecker) GetRequiredImports() []string {
	return []string{"github.com/google/uuid"}
}

func (u *UUIDTypeChecker) SetFile(_ *ast.File) {}

func (u *UUIDTypeChecker) GetName() string {
	return "json2go.UUIDTypeChecker"
}

type IntTypeChecker struct{}

func (i *IntTypeChecker) CouldTypeBeApplied(seenValues map[string]string) bool {
	for value, _ := range seenValues {
		_, err := strconv.Atoi(value)
		if err != nil {
			return false
		}
	}
	return true
}

func (i *IntTypeChecker) GetType() ast.Expr {
	return &ast.Ident{Name: "int"}
}

func (i *IntTypeChecker) GenerateFromTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl {
	if getInputType(functionScaffold) == "float64" {
		functionScaffold.Body = &ast.BlockStmt{
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
		}
	} else {
		functionScaffold.Body = &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: "i",
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
									Name: "Atoi",
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
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.Ident{
							Name: "i",
						},
						&ast.Ident{
							Name: "nil",
						},
					},
				},
			},
		}
	}
	return functionScaffold
}

func (i *IntTypeChecker) GenerateToTypeFunction(functionScaffold *ast.FuncDecl) *ast.FuncDecl {
	if getReturnType(functionScaffold) == "float64" {
		functionScaffold.Body = &ast.BlockStmt{
			List: []ast.Stmt{
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
			},
		}
	} else {
		functionScaffold.Body = &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{
							Name: "i",
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
									Name: "Itoa",
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
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.Ident{
							Name: "i",
						},
						&ast.Ident{
							Name: "nil",
						},
					},
				},
			},
		}
	}
	return functionScaffold
}

func (i *IntTypeChecker) GetRequiredImports() []string {
	return []string{"strconv"}
}

func (i *IntTypeChecker) SetFile(_ *ast.File) {}

func (i *IntTypeChecker) GetName() string {
	return "json2go.IntTypeChecker"
}

func getInputType(functionScaffold *ast.FuncDecl) string {
	for _, expr := range functionScaffold.Type.Params.List {
		n := walkExpressions(&expr.Type)
		switch e := (*n).(type) {
		case *ast.SelectorExpr:
			return e.Sel.Name + "." + e.X.(*ast.Ident).Name
		case *ast.Ident:
			return e.Name
		case *ast.InterfaceType:
			return "interface{}"
		}
	}
	return ""
}

func getReturnType(functionScaffold *ast.FuncDecl) string {
	for _, expr := range functionScaffold.Type.Results.List {
		n := walkExpressions(&expr.Type)
		switch e := (*n).(type) {
		case *ast.SelectorExpr:
			return e.Sel.Name + "." + e.X.(*ast.Ident).Name
		case *ast.Ident:
			return e.Name
		case *ast.InterfaceType:
			return "interface{}"
		}
	}
	return ""
}
