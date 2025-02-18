package utils

import (
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
	"strings"
)

func WalkExpressions(expr *ast.Expr) (*ast.Expr, error) {
	switch e := (*expr).(type) {
	case *ast.Ident:
		return expr, nil
	case *ast.StarExpr:
		return WalkExpressions(&e.X)
	case *ast.ArrayType:
		return WalkExpressions(&e.Elt)
	case *ast.InterfaceType:
		return expr, nil
	case *ast.SelectorExpr:
		return expr, nil
	case *ast.StructType:
		return expr, nil
	}
	return nil, errors.New(fmt.Sprintf("unknown expression: %s", reflect.TypeOf(*expr)))
}

func WalkArrays(expr *ast.Expr) (*ast.Expr, error) {
	switch e := (*expr).(type) {
	case *ast.Ident:
		return expr, nil
	case *ast.StarExpr:
		return expr, nil
	case *ast.ArrayType:
		return WalkExpressions(&e.Elt)
	case *ast.InterfaceType:
		return expr, nil
	case *ast.SelectorExpr:
		return expr, nil
	case *ast.StructType:
		return expr, nil
	}
	return nil, errors.New(fmt.Sprintf("unknown expression: %s", reflect.TypeOf(*expr)))
}

func WalkExpressionsWhitArrayCount(expr *ast.Expr) (*ast.Expr, int, error) {
	var levelOfArrays int
	switch e := (*expr).(type) {
	case *ast.Ident:
		return expr, 0, nil
	case *ast.StarExpr:
		return WalkExpressionsWhitArrayCount(&e.X)
	case *ast.ArrayType:
		levelOfArrays++
		expressions, arrayCount, err := WalkExpressionsWhitArrayCount(&e.Elt)
		if err != nil {
			return nil, 0, err
		}
		levelOfArrays += arrayCount
		return expressions, levelOfArrays, nil
	case *ast.InterfaceType:
		return expr, 0, nil
	case *ast.SelectorExpr:
		return expr, 0, nil
	case *ast.StructType:
		return expr, 0, nil
	}
	return nil, 0, errors.New(fmt.Sprintf("unknown expression: %s", reflect.TypeOf(*expr)))
}

func GeneratedNestedArray(levelOfArrays int, innerExpr ast.Expr) ast.Expr {
	if levelOfArrays == 0 {
		return innerExpr
	}
	var ar ast.Expr
	ar = &ast.ArrayType{}
	var oe ast.Expr
	var ie *ast.Expr
	oe = ar
	ie = &ar

	for range levelOfArrays - 1 {
		(*ie).(*ast.ArrayType).Elt = &ast.ArrayType{}
		*ie = (*ie).(*ast.ArrayType).Elt
	}
	(*ie).(*ast.ArrayType).Elt = innerExpr
	return oe
}

func GetFieldIdentFromPath(path string) []*ast.Ident {
	pathElements := strings.Split(path, ".")
	return []*ast.Ident{{Name: strcase.ToCamel(pathElements[len(pathElements)-1])}}
}

func GetInputType(functionScaffold *ast.FuncDecl) (string, error) {
	for _, expr := range functionScaffold.Type.Params.List {
		n, err := WalkExpressions(&expr.Type)
		if err != nil {
			return "", err
		}
		switch e := (*n).(type) {
		case *ast.SelectorExpr:
			return e.Sel.Name + "." + e.X.(*ast.Ident).Name, nil
		case *ast.Ident:
			return e.Name, nil
		case *ast.InterfaceType:
			return "interface{}", nil
		}
	}
	return "", errors.New("no valid input type")
}

func GetReturnType(functionScaffold *ast.FuncDecl) (string, error) {
	for _, expr := range functionScaffold.Type.Results.List {
		n, err := WalkExpressions(&expr.Type)
		if err != nil {
			return "", err
		}
		switch e := (*n).(type) {
		case *ast.SelectorExpr:
			return e.Sel.Name + "." + e.X.(*ast.Ident).Name, nil
		case *ast.Ident:
			return e.Name, nil
		case *ast.InterfaceType:
			return "interface{}", nil
		}
	}
	return "", errors.New("no valid result type")
}

func GetTypeFromBaseType(baseType string) ast.Expr {
	if baseType == "interface{}" {
		fmt.Println("interface{}")
		return &ast.InterfaceType{Methods: &ast.FieldList{}}
	} else {
		return &ast.Ident{Name: baseType}
	}
}

