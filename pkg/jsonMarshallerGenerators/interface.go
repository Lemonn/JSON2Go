package jsonMarshallerGenerators

import (
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"go/parser"
	"math"
	"strings"
)

type JSONMarshaller interface {
	GetBaseType(path string) (expr ast.Expr)
	IsStruct(path string) bool
	GetFieldName(path string) string
	GetParentFieldName(path string) string
}

type WrappingJSONMarshaller struct {
	seenTypes map[string]*fieldData.PathData
}

func NewWrappingJSONMarshaller(seenTypes map[string]*fieldData.PathData) *WrappingJSONMarshaller {
	return &WrappingJSONMarshaller{
		seenTypes: seenTypes,
	}
}

func (w *WrappingJSONMarshaller) GetBaseType(path string) (expr ast.Expr) {
	//TODO error on path not found

	if w.seenTypes[path].ForceSourceType != nil {
		parseExpr, err := parser.ParseExpr(*w.seenTypes[path].ForceSourceType)
		if err != nil {
			return nil
		}
		return parseExpr
	}

	if len(w.seenTypes[path].Types) == 1 {
		var Type fieldData.Type
		for Type, _ = range w.seenTypes[path].Types {
			break
		}
		if len(w.seenTypes[path].Types[Type]) == 1 {
			pathElements := strings.Split(path, ".")
			fmt.Println(Type)
			if Type == fieldData.Field {
				expr = &ast.StarExpr{X: &ast.Ident{Name: pathElements[len(pathElements)-1]}}
			} else if Type == fieldData.EmptyArray {
				expr = &ast.InterfaceType{Methods: &ast.FieldList{}}
			} else if Type == fieldData.EmptyStruct {
				expr = &ast.InterfaceType{Methods: &ast.FieldList{}}
			} else {
				expr = &ast.Ident{Name: string(Type)}
			}
		} else {
			//TODO replace whit interface{}
			expr = &ast.SelectorExpr{
				X: &ast.Ident{
					Name: "json",
				},
				Sel: &ast.Ident{
					Name: "RawMessage",
				},
			}
		}
	} else {
		//TODO replace whit interface{}
		expr = &ast.SelectorExpr{
			X: &ast.Ident{
				Name: "json",
			},
			Sel: &ast.Ident{
				Name: "RawMessage",
			},
		}
	}

	return expr
}

func (w *WrappingJSONMarshaller) IsStruct(path string) bool {
	if len(w.seenTypes[path].Types) == 1 {
		var Type fieldData.Type
		for Type, _ = range w.seenTypes[path].Types {
			break
		}
		if len(w.seenTypes[path].Types[Type]) == 1 {
			if Type == fieldData.Field {
				return true
			} else {
				return false
			}
		} else {
			return false
		}
	} else {
		return false
	}
}

func (w *WrappingJSONMarshaller) GetFieldName(path string) string {
	pathElements := strings.Split(path, ".")
	return pathElements[len(pathElements)-1]
}

func (w *WrappingJSONMarshaller) GetParentFieldName(path string) string {
	pathElements := strings.Split(path, ".")
	return pathElements[len(pathElements)-2]
}

func (w *WrappingJSONMarshaller) GetLevelOfArrays(path string) int {
	levelOfArrays := math.MaxInt32
	for Type, _ := range w.seenTypes[path].Types {
		for i, _ := range w.seenTypes[path].Types[Type] {
			if levelOfArrays > i {
				levelOfArrays = i
			}
		}
	}
	return levelOfArrays
}

func (w *WrappingJSONMarshaller) GetFieldType(path string) (expr ast.Expr) {
	levelOfArrays := math.MaxInt32
	//TODO error on path not found

	if w.seenTypes[path].ForceSourceType != nil {
		parseExpr, err := parser.ParseExpr(*w.seenTypes[path].ForceSourceType)
		if err != nil {
			return nil
		}
		return parseExpr
	}

	if len(w.seenTypes[path].Types) == 1 {
		var Type fieldData.Type
		for Type, _ = range w.seenTypes[path].Types {
			break
		}
		if len(w.seenTypes[path].Types[Type]) == 1 {
			for levelOfArrays, _ = range w.seenTypes[path].Types[Type] {
				break
			}
			pathElements := strings.Split(path, ".")
			if Type == fieldData.Field {
				expr = utils.GeneratedNestedArray(levelOfArrays, &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: pathElements[len(pathElements)-1]}, Sel: &ast.Ident{Name: pathElements[len(pathElements)-1]}}})
			} else if Type == fieldData.EmptyArray {
				expr = utils.GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
			} else if Type == fieldData.EmptyStruct {
				expr = utils.GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
			} else {
				expr = utils.GeneratedNestedArray(levelOfArrays, &ast.Ident{Name: string(Type)})
			}
		} else {
			for i, _ := range w.seenTypes[path].Types[Type] {
				if levelOfArrays > i {
					levelOfArrays = i
				}
			}
			expr = utils.GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
		}
	} else {
		for _, m := range w.seenTypes[path].Types {
			for i, _ := range m {
				if levelOfArrays > i {
					levelOfArrays = i
				}
			}
		}
		expr = utils.GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
	}
	return expr
}
