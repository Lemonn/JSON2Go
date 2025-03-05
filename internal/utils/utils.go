package utils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"reflect"
	"strconv"
	"strings"
	"unicode"
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
	return []*ast.Ident{{Name: JsonNameToGoName(pathElements[len(pathElements)-1])}}
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

func StringToPointer(str string) *string {
	return &str
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

func GetEmptyFile(packageName string) *ast.File {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", "package "+packageName, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	file.Decls = []ast.Decl{}
	return file
}

// TODO this function needs to handle more possible edge cases
func JsonNameToGoName(str string) string {
	if unicode.IsNumber(rune(str[0])) {
		str = "number_" + str
	}
	if len(str) == 2 && str[1] == '_' {
		return str
	} else if len(str) == 1 && unicode.IsLower(rune(str[0])) {
		return strcase.ToCamel(str) + "_"
	} else {
		return strcase.ToCamel(str)
	}
}

/*
func GetFieldType(seenTypes map[string]*fieldData.PathData, path string) (expr ast.Expr, structType bool) {
	levelOfArrays := math.MaxInt32
	//TODO error on path not found
	if len(seenTypes[path].Types) == 1 {
		var Type fieldData.Type
		for Type, _ = range seenTypes[path].Types {
			break
		}
		if len(seenTypes[path].Types[Type]) == 1 {
			for levelOfArrays, _ = range seenTypes[path].Types[Type] {
				break
			}
			pathElements := strings.Split(path, ".")
			if Type == fieldData.Field {
				structType = true
				expr = GeneratedNestedArray(levelOfArrays, &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: pathElements[len(pathElements)-1]}, Sel: &ast.Ident{Name: pathElements[len(pathElements)-1]}}})
			} else if Type == fieldData.EmptyArray {
				expr = GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
			} else if Type == fieldData.EmptyStruct {
				expr = GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
			} else {
				expr = GeneratedNestedArray(levelOfArrays, &ast.Ident{Name: string(Type)})
			}
		} else {
			for i, _ := range seenTypes[path].Types[Type] {
				if levelOfArrays > i {
					levelOfArrays = i
				}
			}
			expr = GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
		}
	} else {
		for _, m := range seenTypes[path].Types {
			for i, _ := range m {
				if levelOfArrays > i {
					levelOfArrays = i
				}
			}
		}
		expr = GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
	}
	return expr, structType
}

*/

func GetFieldName(path string) string {
	pathElements := strings.Split(path, ".")
	return pathElements[len(pathElements)-1]
}

func GetParentPath(path string) string {
	pathElements := strings.Split(path, ".")
	if len(pathElements) > 1 {
		return strings.Join(pathElements[:len(pathElements)-1], ".")
	} else {
		return path
	}
}

func GetParentFieldName(path string) string {
	pathElements := strings.Split(path, ".")
	return pathElements[len(pathElements)-2]
}

func GetPackageNameFromImportPath(importPath string) string {
	pathElements := strings.Split(importPath, "/")
	return pathElements[len(pathElements)-1]
}

func GetAllWrappedErrors(e error) []error {
	if e == nil {
		return nil
	}

	var result []error
UNWRAP:
	switch err := e.(type) {
	case interface {
		Unwrap() []error
	}:
		if reflect.TypeOf(err).String() != "*errors.joinError" {
			result = append(result, err.(error))
		}

		if len(err.Unwrap()) > 0 {
			e = err.Unwrap()[0]
			goto UNWRAP
		} else {
			return result
		}
	default:
		if len(result) > 0 {
			return result
		} else {
			result = append(result, e)
			return result
		}
	}
}

func ExprToString(expr ast.Expr) (string, error) {
	out := bytes.NewBuffer([]byte{})
	err := printer.Fprint(out, token.NewFileSet(), expr)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func IsRootPath(path string) bool {
	if len(strings.Split(path, ".")) == 1 {
		return true
	}
	return false
}
