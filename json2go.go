package JSON2Go

import (
	"encoding/json"
	"go/ast"
	"go/token"
	"reflect"
)

func GenerateCodeIntoDecl(jsonString string, decls []ast.Decl, structName string) ([]ast.Decl, error) {
	var JsonData interface{}
	err := json.Unmarshal([]byte(jsonString), &JsonData)
	if err != nil {
		return nil, err
	}
	wrappingStruct := &ast.StructType{
		Fields: &ast.FieldList{},
	}
	codeGen(nil, JsonData, &wrappingStruct.Fields.List)
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

func codeGen(fieldName *string, jsonData interface{}, fields *[]*ast.Field) interface{} {
	switch result := jsonData.(type) {
	case map[string]interface{}:
		processStruct(fieldName, result, fields)
	case []interface{}:

		processSlice(fieldName, result, fields)
	default:
		processField(fieldName, result, fields)
	}
	return nil
}

// Processes JSON-Struct elements
func processStruct(fieldName *string, structData map[string]interface{}, fields *[]*ast.Field) {
	var structField *ast.Field
	structType := &ast.StructType{
		Fields: &ast.FieldList{},
	}
	for n, i := range structData {
		codeGen(&n, i, &structType.Fields.List)

	}

	// If name is not null, we need to generate a new struct, because we're processing a nested structure:
	// If name is nil, we're processing fields inside an already generated structure.
	if fieldName != nil {
		structField = &ast.Field{
			Names: []*ast.Ident{
				&ast.Ident{
					Name: SetExported(*name),
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
}

// Processes JSON-Array elements
func processSlice(fieldName *string, sliceData []interface{}, fields *[]*ast.Field) {
	internalFields := &ast.FieldList{
		List: []*ast.Field{},
	}
	for _, i := range sliceData {
		codeGen(fieldName, i, &internalFields.List)
	}
	foundFields := combineDuplicateFields(internalFields.List)

	if len(foundFields) > 0 {
		gen := &ast.Field{
			Names: []*ast.Ident{
				&ast.Ident{
					Name: internalFields.List[0].Names[0].Name,
				},
			},
			Type: &ast.ArrayType{
				Elt: &ast.StructType{
					Fields: &ast.FieldList{
						List: func() []*ast.Field {
							var ff []*ast.Field
							for name, f2 := range foundFields {
								ff = append(ff, &ast.Field{
									Names: []*ast.Ident{
										&ast.Ident{
											Name: SetExported(name),
										},
									},
									Type: f2,
									Tag: &ast.BasicLit{
										Kind:  token.STRING,
										Value: "`json:\"" + name + "\"`",
									},
								})
							}
							return ff
						}(),
					},
				},
			},
		}
		*fields = append(*fields, gen)
	} else {
		gen := &ast.Field{
			Names: []*ast.Ident{
				&ast.Ident{
					Name: SetExported(*name),
				},
			},
			Type: &ast.ArrayType{
				Elt: &ast.InterfaceType{
					Methods: &ast.FieldList{},
				},
			},
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "`json:\"" + *fieldName + "\"`",
			},
		}
		*fields = append(*fields, gen)
	}
}

func processField(fieldName *string, fieldData interface{}, fields *[]*ast.Field) {
	*fields = append(*fields, &ast.Field{
		Names: []*ast.Ident{
			&ast.Ident{
				Name: SetExported(*name),
			},
		},
		Type: &ast.Ident{
			Name: reflect.TypeOf(fieldData).String(),
		},
		Tag: &ast.BasicLit{
			Kind:  token.STRING,
			Value: "`json:\"" + *fieldName + "\"`",
		},
	})
}

func combineStructFields(oldElement, newElement *ast.StructType) []*ast.Field {
	var combinedFields []*ast.Field
	oldFields := map[string]*ast.Field{}
	for _, newField := range oldElement.Fields.List {
		oldFields[newField.Names[0].Name] = newField
	}
	for _, newField := range newElement.Fields.List {

		if _, ok := oldFields[newField.Names[0].Name]; !ok {
			oldFields[newField.Names[0].Name] = newField
		} else if _, ok := newField.Type.(*ast.InterfaceType); !ok {
			oldFields[newField.Names[0].Name] = newField
		}
	}
	for _, f := range oldFields {
		combinedFields = append(combinedFields, f)
	}
	return combinedFields
}

func combineDuplicateFields(fields []*ast.Field) map[string]ast.Expr {
	foundFields := map[string]ast.Expr{}
	for _, structField := range fields {
		if structType, ok := structField.Type.(*ast.StructType); ok {
			for _, structTypeField := range structType.Fields.List {
				if _, ok := foundFields[structTypeField.Names[0].Name]; !ok {
					foundFields[structTypeField.Names[0].Name] = structTypeField.Type
				} else if str, ok := structTypeField.Type.(*ast.StructType); ok {
					c := combineStructFields(foundFields[structTypeField.Names[0].Name].(*ast.StructType), str)
					foundFields[structTypeField.Names[0].Name].(*ast.StructType).Fields.List = c
				} else if str, ok := structTypeField.Type.(*ast.ArrayType); ok {
					if _, ok := str.Elt.(*ast.StructType); !ok {
						continue
					} else {
						c := combineStructFields(foundFields[structTypeField.Names[0].Name].(*ast.ArrayType).Elt.(*ast.StructType), str.Elt.(*ast.StructType))
						foundFields[structTypeField.Names[0].Name].(*ast.ArrayType).Elt.(*ast.StructType).Fields.List = c
					}
				} else {
					if _, ok := structTypeField.Type.(*ast.InterfaceType); !ok {
						foundFields[structTypeField.Names[0].Name] = structTypeField.Type
					}
				}
			}
		}
	}
	return foundFields
}
