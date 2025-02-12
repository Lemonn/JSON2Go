package typeAdjustment

import (
	"errors"
	"github.com/Lemonn/AstUtils"
	"github.com/Lemonn/JSON2Go/internal/utils"
	j2gErrors "github.com/Lemonn/JSON2Go/pkg/errors"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"reflect"
	"strings"
	"time"
)

type TypeAdjuster struct {
	data map[string]*fieldData.FieldData
	file *ast.File
}

func NewTypeAdjuster(file *ast.File, data map[string]*fieldData.FieldData) *TypeAdjuster {
	return &TypeAdjuster{data: data, file: file}
}

// AdjustTypes Goes through all fields and looks at the json2go FieldData, to determine if there's a better suiting type
// for the seen float and string values.
// Floats which could be represented as an int, are changed to int
// Strings which could be represented as UUID are change into uuid.UUID
// Strings which could be represented as time, are changed into time.Time
func (ta *TypeAdjuster) AdjustTypes(registeredTypeCheckers []TypeDeterminationFunction, skipPreviouslyFailed bool) error {
	var foundNodes []*AstUtils.FoundNodes
	var completed bool
	var requiredImports []string
	AstUtils.SearchNodes(ta.file, &foundNodes, []*ast.Node{}, func(n *ast.Node, parents []*ast.Node, completed *bool) bool {
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
			localRequiredImports, err := ta.runTypeCheckers(registeredTypeCheckers, path, path, &e)
			if err != nil {
				return err
			}
			requiredImports = append(requiredImports, localRequiredImports...)
		case *ast.StructType:
			for _, field := range t.Fields.List {
				localRequiredImports, err := ta.runTypeCheckers(registeredTypeCheckers, path+"."+field.Names[0].Name, field.Names[0].Name, &field.Type)
				if err != nil {
					return err
				}
				requiredImports = append(requiredImports, localRequiredImports...)
			}
		}

	}
	AstUtils.AddMissingImports(ta.file, requiredImports)
	return nil
}

func (ta *TypeAdjuster) runTypeCheckers(registeredTypeCheckers []TypeDeterminationFunction, path string, name string, e *ast.Expr) ([]string, error) {
	var requiredImports []string
	var typeReplaced bool

	json2GoTag := ta.data[path]
	if json2GoTag == nil || len(json2GoTag.SeenValues) == 0 || json2GoTag.BaseType != nil {
		return nil, nil
	}

	// If a checker is active from a previous run, prefer this checker over all others. Only if this one fails,
	// run the others
	if json2GoTag.NameOfActiveTypeAdjuster != nil {
		for _, checker := range registeredTypeCheckers {
			if checker.GetName() == *json2GoTag.NameOfActiveTypeAdjuster {
				runCheckerState, imp, err := ta.runChecker(checker, json2GoTag, path, e, typeReplaced)
				if err != nil {
					return nil, err
				}
				if runCheckerState == StateApplicable {
					requiredImports = append(requiredImports, imp...)
					typeReplaced = true
					break
				} else if runCheckerState == StateFailed {
					//TODO log
				}
			}
		}
	}

	// Is always run. But when typeReplaced is set. It only serves the purpose to populate the CheckedNonMatchingTypes
	// field of FieldData.
	for _, checker := range registeredTypeCheckers {
		runCheckerState, i, err := ta.runChecker(checker, json2GoTag, path, e, typeReplaced)
		if err != nil {
			return nil, err
		}
		if runCheckerState == StateApplicable {
			typeReplaced = true
			requiredImports = append(requiredImports, i...)

		} else if runCheckerState == StateUndecided {
			typeReplaced = true
			//TODO log this
		}
	}
	return requiredImports, nil
}

func (ta *TypeAdjuster) runChecker(checker TypeDeterminationFunction, fData *fieldData.FieldData, path string, e *ast.Expr, runCheckOnly bool) (State, []string, error) {
	var err error
	var requiredImports []string
	baseName := strings.ReplaceAll(path, ".", "")

	//Get input type
	var originalType string
	var exp *ast.Expr

	expr, err := utils.WalkExpressions(e)
	if err != nil {
		return StateFailed, nil, err
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

	// Ignore imperviously failed checkers
	if fData.SeenValues != nil {
		if _, ok := fData.CheckedNonMatchingTypes[checker.GetName()]; ok {
			return StateFailed, nil, nil
		}
	}

	//Init checker
	checker.SetFile(ta.file)
	err = checker.SetState(fData.TypeAdjusterData, path)
	if err != nil {
		return StateFailed, nil, err
	}

	state, err := checker.CouldTypeBeApplied(fData.SeenValues)

	if err != nil {
		var incompatibleCustomTypeError *j2gErrors.IncompatibleCustomTypeError
		if errors.As(err, &incompatibleCustomTypeError) {
			fData.Error = incompatibleCustomTypeError
		} else {
			return StateUndecided, nil, err
		}
	}

	if state == StateApplicable && !runCheckOnly {
		runCheckOnly = true
		ri, err := ta.replaceType(fData, baseName, originalType, checker, exp, requiredImports)
		if err != nil {
			return state, nil, err
		}
		requiredImports = append(requiredImports, ri...)
	} else if state == StateUndecided {
		return state, nil, nil
	} else {
		if fData.CheckedNonMatchingTypes == nil {
			fData.CheckedNonMatchingTypes = map[string]int64{}
		}
		fData.CheckedNonMatchingTypes[checker.GetName()] = time.Now().Unix()
	}
	return state, requiredImports, nil
}

// TODO look if all params are really needed
func (ta *TypeAdjuster) replaceType(json2GoTag *fieldData.FieldData, baseName string, originalType string, checker TypeDeterminationFunction, exp *ast.Expr, requiredImports []string) ([]string, error) {
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
	ta.file.Decls = append(ta.file.Decls, fromTypeFunction)
	ta.file.Decls = append(ta.file.Decls, toTypeFunction)
	*exp = checker.GetType()
	json2GoTag.BaseType = &originalType
	json2GoTag.TypeAdjusterData, err = checker.GetState()
	if err != nil {
		return nil, err
	}
	requiredImports = append(requiredImports, checker.GetRequiredImports()...)
	return requiredImports, nil
}
