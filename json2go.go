package JSON2Go

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/Lemonn/AstUtils"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
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
		err = processSlice(fieldName, result, fields)
		if err != nil {
			return err
		}
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
					Name: AstUtils.SetExported(*fieldName),
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
func processSlice(fieldName *string, sliceData []interface{}, field *[]*ast.Field) error {
	fields := &ast.FieldList{
		List: []*ast.Field{},
	}
	for _, i := range sliceData {
		err := codeGen(fieldName, i, &fields.List)
		if err != nil {
			return err
		}
	}
	foundFields, err := combineDuplicateFields(fields.List)
	if err != nil {
		return err
	}

	if len(foundFields) > 0 {
		gen := &ast.Field{
			Names: []*ast.Ident{
				&ast.Ident{
					Name: fields.List[0].Names[0].Name,
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
											Name: AstUtils.SetExported(name),
										},
									},
									Type: f2.Type,
									Tag:  f2.Tag,
								})
							}
							return ff
						}(),
					},
				},
			},
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "`json:\"" + *fieldName + ",omitempty\"`",
			},
		}
		*field = append(*field, gen)
	} else {
		gen := &ast.Field{
			Names: []*ast.Ident{
				&ast.Ident{
					Name: AstUtils.SetExported(*fieldName),
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
		*field = append(*field, gen)
	}
	return nil
}

func processField(fieldName *string, fieldData interface{}, fields *[]*ast.Field) error {
	jsonTag, err := func() (string, error) {
		v := func() string {
			switch t := fieldData.(type) {
			case float64:
				return strconv.FormatFloat(t, 'f', -1, 64)
			case bool:
				if t {
					return "true"
				}
				return "false"
			default:
				return fieldData.(string)
			}
		}()

		var r []byte
		r, _ = json.Marshal(Tag{
			SeenValues: []string{v},
		})
		b64 := bytes.NewBuffer([]byte{})
		raw := base64.NewEncoder(base64.StdEncoding, b64)
		_, err := raw.Write(r)
		if err != nil {
			return "", err
		}
		err = raw.Close()
		return b64.String(), nil
	}()
	if err != nil {
		return err
	}
	*fields = append(*fields, &ast.Field{
		Names: []*ast.Ident{
			&ast.Ident{
				Name: AstUtils.SetExported(*fieldName),
			},
		},
		Type: &ast.Ident{
			Name: reflect.TypeOf(fieldData).String(),
		},
		Tag: &ast.BasicLit{
			Kind:  token.STRING,
			Value: "`json:\"" + *fieldName + "\" json2go:\"" + jsonTag + "\"`",
		},
	})
	return nil
}

func combineStructFields(oldElement, newElement *ast.StructType) ([]*ast.Field, error) {
	var combinedFields []*ast.Field
	oldFields := map[string]*ast.Field{}
	for _, newField := range oldElement.Fields.List {
		oldFields[newField.Names[0].Name] = newField
	}
	for _, newField := range newElement.Fields.List {
		if _, ok := oldFields[newField.Names[0].Name]; !ok {
			oldFields[newField.Names[0].Name] = newField
		} else if _, ok := newField.Type.(*ast.InterfaceType); !ok {
			combinedTags, err := combineTags(oldFields[newField.Names[0].Name].Tag, newField.Tag)
			if err != nil {
				return nil, err
			}
			oldFields[newField.Names[0].Name] = &ast.Field{
				Doc:     newField.Doc,
				Names:   newField.Names,
				Type:    newField.Type,
				Tag:     combinedTags,
				Comment: newField.Comment,
			}
		}
	}
	for _, f := range oldFields {
		combinedFields = append(combinedFields, f)
	}
	return combinedFields, nil
}

func combineDuplicateFields(fields []*ast.Field) (map[string]*ast.Field, error) {
	foundFields := map[string]*ast.Field{}
	for _, structField := range fields {
		if structType, ok := structField.Type.(*ast.StructType); ok {
			for _, structTypeField := range structType.Fields.List {
				if _, ok := foundFields[structTypeField.Names[0].Name]; !ok {
					foundFields[structTypeField.Names[0].Name] = structTypeField
				} else if str, ok := structTypeField.Type.(*ast.StructType); ok {
					c, err := combineStructFields(foundFields[structTypeField.Names[0].Name].Type.(*ast.StructType), str)
					if err != nil {
						return nil, err
					}
					foundFields[structTypeField.Names[0].Name].Type.(*ast.StructType).Fields.List = c
				} else if str, ok := structTypeField.Type.(*ast.ArrayType); ok {
					if _, ok := str.Elt.(*ast.StructType); !ok {
						continue
					} else {
						c, err := combineStructFields(foundFields[structTypeField.Names[0].Name].Type.(*ast.ArrayType).Elt.(*ast.StructType), str.Elt.(*ast.StructType))
						if err != nil {
							return nil, err
						}
						foundFields[structTypeField.Names[0].Name].Type.(*ast.ArrayType).Elt.(*ast.StructType).Fields.List = c
					}
				} else {
					if _, ok := structTypeField.Type.(*ast.InterfaceType); !ok {
						foundFields[structTypeField.Names[0].Name] = structTypeField
					}
				}
			}
		}
	}
	return foundFields, nil
}
