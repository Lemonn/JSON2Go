package JSON2Go

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lemonn/AstUtils"
	"github.com/iancoleman/strcase"
	"go/ast"
	"go/token"
	"reflect"
	"time"
)

func GenerateCodeIntoDecl(jsonData []byte, decls []ast.Decl, structName string) ([]ast.Decl, error) {
	var JsonData interface{}
	err := json.Unmarshal(jsonData, &JsonData)
	if err != nil {
		return nil, err
	}
	wrappingStruct := &ast.StructType{
		Fields: &ast.FieldList{},
	}
	err = codeGen(nil, JsonData, &wrappingStruct.Fields.List)
	if err != nil {
		return nil, err
	}
	decls = append(decls, &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{
					Name: structName,
				},
				Type: wrappingStruct,
			},
		},
	})
	return decls, nil
}

func codeGen(fieldName *string, jsonData interface{}, fields *[]*ast.Field) error {
	var err error
	switch result := jsonData.(type) {
	case map[string]interface{}:
		err = processStruct(fieldName, result, fields)
		if err != nil {
			return err
		}
	case []interface{}:
		f, err := processSlice(fieldName, result)
		if err != nil {
			return err
		}
		*fields = append(*fields, &ast.Field{
			Names: []*ast.Ident{&ast.Ident{Name: strcase.ToCamel(*fieldName)}},
			Type:  f.Type,
			Tag:   f.Tag,
		})
	default:
		err = processField(fieldName, result, fields)
		if err != nil {
			return err
		}
	}
	return nil
}

// Processes JSON-Struct elements
func processStruct(fieldName *string, structData map[string]interface{}, fields *[]*ast.Field) error {
	var structField *ast.Field
	structType := &ast.StructType{
		Fields: &ast.FieldList{},
	}
	for n, i := range structData {
		err := codeGen(&n, i, &structType.Fields.List)
		if err != nil {
			return err
		}
	}
	// If name is not null, we need to generate a new struct, because we're processing a nested structure:
	// If name is nil, we're processing fields inside an already generated structure.
	if fieldName != nil {
		structField = &ast.Field{
			Names: []*ast.Ident{
				&ast.Ident{
					Name: strcase.ToCamel(*fieldName),
				},
			},
			Type: structType,
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "`json:\"" + *fieldName + ",omitempty\"`",
			},
		}
		*fields = append(*fields, structField)
	} else {
		*fields = append(*fields, structType.Fields.List...)
	}
	return nil
}

