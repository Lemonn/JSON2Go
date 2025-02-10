package utils

import (
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	"go/ast"
	"reflect"
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

func GeneratedNestedArray(levelOfArrays int, InnerExpr *ast.Expr, OuterExpr ast.Expr) (*ast.Expr, ast.Expr) {
	for range levelOfArrays {
		if InnerExpr != nil && reflect.TypeOf(*InnerExpr) == reflect.TypeOf(&ast.ArrayType{}) {
			(*InnerExpr).(*ast.ArrayType).Elt = &ast.ArrayType{}
			*InnerExpr = (*InnerExpr).(*ast.ArrayType).Elt
		} else {
			var k ast.Expr
			k = &ast.ArrayType{}
			InnerExpr = &k
			OuterExpr = k
		}
	}
	return InnerExpr, OuterExpr
}

func GetFieldIdentFromPath(path string) []*ast.Ident {
	pathElements := strings.Split(path, ".")
	return []*ast.Ident{{Name: strcase.ToCamel(pathElements[len(pathElements)-1])}}
}
