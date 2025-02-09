package structGenerator

import (
	"encoding/json"
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/iancoleman/strcase"
	"go/token"

	//"github.com/iancoleman/strcase"
	"go/ast"
	"reflect"
	"strings"
)

var Tags = map[string]*fieldData.Data{}

func GenerateCodeIntoDecl(jsonData []byte, decls []ast.Decl, structName string) ([]ast.Decl, map[string]*fieldData.Data, error) {
	var JsonData interface{}
	err := json.Unmarshal(jsonData, &JsonData)
	if err != nil {
		return nil, nil, err
	}
	/*
		wrappingStruct := &ast.StructType{
			Fields: &ast.FieldList{},
		}

	*/
	fields, err := codeGen(JsonData, structName)
	if err != nil {
		return nil, nil, err
	}

	/*
		if len(wrappingStruct.Fields.List) == 1 && wrappingStruct.Fields.List[0].Names == nil || wrappingStruct.Fields.List[0].Names[0] == nil {
			expressions, err := utils.WalkExpressions(&wrappingStruct.Fields.List[0].Type)
			if err != nil {
				return nil, nil, err
			}
			if _, ok := (*expressions).(*ast.StructType); !ok {
				decls = append(decls, &ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: structName,
							},
							Type:    wrappingStruct.Fields.List[0].Type,
							Comment: wrappingStruct.Fields.List[0].Comment,
						},
					},
				})
			} else {
				decls = append(decls, &ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: structName + "AnonymousArrayType",
							},
							Type: &ast.StructType{
								Fields: &ast.FieldList{
									List: (*expressions).(*ast.StructType).Fields.List,
								},
							},
						},
					},
				})

				*expressions = &ast.StarExpr{X: &ast.Ident{Name: structName + "AnonymousArrayType"}}
				decls = append(decls, &ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: structName,
							},
							Type:    wrappingStruct.Fields.List[0].Type,
							Comment: wrappingStruct.Fields.List[0].Comment,
						},
					},
				})

				for s, tag := range Tags {
					t := strings.Split(s, ".")
					var r string
					for _, s2 := range t {
						if r != "" {
							r += "."
						}
						if s2 == structName {
							r += structName + "AnonymousArrayType"
						} else {
							r += s2
						}
					}
					Tags[r] = tag
					if r != s {
						delete(Tags, s)
					}

				}

			}
		} else {
			decls = append(decls, &ast.GenDecl{
				Tok: token.TYPE,
				Specs: []ast.Spec{
					&ast.TypeSpec{
						Name: &ast.Ident{
							Name: structName,
						},
						Type:    wrappingStruct,
						Comment: wrappingStruct.Fields.List[0].Comment,
					},
				},
			})
		}

	*/

	if len(fields) == 1 {
		expr, levelOfArrays, err := utils.WalkExpressionsWhitArrayCount(&fields[0].Type)
		if err != nil {
			return nil, nil, err
		}

		decls = append(decls, &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: ast.NewIdent(structName),
					Type: func() ast.Expr {
						if levelOfArrays == 0 {
							return *expr
						}
						var oe ast.Expr
						var ie *ast.Expr
						ie, oe = utils.GeneratedNestedArray(levelOfArrays, ie, oe)
						(*ie).(*ast.ArrayType).Elt = *expr
						return oe
					}(),
				},
			},
		})

		return decls, Tags, nil

	}
	fmt.Println("sdfsdfas")
	decls = append(decls, &ast.GenDecl{Tok: token.TYPE, Specs: []ast.Spec{&ast.TypeSpec{Name: &ast.Ident{Name: "Test"}, Type: &ast.StructType{Fields: &ast.FieldList{List: fields}}}}})

	return decls, Tags, nil
}

var _ = &ast.File{
	Package: 1,
	Name: &ast.Ident{
		Name: "main",
	},
	Decls: []ast.Decl{
		&ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{
						Name: "test1234",
					},
					Type: &ast.Ident{
						Name: "int",
					},
				},
			},
		},
	},
}