// Processes JSON-Array elements
func processSlice(fieldName *string, sliceData []interface{}) (*ast.Field, error) {
	var expressionList []*ast.Field
	// A JSON-Slice could have five cases:
	// 1. Contains other slices
	// 2. Contains other structs
	// 3. Is of basic type, such as int, string etc.
	// 4. An empty slice [], is not handled in te loop but later on.
	// 5. Contains a mixed set of types, in this case []interface{} is used as type.
	for _, i := range sliceData {
		fieldList := &ast.FieldList{
			List: []*ast.Field{},
		}
		switch v := i.(type) {
		case []interface{}:
			f, err := processSlice(nil, v)
			if err != nil {
				return nil, err
			}
			expressionList = append(expressionList, &ast.Field{
				Type: &ast.ArrayType{Elt: f.Type},
			})
		case map[string]interface{}:
			err := processStruct(nil, v, &fieldList.List)
			if err != nil {
				return nil, err
			}
			expressionList = append(expressionList, &ast.Field{
				Doc: nil,
				Names: func() []*ast.Ident {
					if fieldName == nil {
						return nil
					}
					return []*ast.Ident{&ast.Ident{Name: *fieldName}}
				}(),
				Type: &ast.ArrayType{Elt: &ast.StructType{Fields: fieldList}},
				Tag: func() *ast.BasicLit {
					if fieldName == nil {
						return nil
					}
					return &ast.BasicLit{
						//TODO add json2go last seen tag
						Value: "`json:\"" + *fieldName + ",omitempty\"`",
					}
				}(),
			})
		case interface{}:
			err := processField(fieldName, v, &fieldList.List)
			if err != nil {
				return nil, err
			}
			expressionList = append(expressionList, fieldList.List[0])
		}
	}

	// Add an empty []Interface{}, in case of an empty array
	if len(expressionList) == 0 {
		tagString, err := (&Tag{SeenValues: map[string]string{"": "interface{}"}, LastSeenTimestamp: time.Now().Unix()}).ToBasicLit()
		if err != nil {
			return nil, err
		}
		var j *ast.BasicLit
		if fieldName != nil {
			j = &ast.BasicLit{Value: "`json:\"" + *fieldName + ",omitempty\"`"}
		}
		tag, err := combineTags(j, tagString)
		if err != nil {
			return nil, err
		}
		expressionList = append(expressionList, &ast.Field{
			Type: &ast.ArrayType{
				Elt: &ast.InterfaceType{
					Methods: &ast.FieldList{},
				},
			},
			Tag: tag,
		})
	}

	//Check if all expressions are of equal type, if not use []interface{} as type. If equal, deep combine them into
	// one value per field.
	if !expressionsEqual(expressionList) {
		tagString, err := (&Tag{SeenValues: map[string]string{"": "interface{}"}, LastSeenTimestamp: time.Now().Unix(), MixedTypes: true}).ToBasicLit()
		if err != nil {
			return nil, err
		}
		tag, err := combineFieldTags(expressionList)
		if err != nil {
			return nil, err
		}
		tag, err = combineTags(tag, tagString)
		if err != nil {
			return nil, err
		}

		return &ast.Field{
			Names: expressionList[0].Names,
			Type: &ast.ArrayType{Elt: &ast.InterfaceType{
				Methods: &ast.FieldList{},
			}},
			Tag: tag,
		}, nil
	} else {
		switch expressionList[0].Type.(type) {
		case *ast.ArrayType:
			return combineArrays(expressionList), nil
		case *ast.StructType:
			return combineStructs(expressionList), nil
		case *ast.Ident:
			return combineFields(expressionList), nil
		case *ast.InterfaceType:
			return expressionList[0], nil
		}

	}
	return nil, errors.New(fmt.Sprintf("unkown element type: %s", reflect.TypeOf(expressionList[0].Type).String()))
}

func expressionsEqual(expressions []*ast.Field) bool {
	foundTypesMap := make(map[string]struct{})
	foundFieldTypes := make(map[string]struct{})
	for _, expr := range expressions {
		foundTypesMap[reflect.TypeOf(expr.Type).String()] = struct{}{}
		// Check the field types
		// Only check for *ast.Ident, because other types such as *ast.StarExpr should not occur in the generated code
		if v, ok := expr.Type.(*ast.Ident); ok {
			foundFieldTypes[v.Name] = struct{}{}
		}
	}
	if len(foundTypesMap) > 1 || len(foundFieldTypes) > 1 {
		return false
	}
	return true
}

func combineArrays(arrays []*ast.Field) *ast.Field {
	var elements []*ast.Field
	for _, array := range arrays {
		elements = append(elements, &ast.Field{
			Names: array.Names,
			Type:  array.Type.(*ast.ArrayType).Elt,
			Tag:   array.Tag,
		})
	}
	if !expressionsEqual(elements) {
		//TODO handle error
		tag, err := combineFieldTags(elements)
		if err != nil {
			return nil
		}
		tag, err = SetMixedTypes(tag)
		if err != nil {
			return nil
		}
		return &ast.Field{
			Tag: tag,
			Type: &ast.ArrayType{Elt: &ast.InterfaceType{
				Methods: &ast.FieldList{},
			}},
		}
	}
	switch elements[0].Type.(type) {
	case *ast.ArrayType:
		f := combineArrays(elements)
		return &ast.Field{
			Names: elements[0].Names,
			Type:  &ast.ArrayType{Elt: f.Type},
			Tag:   f.Tag,
		}
	case *ast.StructType:
		f := combineStructs(elements)
		return &ast.Field{
			Names: elements[0].Names,
			Type:  &ast.ArrayType{Elt: f.Type},
			Tag:   f.Tag,
		}
	case *ast.Ident:
		f := combineFields(elements)
		return &ast.Field{
			Names: elements[0].Names,
			Type:  &ast.ArrayType{Elt: f.Type},
			Tag:   f.Tag,
		}
	default:
		return &ast.Field{
			Names: elements[0].Names,
			Type: &ast.ArrayType{Elt: &ast.InterfaceType{
				Methods: &ast.FieldList{},
			}},
			Tag: elements[0].Tag,
		}
	}
}