func GenerateIndexExpr(levelOfArrays int, innerExpr ast.Expr, indexName string) ast.Expr {
	if levelOfArrays == 0 {
		return innerExpr
	}
	var ind ast.Expr
	ind = &ast.IndexExpr{}
	var oe ast.Expr
	var ie *ast.Expr
	oe = ind
	ie = &ind

	for i := range levelOfArrays - 1 {
		(*ie).(*ast.IndexExpr).Index = &ast.Ident{Name: indexName + strconv.Itoa(levelOfArrays-1-i)}
		(*ie).(*ast.IndexExpr).X = &ast.IndexExpr{}
		*ie = (*ie).(*ast.IndexExpr).X
	}
	(*ie).(*ast.IndexExpr).X = innerExpr
	(*ie).(*ast.IndexExpr).Index = &ast.Ident{Name: indexName + "0"}
	return oe
}

func GenerateAssignStmt(levelOfArraysLeft, levelOfArraysRight int, leftSide, rightSide ast.Expr, indexName string) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			GenerateIndexExpr(levelOfArraysLeft, leftSide, indexName),
		},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			func() ast.Expr {
				if levelOfArraysRight == 0 {
					return GeneratedNestedArray(levelOfArraysRight, rightSide)
				} else {
					return &ast.CompositeLit{
						Type: GeneratedNestedArray(levelOfArraysRight, rightSide),
					}
				}
			}(),
		},
	}

}

func GenerateAppendStatement(levelOfArraysLeft, levelOfArraysRight int, leftSide, rightSide ast.Expr, indexName string) *ast.AssignStmt {
	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			GenerateIndexExpr(levelOfArraysLeft, leftSide, indexName),
		},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.Ident{
					Name: "append",
				},
				Args: []ast.Expr{
					GenerateIndexExpr(levelOfArraysLeft, leftSide, indexName),
					func() ast.Expr {
						if levelOfArraysRight == 0 {
							return GeneratedNestedArray(levelOfArraysRight, rightSide)
						} else {
							return &ast.CompositeLit{
								Type: GeneratedNestedArray(levelOfArraysRight, rightSide),
							}
						}
					}(),
				},
			},
		},
	}
}

func GenerateNestedRangeStmt(levelOfArrays int, innerStmts []ast.Stmt, rangeExpr ast.Expr, name ast.Expr, appendType ast.Expr) ast.Stmt {
	if levelOfArrays == 0 {
		return innerStmts[0]
	}

	var ind ast.Stmt
	ind = &ast.RangeStmt{
		Key: &ast.Ident{
			Name: "index0",
		},
		Value: &ast.Ident{
			Name: "value0",
		},
		Tok:  token.DEFINE,
		X:    rangeExpr,
		Body: &ast.BlockStmt{List: []ast.Stmt{}},
	}
	var oe ast.Stmt
	var ie *ast.Stmt
	oe = ind
	ie = &ind
	for i := range levelOfArrays - 1 {
		(*ie).(*ast.RangeStmt).Body.List = append((*ie).(*ast.RangeStmt).Body.List, GenerateAppendStatement(i, levelOfArrays-1-i, name, appendType, "index"))
		(*ie).(*ast.RangeStmt).Body.List = append((*ie).(*ast.RangeStmt).Body.List, &ast.RangeStmt{
			Key: &ast.Ident{
				Name: func() string {
					if i == levelOfArrays-2 {
						return "_"
					} else {
						return "index" + strconv.Itoa(i+1)
					}
				}(),
			},
			Value: &ast.Ident{
				Name: func() string {
					if i == levelOfArrays-2 {
						return "baseValue"
					} else {
						return "value" + strconv.Itoa(i+1)
					}
				}(),
			},
			Tok:  token.DEFINE,
			X:    &ast.Ident{Name: "value" + strconv.Itoa(i)},
			Body: &ast.BlockStmt{List: []ast.Stmt{}},
		})
		*ie = (*ie).(*ast.RangeStmt).Body.List[1]
	}
	(*ie).(*ast.RangeStmt).Body.List = innerStmts
	return oe
}

func GetLevelOfArrays(expr ast.Expr) int {
	var levelOfArrays int
	var fieldType ast.Expr
	fieldType = expr
	for {
		if _, ok := fieldType.(*ast.ArrayType); ok {
			levelOfArrays++
			fieldType = fieldType.(*ast.ArrayType).Elt
		} else {
			break
		}
	}
	return levelOfArrays
}

func GenerateTypeWhitInitializedArrays(str *ast.StructType) *ast.AssignStmt {
	var kv []ast.Expr
	for _, field := range str.Fields.List {
		if levelOfArrays := GetLevelOfArrays(field.Type); levelOfArrays > 0 {
			kv = append(kv, &ast.KeyValueExpr{
				Key: &ast.Ident{
					Name: field.Names[0].Name,
				},
				Value: &ast.CompositeLit{
					Type: field.Type,
				},
			})
		}
	}
	return &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{
				Name: "lt",
			},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CompositeLit{
				Type: &ast.Ident{
					Name: "localType",
				},
				Elts: kv,
			},
		},
	}
}