func codeGen(jsonData interface{}, path string) ([]*ast.Field, error) {
	var fields []*ast.Field

	switch result := jsonData.(type) {
	case map[string]interface{}:
		str, err := processStruct(result, path)
		if err != nil {
			return nil, err
		}
		fields = append(fields, &ast.Field{
			Names: getFieldNamesFromPath(path),
			Type:  str,
		})
	case []interface{}:
		slice, err := processSlice(result, path)
		if err != nil {
			return nil, err
		}
		fields = append(fields, &ast.Field{
			Names: getFieldNamesFromPath(path),
			Type:  slice,
		})
	default:
		field, err := processField(result, path)
		if err != nil {
			return nil, err
		}
		fields = append(fields, &ast.Field{
			Names: getFieldNamesFromPath(path),
			Type:  field,
		})
	}

	return fields, nil
}

func getFieldNamesFromPath(path string) []*ast.Ident {
	pathElements := strings.Split(path, ".")
	return []*ast.Ident{{Name: strcase.ToCamel(pathElements[len(pathElements)-1])}}
}

// Processes JSON-Struct elements
func processStruct(structData map[string]interface{}, path string) (*ast.StructType, error) {
	var localFields []*ast.Field
	for fieldName, field := range structData {
		f, err := codeGen(field, path+"."+fieldName)
		if err != nil {
			return nil, err
		}
		localFields = append(localFields, f...)
	}
	return &ast.StructType{Fields: &ast.FieldList{List: localFields}}, nil
}