func combineStructs(structs []*ast.Field) *ast.Field {
	fields := make(map[string][]*ast.Field)
	var combinedFields []*ast.Field
	for _, str := range structs {
		for _, field := range str.Type.(*ast.StructType).Fields.List {
			if _, ok := fields[field.Names[0].Name]; !ok {
				fields[field.Names[0].Name] = []*ast.Field{}
			}
			fields[field.Names[0].Name] = append(fields[field.Names[0].Name], field)
		}
	}
	for s, exprs := range fields {
		if !expressionsEqual(exprs) {
			tag, err := combineFieldTags(exprs)
			if err != nil {
				return nil
			}
			tag, err = SetMixedTypes(tag)
			if err != nil {
				//TODO handle error
				return nil
			}
			combinedFields = append(combinedFields, &ast.Field{
				Names: []*ast.Ident{&ast.Ident{Name: s}},
				Type: &ast.InterfaceType{
					Methods: &ast.FieldList{},
				},
				Tag: tag,
			})
		} else {
			switch exprs[0].Type.(type) {
			case *ast.ArrayType:
				combinedFields = append(combinedFields, combineArrays(exprs))
			case *ast.StructType:
				combinedFields = append(combinedFields, combineStructs(exprs))
			case *ast.Ident:
				combinedFields = append(combinedFields, combineFields(exprs))
			case *ast.InterfaceType:
				combinedFields = append(combinedFields, &ast.Field{
					Names: []*ast.Ident{&ast.Ident{Name: s}},
					Type:  exprs[0].Type})
			}
		}
	}

	return &ast.Field{
		Names: structs[0].Names,
		Type: &ast.StructType{
			Fields: &ast.FieldList{List: combinedFields},
		},
		Tag: structs[0].Tag,
	}
}

func combineFields(fields []*ast.Field) *ast.Field {
	if !expressionsEqual(fields) {
		tag, err := combineFieldTags(fields)
		if err != nil {
			return nil
		}
		tag, err = SetMixedTypes(tag)
		if err != nil {
			//TODO handle error
			return nil
		}

		return &ast.Field{
			Type: &ast.InterfaceType{
				Methods: &ast.FieldList{},
			},
			Tag: tag,
		}
	}
	var err error
	tag := fields[0].Tag
	for i := 1; i < len(fields); i++ {
		tag, err = combineTags(tag, fields[i].Tag)
		if err != nil {
			panic(err)
		}
	}

	//TODO really combine the fields

	return &ast.Field{
		Doc:     nil,
		Names:   fields[0].Names,
		Type:    fields[0].Type,
		Tag:     tag,
		Comment: nil,
	}
}

func processField(fieldName *string, fieldData interface{}, fields *[]*ast.Field) error {
	json2GoTagString, err := newTagFromFieldData(fieldData).ToTagString()
	if err != nil {
		return err
	}
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
		Tag: &ast.BasicLit{
			Kind: token.STRING,
			Value: func() string {
				tagString := "`" + json2GoTagString
				if fieldName != nil {
					tagString += " json:\"" + *fieldName + "\""
				}
				tagString += "`"
				return tagString
			}(),
		},
	})
	return nil
}

