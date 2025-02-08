package utils

import (
	"errors"
	"fmt"
	"go/ast"
	"reflect"
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