// Processes JSON-Array elements
func processSlice(sliceData []interface{}, path string) (ast.Expr, error) {
	var expressionList []ast.Expr
	var err error
	// A JSON-Slice could have five cases:
	// 1. Contains other slices
	// 2. Contains other structs
	// 3. Is of basic type, such as int, string etc.
	// 4. An empty slice [], is not handled in the loop but later on. In this case []interface{} is used as type.
	// 5. Contains a mixed set of types, in this case []interface{} is used as type.
	// Note do distinguish between conflicting types and empty types the json2go tag is used, if conflicting fields are
	// seen, MixedTypes is set.
	for _, i := range sliceData {
		switch v := i.(type) {
		case []interface{}:
			slice, err := processSlice(v, path)
			if err != nil {
				return nil, err
			}
			expressionList = append(expressionList, &ast.ArrayType{Elt: slice})
		case map[string]interface{}:
			str, err := processStruct(v, path)
			if err != nil {
				return nil, err
			}
			expressionList = append(expressionList, &ast.ArrayType{Elt: str})
		case interface{}:
			ident, err := processField(v, path)
			if err != nil {
				return nil, err
			}
			expressionList = append(expressionList, &ast.ArrayType{Elt: ident})
		}
	}
	// Add an empty []Interface{}, in case of an empty array
	if len(expressionList) == 0 {
		//tagString, err := (&fieldData.Data{EmptyValuePresent: true, LastSeenTimestamp: time.Now().Unix()}).ToBasicLit()
		if err != nil {
			return nil, err
		}
		expressionList = append(expressionList, &ast.ArrayType{Elt: &ast.InterfaceType{
			Methods: &ast.FieldList{}},
		})
		//TODO set empty field
	}
	f := expressionList[0]
	for i := 1; i < len(expressionList); i++ {
		f, err = combineFields(f, expressionList[i], path)
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}

func processField(field interface{}, path string) (*ast.Ident, error) {
	if _, ok := Tags[path]; ok {
		combine, err := Tags[path].Combine(fieldData.NewTagFromFieldData(field))
		if err != nil {
			return nil, err
		}
		Tags[path] = combine
	} else {
		Tags[path] = fieldData.NewTagFromFieldData(field)
	}

	return &ast.Ident{
		Name: reflect.TypeOf(field).String(),
	}, nil

}

func combineStructFields(oldElement, newElement *ast.StructType, path string) ([]*ast.Field, error) {
	var combinedFields []*ast.Field
	fields := map[string][]*ast.Field{}

	for _, oldField := range oldElement.Fields.List {
		if _, ok := fields[oldField.Names[0].Name]; !ok {
			fields[oldField.Names[0].Name] = []*ast.Field{}
		}
		fields[oldField.Names[0].Name] = append(fields[oldField.Names[0].Name], oldField)
	}

	for _, newField := range newElement.Fields.List {
		if _, ok := fields[newField.Names[0].Name]; !ok {
			fields[newField.Names[0].Name] = []*ast.Field{}
		}
		fields[newField.Names[0].Name] = append(fields[newField.Names[0].Name], newField)
	}

	for _, exprs := range fields {
		if len(exprs) == 1 {
			combinedFields = append(combinedFields, exprs[0])
			continue
		}
		combinedExpr, err := combineFields(exprs[0].Type, exprs[1].Type, path+"."+exprs[0].Names[0].Name)
		if err != nil {
			return nil, err
		}
		combinedFields = append(combinedFields, &ast.Field{
			Doc:     exprs[0].Doc,
			Names:   exprs[0].Names,
			Type:    combinedExpr,
			Tag:     exprs[0].Tag,
			Comment: exprs[0].Comment,
		})
	}
	return combinedFields, nil
}

// TODO set conflicting field json2go tag value
func combineFields(expr0, expr1 ast.Expr, path string) (ast.Expr, error) {
	//Get Tags of both fields
	json2go0 := Tags[path]
	json2go1 := Tags[path]

	//Traverse as long as both fields are of type *ast.Array
	var level int
	for reflect.TypeOf(expr0) == reflect.TypeOf(expr1) {
		if _, ok := expr0.(*ast.ArrayType); ok {
			if _, ok := expr1.(*ast.ArrayType); ok {
				level++
				expr0 = expr0.(*ast.ArrayType).Elt
				expr1 = expr1.(*ast.ArrayType).Elt
			} else {
				break
			}
		} else {
			break
		}
	}

	var finalExpr ast.Expr
	// As we traversed down all potential array layers, the only possible cases are now
	// 1. Both types are equal, in this case it's either a StructType, which needs to be combined.
	// Or an equal field type such as InterfaceType which does not need special treatment.

	// 2. One field is of interface type. In this case we look if it's a mixed type. If that's the case, we set it
	// to InterfaceType. Otherwise, it's case of a previously empty field, for which we're seen values now. This means
	// we set the field to the new values

	// 3 Fields of mixed types, that could not be combined into one, in this case set field to InterfaceType
	if reflect.TypeOf(expr0) == reflect.TypeOf(expr1) {
		finalExpr = expr0
		if _, ok := expr0.(*ast.StructType); ok {
			fields, err := combineStructFields(expr0.(*ast.StructType), expr1.(*ast.StructType), path)
			if err != nil {
				return nil, err
			}
			finalExpr = &ast.StructType{Fields: &ast.FieldList{List: fields}}
		} else if v, ok := expr0.(*ast.Ident); ok && v.Name != expr1.(*ast.Ident).Name {
			Tags[path].MixedTypes = true
			finalExpr = &ast.InterfaceType{Methods: &ast.FieldList{}}
		}

		for range level {
			finalExpr = &ast.ArrayType{Elt: finalExpr}
		}
		return finalExpr, nil
	} else if _, ok := expr0.(*ast.InterfaceType); ok {
		if json2go0 != nil && json2go0.MixedTypes {
			finalExpr = &ast.InterfaceType{Methods: &ast.FieldList{}}
			for range level {
				finalExpr = &ast.ArrayType{Elt: finalExpr}
			}
			return finalExpr, nil
		} else {
			finalExpr = expr1
			for range level {
				finalExpr = &ast.ArrayType{Elt: finalExpr}
			}
			return finalExpr, nil
		}

	} else if _, ok := expr1.(*ast.InterfaceType); ok {
		if json2go1 != nil && json2go1.MixedTypes {
			finalExpr = &ast.InterfaceType{Methods: &ast.FieldList{}}
			for range level {
				finalExpr = &ast.ArrayType{Elt: finalExpr}
			}
			return finalExpr, nil
		} else {
			finalExpr = expr0
			for range level {
				finalExpr = &ast.ArrayType{Elt: finalExpr}
			}
			return finalExpr, nil
		}
	} else {
		finalExpr = &ast.InterfaceType{Methods: &ast.FieldList{}}
		for range level {
			finalExpr = &ast.ArrayType{Elt: finalExpr}
		}
		Tags[path].MixedTypes = true
		return finalExpr, nil
	}
}

func resetToBaseType(expr *ast.Expr, json2go *fieldData.Data) {
	if json2go != nil && json2go.BaseType != nil {
		if *json2go.BaseType == "interface{}" {
			*expr = &ast.InterfaceType{Methods: &ast.FieldList{}}
		}
		*expr = &ast.Ident{
			Name: *json2go.BaseType,
		}
	}
}

func RenamePaths(tags map[string]*fieldData.Data) map[string]*fieldData.Data {
	for s, tag := range tags {
		pathElements := strings.Split(s, ".")
		if len(pathElements) > 1 {
			if tag.StructType {
				Tags[pathElements[len(pathElements)-1]] = &fieldData.Data{LastSeenTimestamp: tag.LastSeenTimestamp, StructType: true}
			}
			tag.StructType = false
			Tags[pathElements[len(pathElements)-2]+"."+pathElements[len(pathElements)-1]] = tag
			if s != pathElements[len(pathElements)-2]+"."+pathElements[len(pathElements)-1] {
				delete(Tags, s)
			}
		}
	}
	return tags
}

/*
func CombineStructsOfFile(file, file1 *ast.File) (*ast.File, error) {
	var foundNodes []*AstUtils.FoundNodes
	var completed = false
	foundStructs := map[string][]*ast.StructType{}
	AstUtils.SearchNodes(file, &foundNodes, []*ast.Node{}, func(n *ast.Node, parents []*ast.Node, completed *bool) bool {
		if _, ok := (*n).(*ast.StructType); ok && len(parents) > 0 {
			if _, ok := (*parents[0]).(*ast.TypeSpec); ok {
				return true
			}
		}
		return false
	}, &completed)

	for _, node := range foundNodes {
		typeSpec := (*node.Parents[0]).(*ast.TypeSpec)
		ignoreStruct, err := shouldIgnoreStruct(typeSpec.Type.(*ast.StructType))
		if err != nil {
			return nil, err
		}
		if ignoreStruct {
			continue
		}
		foundStructs[typeSpec.Name.Name] = []*ast.StructType{typeSpec.Type.(*ast.StructType)}
	}

	var foundNodes1 []*AstUtils.FoundNodes
	var completed1 = false
	AstUtils.SearchNodes(file1, &foundNodes1, []*ast.Node{}, func(n *ast.Node, parents []*ast.Node, completed *bool) bool {
		if _, ok := (*n).(*ast.StructType); ok {
			return true
		}
		return false
	}, &completed1)

	for _, node := range foundNodes1 {
		typeSpec := (*node.Parents[0]).(*ast.TypeSpec)
		ignoreStruct, err := shouldIgnoreStruct(typeSpec.Type.(*ast.StructType))
		if err != nil {
			return nil, err
		}
		if ignoreStruct {
			continue
		}
		if _, ok := foundStructs[typeSpec.Name.Name]; !ok {
			foundStructs[typeSpec.Name.Name] = []*ast.StructType{typeSpec.Type.(*ast.StructType)}
		} else {
			foundStructs[typeSpec.Name.Name] = append(foundStructs[typeSpec.Name.Name], typeSpec.Type.(*ast.StructType))
		}
	}

	outFile, err := AstUtils.GetEmptyFile(file.Name.Name)
	if err != nil {
		return nil, err
	}

	for name, structs := range foundStructs {
		var fields []*ast.Field
		var err error
		if len(structs) == 1 {
			fields = structs[0].Fields.List
		} else {
			fields, err = combineStructFields(structs[0], structs[1], path)
			if err != nil {
				return nil, err
			}
		}

		outFile.Decls = append(outFile.Decls, &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{&ast.TypeSpec{
				Name: &ast.Ident{
					Name: name,
				},
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: fields,
					},
				},
			}},
		})
	}
	return outFile, nil
}

*/
