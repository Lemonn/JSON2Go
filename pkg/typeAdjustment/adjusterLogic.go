package typeAdjustment

import (
	"github.com/Lemonn/AstUtils"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"reflect"
	"strings"
	"time"
)

type TypeAdjuster struct {
	data map[string]*fieldData.FieldData
}

func NewTypeAdjuster(data map[string]*fieldData.FieldData) *TypeAdjuster {
	return &TypeAdjuster{data: data}
}

// AdjustTypes Goes through all fields and looks at the json2go FieldData, to determine if there's a better suiting type
// for the seen float and string values.
// Floats which could be represented as an int, are changed to int
// Strings which could be represented as UUID are change into uuid.UUID
// Strings which could be represented as time, are changed into time.Time
func (ta *TypeAdjuster) AdjustTypes(file *ast.File, registeredTypeCheckers []TypeDeterminationFunction, skipPreviouslyFailed bool) error {
	var foundNodes []*AstUtils.FoundNodes
	var completed bool
	var requiredImports []string
	AstUtils.SearchNodes(file, &foundNodes, []*ast.Node{}, func(n *ast.Node, parents []*ast.Node, completed *bool) bool {
		if _, ok := (*n).(*ast.StructType); ok {
			return true
		} else if _, ok := (*n).(*ast.ArrayType); ok {
			return true
		}
		return false
	}, &completed)
	for _, node := range foundNodes {

		var path string
		for _, parent := range node.Parents {
			if v, ok := (*parent).(*ast.TypeSpec); ok {
				path += v.Name.Name
				break
			} else if v, ok := (*parent).(*ast.Field); ok {
				path += "." + v.Names[0].Name
			} else if _, ok := (*parent).(*ast.StructType); ok {
				//Ignore nested structs
				continue
			}
		}

		switch t := (*node.Node).(type) {
		case *ast.ArrayType:
			// Ignore if array is of type *ast.Struct
			expr, err := utils.WalkExpressions(&t.Elt)
			if err != nil {
				return err
			} else if reflect.TypeOf(expr) == reflect.TypeOf(&ast.StructType{}) {
				continue
			}
			e := ast.Expr(t)
			localRequiredImports, err := ta.runTypeCheckers(file, registeredTypeCheckers, path, path, &e)
			if err != nil {
				return err
			}
			requiredImports = append(requiredImports, localRequiredImports...)
		case *ast.StructType:
			for _, field := range t.Fields.List {
				localRequiredImports, err := ta.runTypeCheckers(file, registeredTypeCheckers, path+"."+field.Names[0].Name, field.Names[0].Name, &field.Type)
				if err != nil {
					return err
				}
				requiredImports = append(requiredImports, localRequiredImports...)
			}
		}

	}
	AstUtils.AddMissingImports(file, requiredImports)
	return nil
}

func (ta *TypeAdjuster) runTypeCheckers(file *ast.File, registeredTypeCheckers []TypeDeterminationFunction, path string, name string, e *ast.Expr) ([]string, error) {
	var requiredImports []string
	var err error

	//Get input type
	var originalType string
	var exp *ast.Expr

	expr, err := utils.WalkExpressions(e)
	if err != nil {
		return nil, err
	}
	switch e := (*expr).(type) {
	case *ast.SelectorExpr:
		originalType = e.Sel.Name + "." + e.X.(*ast.Ident).Name
		exp = expr
	case *ast.Ident:
		originalType = e.Name
		exp = expr
	case *ast.InterfaceType:
		originalType = "interface{}"
		exp = expr
	}

	baseName := strings.ReplaceAll(path, ".", "")
	json2GoTag := ta.data[path]
	if json2GoTag == nil || len(json2GoTag.SeenValues) == 0 || json2GoTag.BaseType != nil {
		return nil, nil
	}

	for _, checker := range registeredTypeCheckers {
		if json2GoTag.SeenValues != nil {
			if _, ok := json2GoTag.SeenValues[checker.GetName()]; ok {
				continue
			}
		}
		checker.SetFile(file)
		if checker.CouldTypeBeApplied(json2GoTag.SeenValues) {
			//Set FieldData
			json2GoTag.ParseFunctions = &fieldData.ParseFunctions{
				FromTypeParseFunction: "from" + baseName,
				ToTypeParseFunction:   "to" + baseName,
			}
			json2GoTag.BaseType = &originalType
			checkerName := checker.GetName()
			json2GoTag.NameOfActiveTypeAdjuster = &checkerName

			fromTypeFunction, err := checker.GenerateFromTypeFunction(&ast.FuncDecl{
				Name: &ast.Ident{
					Name: json2GoTag.ParseFunctions.FromTypeParseFunction,
				},
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{
										Name: "baseValue",
									},
								},
								Type: &ast.Ident{
									Name: originalType,
								},
							},
						},
					},
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: checker.GetType(),
							},
							{
								Type: &ast.Ident{
									Name: "error",
								},
							},
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{},
				},
			})
			if err != nil {
				return nil, err
			}
			toTypeFunction, err := checker.GenerateToTypeFunction(&ast.FuncDecl{
				Name: &ast.Ident{
					Name: json2GoTag.ParseFunctions.ToTypeParseFunction,
				},
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{
									{
										Name: "baseValue",
									},
								},
								Type: checker.GetType(),
							},
						},
					},
					Results: &ast.FieldList{
						List: []*ast.Field{
							{
								Type: &ast.Ident{
									Name: originalType,
								},
							},
							{
								Type: &ast.Ident{
									Name: "error",
								},
							},
						},
					},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{},
				},
			})
			if err != nil {
				return nil, err
			}
			file.Decls = append(file.Decls, fromTypeFunction)
			file.Decls = append(file.Decls, toTypeFunction)
			*exp = checker.GetType()
			json2GoTag.BaseType = &originalType
			requiredImports = append(requiredImports, checker.GetRequiredImports()...)
			break
		} else {
			if json2GoTag.CheckedNonMatchingTypes == nil {
				json2GoTag.CheckedNonMatchingTypes = map[string]int64{}
			}
			json2GoTag.CheckedNonMatchingTypes[checker.GetName()] = time.Now().Unix()
			if err != nil {
				return nil, err
			}
		}
	}
	return requiredImports, nil
}
