package pkg

import (
	"encoding/json"
	"github.com/iancoleman/strcase"
	"go/ast"
	"go/token"
	"reflect"
	"strings"
	"time"
)

var Tags = map[string]*Tag{}

func GenerateCodeIntoDecl(jsonData []byte, decls []ast.Decl, structName string) ([]ast.Decl, error) {
	var JsonData interface{}
	err := json.Unmarshal(jsonData, &JsonData)
	if err != nil {
		return nil, err
	}
	wrappingStruct := &ast.StructType{
		Fields: &ast.FieldList{},
	}
	err = codeGen(nil, JsonData, &wrappingStruct.Fields.List, structName)
	if err != nil {
		return nil, err
	}
	if len(wrappingStruct.Fields.List) == 1 && wrappingStruct.Fields.List[0].Names == nil || wrappingStruct.Fields.List[0].Names[0] == nil {
		expressions, err := walkExpressions(&wrappingStruct.Fields.List[0].Type)
		if err != nil {
			return nil, err
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
	return decls, nil
}

func codeGen(fieldName *string, jsonData interface{}, fields *[]*ast.Field, path string) error {
	var err error

	switch result := jsonData.(type) {
	case map[string]interface{}:
		err = processStruct(fieldName, result, fields, path)
		if err != nil {
			return err
		}
	case []interface{}:
		var f *ast.Field
		f, err = processSlice(fieldName, result, path)
		if err != nil {
			return err
		}
		*fields = append(*fields, &ast.Field{
			Doc:     f.Doc,
			Names:   f.Names,
			Type:    f.Type,
			Tag:     f.Tag,
			Comment: nil,
		})
	default:
		err = processField(fieldName, result, fields, path)
		if err != nil {
			return err
		}
	}
	return nil
}

// Processes JSON-Struct elements
func processStruct(fieldName *string, structData map[string]interface{}, fields *[]*ast.Field, path string) error {
	var localFields []*ast.Field
	var err error
	if fieldName != nil {
		path += "." + strcase.ToCamel(*fieldName)
	}
	for n, i := range structData {
		err = codeGen(&n, i, &localFields, path)
		if err != nil {
			return err
		}
	}

	if _, ok := Tags[path]; ok {
		combine, err := Tags[path].Combine(&Tag{LastSeenTimestamp: time.Now().Unix()})
		if err != nil {
			return err
		}
		Tags[path] = combine
	} else {
		Tags[path] = &Tag{LastSeenTimestamp: time.Now().Unix()}
	}

	// If name is not null, we need to generate a new struct, because we're processing a nested structure:
	// If name is nil, we're processing fields inside an already generated structure.
	if fieldName != nil {
		structField := &ast.Field{
			Names: []*ast.Ident{
				&ast.Ident{
					Name: strcase.ToCamel(*fieldName),
				},
			},
			Type: &ast.StructType{Fields: &ast.FieldList{List: localFields}},
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "`json:\"" + *fieldName + ",omitempty\" `",
			},
		}
		*fields = append(*fields, structField)
	} else {
		*fields = append(*fields, localFields...)
	}
	return nil
}

// Processes JSON-Array elements
func processSlice(fieldName *string, sliceData []interface{}, path string) (*ast.Field, error) {
	var expressionList []*ast.Field
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
		fieldList := &ast.FieldList{
			List: []*ast.Field{},
		}

		if fieldName != nil {
			v := strings.Split(path, ".")
			if v[len(v)-1] != strcase.ToCamel(*fieldName) {
				path += "." + strcase.ToCamel(*fieldName)
			}
		}

		switch v := i.(type) {
		case []interface{}:
			var f *ast.Field
			f, err = processSlice(nil, v, path)
			if err != nil {
				return nil, err
			}
			tag := f.Tag
			if fieldName != nil {
				tag, err = combineTags(&(ast.BasicLit{Kind: token.STRING, Value: "`json:\"" + *fieldName + ",omitempty\" `"}), f.Tag)
				if err != nil {
					return nil, err
				}
			}

			expressionList = append(expressionList, &ast.Field{
				Names: func() []*ast.Ident {
					if fieldName == nil {
						return nil
					}
					return []*ast.Ident{&ast.Ident{Name: strcase.ToCamel(*fieldName)}}
				}(),
				Type: &ast.ArrayType{Elt: f.Type},
				Tag:  tag,
			})
		case map[string]interface{}:
			err = processStruct(nil, v, &fieldList.List, path)
			if err != nil {
				return nil, err
			}
			expressionList = append(expressionList, &ast.Field{
				Doc: nil,
				Names: func() []*ast.Ident {
					if fieldName == nil {
						return nil
					}
					return []*ast.Ident{&ast.Ident{Name: strcase.ToCamel(*fieldName)}}
				}(),
				Type: &ast.ArrayType{Elt: &ast.StructType{Fields: fieldList}},
				Tag: func() *ast.BasicLit {
					if fieldName == nil {
						return nil
					}
					return &ast.BasicLit{
						Kind: token.STRING,
						//TODO add json2go last seen tag
						Value: "`json:\"" + *fieldName + ",omitempty\"`",
					}
				}(),
			})
		case interface{}:
			err = processField(fieldName, v, &fieldList.List, path)
			if err != nil {
				return nil, err
			}
			expressionList = append(expressionList, &ast.Field{
				Doc:     fieldList.List[0].Doc,
				Names:   fieldList.List[0].Names,
				Type:    &ast.ArrayType{Elt: fieldList.List[0].Type},
				Tag:     fieldList.List[0].Tag,
				Comment: fieldList.List[0].Comment,
			})
		}
	}

	// Add an empty []Interface{}, in case of an empty array
	if len(expressionList) == 0 {
		tagString, err := (&Tag{EmptyValuePresent: true, LastSeenTimestamp: time.Now().Unix()}).ToBasicLit()
		if err != nil {
			return nil, err
		}
		var j *ast.BasicLit
		if fieldName != nil {
			j = &ast.BasicLit{Value: "`json:\"" + *fieldName + ",omitempty\" `"}
		}
		tag, err := combineTags(j, tagString)
		if err != nil {
			return nil, err
		}
		expressionList = append(expressionList, &ast.Field{
			Type: &ast.ArrayType{Elt: &ast.InterfaceType{
				Methods: &ast.FieldList{}},
			},
			Tag: tag,
			Names: func() []*ast.Ident {
				if fieldName == nil {
					return nil
				}
				return []*ast.Ident{{Name: strcase.ToCamel(*fieldName)}}
			}(),
		})
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

func processField(fieldName *string, fieldData interface{}, fields *[]*ast.Field, path string) error {
	/*json2GoTagString, err := newTagFromFieldData(fieldData).ToTagString()
	if err != nil {
		return err
	}

	*/
	*fields = append(*fields, &ast.Field{
		Names: []*ast.Ident{
			func() *ast.Ident {
				if fieldName == nil {
					return nil
				}
				return &ast.Ident{
					Name: strcase.ToCamel(*fieldName),
				}
			}(),
		},
		Type: &ast.Ident{
			Name: reflect.TypeOf(fieldData).String(),
		},
	})
	if fieldName != nil {
		v := strings.Split(path, ".")
		if v[len(v)-1] != strcase.ToCamel(*fieldName) {
			path += "." + strcase.ToCamel(*fieldName)
		}
	}
	if _, ok := Tags[path]; ok {
		combine, err := Tags[path].Combine(newTagFromFieldData(fieldData, fieldName))
		if err != nil {
			return err
		}
		Tags[path] = combine
	} else {
		Tags[path] = newTagFromFieldData(fieldData, fieldName)
	}
	return nil
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
		field, err := combineFields(exprs[0], exprs[1], path+"."+exprs[0].Names[0].Name)
		if err != nil {
			return nil, err
		}
		combinedFields = append(combinedFields, field)
	}
	return combinedFields, nil
}

// TODO set conflicting field json2go tag value
func combineFields(field0, field1 *ast.Field, path string) (*ast.Field, error) {
	//Get Tags of both fields

	json2go0 := Tags[path]
	json2go1 := Tags[path]
	/*
		json2go0, err := GetJson2GoTagFromBasicLit(field0.Tag)
		if err != nil {
			return nil, err
		}
		json2go1, err := GetJson2GoTagFromBasicLit(field1.Tag)
		if err != nil {
			return nil, err
		}


		//Combine both tags into one
		tag, err := combineTags(field0.Tag, field1.Tag)
		if err != nil {
			return nil, err
		}

	*/

	//Traverse as long as both fields are of type *ast.Array
	expr0 := field0.Type
	expr1 := field1.Type
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

	/*
		//Reset to base types
		resetToBaseType(&expr0, json2go0)
		resetToBaseType(&expr1, json2go1)

		// Delete potentially present TypeAdjusterValues, as the TypeAdjuster needs to rerun after the merge
		tag, err = deleteTypeAdjusterValues(tag)
		if err != nil {
			return nil, err
		}

	*/

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
			var l string
			if field0.Names != nil && len(field1.Names) > 0 {
				v := strings.Split(path, ".")
				if v[len(v)-1] != strcase.ToCamel(field0.Names[0].Name) {
					path += "." + strcase.ToCamel(field0.Names[0].Name)
				} else {
					l = path
				}

			} else {
				l = path
			}
			fields, err := combineStructFields(expr0.(*ast.StructType), expr1.(*ast.StructType), l)
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

		return &ast.Field{
			Doc:   field0.Doc,
			Names: field0.Names,
			Type:  finalExpr,
			//Tag:     tag,
			Comment: field0.Comment,
		}, nil
	} else if _, ok := expr0.(*ast.InterfaceType); ok {
		if json2go0 != nil && json2go0.MixedTypes {
			finalExpr = &ast.InterfaceType{Methods: &ast.FieldList{}}
			for range level {
				finalExpr = &ast.ArrayType{Elt: finalExpr}
			}
			return &ast.Field{
				Doc:   field0.Doc,
				Names: field0.Names,
				Type:  finalExpr,
				//Tag:     tag,
				Comment: field0.Comment,
			}, nil
		} else {
			finalExpr = expr1
			for range level {
				finalExpr = &ast.ArrayType{Elt: finalExpr}
			}
			return &ast.Field{
				Doc:   field0.Doc,
				Names: field0.Names,
				Type:  finalExpr,
				//Tag:     tag,
				Comment: field0.Comment,
			}, nil
		}

	} else if _, ok := expr1.(*ast.InterfaceType); ok {
		if json2go1 != nil && json2go1.MixedTypes {
			finalExpr = &ast.InterfaceType{Methods: &ast.FieldList{}}
			for range level {
				finalExpr = &ast.ArrayType{Elt: finalExpr}
			}
			return &ast.Field{
				Doc:   field0.Doc,
				Names: field0.Names,
				Type:  finalExpr,
				//Tag:     tag,
				Comment: field0.Comment,
			}, nil
		} else {
			finalExpr = expr0
			for range level {
				finalExpr = &ast.ArrayType{Elt: finalExpr}
			}
			return &ast.Field{
				Doc:   field0.Doc,
				Names: field0.Names,
				Type:  finalExpr,
				//Tag:     tag,
				Comment: field0.Comment,
			}, nil
		}
	} else {
		finalExpr = &ast.InterfaceType{Methods: &ast.FieldList{}}
		for range level {
			finalExpr = &ast.ArrayType{Elt: finalExpr}
		}

		Tags[path].MixedTypes = true
		/*
			tag, err = (&Tag{MixedTypes: true}).AppendToTag(tag)
			if err != nil {
				return nil, err
			}

		*/
		return &ast.Field{
			Doc:   field0.Doc,
			Names: field0.Names,
			Type:  finalExpr,
			//Tag:     tag,
			Comment: field0.Comment,
		}, nil
	}
}

func resetToBaseType(expr *ast.Expr, json2go *Tag) {
	if json2go != nil && json2go.BaseType != nil {
		if *json2go.BaseType == "interface{}" {
			*expr = &ast.InterfaceType{Methods: &ast.FieldList{}}
		}
		*expr = &ast.Ident{
			Name: *json2go.BaseType,
		}
	}
}

func shouldIgnoreStruct(st *ast.StructType) (bool, error) {
	for _, field := range st.Fields.List {
		lit, err := GetJson2GoTagFromBasicLit(field.Tag)
		if err != nil {
			return true, err
		}
		if lit == nil {
			return true, nil
		}
	}
	return false, nil
}

func RenamePaths() {
	for s, tag := range Tags {
		pathElements := strings.Split(s, ".")
		if len(pathElements) > 1 && s != pathElements[len(pathElements)-2]+"."+pathElements[len(pathElements)-1] {
			Tags[pathElements[len(pathElements)-2]+"."+pathElements[len(pathElements)-1]] = tag
			delete(Tags, s)
		}

	}
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