func combineStructFields(oldElement, newElement *ast.StructType) ([]*ast.Field, error) {
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
		json2go0, err := GetJson2GoTagFromBasicLit(exprs[0].Tag)
		if err != nil {
			return nil, err
		}
		json2go1, err := GetJson2GoTagFromBasicLit(exprs[1].Tag)
		if err != nil {
			return nil, err
		}

		//TODO check if to incompatible types result in the new type interface{}
		tag, err := combineTags(exprs[0].Tag, exprs[1].Tag)
		if err != nil {
			return nil, err
		}

		expr0 := exprs[0].Type
		expr1 := exprs[1].Type

		fmt.Println(reflect.TypeOf(expr0), reflect.TypeOf(expr1))
		var level int
		//Traverse arrays
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

		//Reset to base types
		if json2go0.BaseType != nil {
			expr0 = &ast.Ident{
				Name: *json2go0.BaseType,
			}
		}

		if json2go1.BaseType != nil {
			expr1 = &ast.Ident{
				Name: *json2go1.BaseType,
			}
		}

		tag, err = deleteTypeAdjusterValues(tag)
		if err != nil {
			return nil, err
		}

		var finalExpr ast.Expr
		if expr0 == expr1 {
			finalExpr = expr0
			for range level {
				finalExpr = &ast.ArrayType{Elt: finalExpr}
			}
			combinedFields = append(combinedFields, &ast.Field{
				Doc:     exprs[0].Doc,
				Names:   exprs[0].Names,
				Type:    finalExpr,
				Tag:     tag,
				Comment: exprs[0].Comment,
			})
		} else if _, ok := expr0.(*ast.InterfaceType); ok {
			if json2go0.MixedTypes {
				finalExpr = &ast.InterfaceType{Methods: &ast.FieldList{}}
				for range level {
					finalExpr = &ast.ArrayType{Elt: finalExpr}
				}
				combinedFields = append(combinedFields, &ast.Field{
					Doc:     exprs[0].Doc,
					Names:   exprs[0].Names,
					Type:    finalExpr,
					Tag:     tag,
					Comment: exprs[0].Comment,
				})
			} else {
				finalExpr = expr1
				for range level {
					finalExpr = &ast.ArrayType{Elt: finalExpr}
				}
				combinedFields = append(combinedFields, &ast.Field{
					Doc:     exprs[0].Doc,
					Names:   exprs[0].Names,
					Type:    finalExpr,
					Tag:     tag,
					Comment: exprs[0].Comment,
				})
			}

		} else if _, ok := expr1.(*ast.InterfaceType); ok {
			if json2go1.MixedTypes {
				finalExpr = &ast.InterfaceType{Methods: &ast.FieldList{}}
				for range level {
					finalExpr = &ast.ArrayType{Elt: finalExpr}
				}
				combinedFields = append(combinedFields, &ast.Field{
					Doc:     exprs[0].Doc,
					Names:   exprs[0].Names,
					Type:    finalExpr,
					Tag:     tag,
					Comment: exprs[0].Comment,
				})
			} else {
				finalExpr = expr0
				for range level {
					finalExpr = &ast.ArrayType{Elt: finalExpr}
				}
				combinedFields = append(combinedFields, &ast.Field{
					Doc:     exprs[0].Doc,
					Names:   exprs[0].Names,
					Type:    finalExpr,
					Tag:     tag,
					Comment: exprs[0].Comment,
				})
			}
		} else {
			finalExpr = &ast.InterfaceType{Methods: &ast.FieldList{}}
			for range level {
				finalExpr = &ast.ArrayType{Elt: finalExpr}
			}
			combinedFields = append(combinedFields, &ast.Field{
				Doc:     exprs[0].Doc,
				Names:   exprs[0].Names,
				Type:    finalExpr,
				Tag:     tag,
				Comment: exprs[0].Comment,
			})
		}
	}
	return combinedFields, nil
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
			fields, err = combineStructFields(structs[0], structs[1])
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
