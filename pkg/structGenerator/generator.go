package structGenerator

import (
	"encoding/json"
	"github.com/Lemonn/AstUtils"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/iancoleman/strcase"
	"go/ast"
	"go/token"
	"reflect"
	"strings"
	"time"
)

type StructGenerator struct {
	data map[string]*fieldData.FieldData
	file *ast.File
}

func NewCodeGenerator(data *map[string]*fieldData.FieldData) *StructGenerator {
	if data == nil {
		data = &map[string]*fieldData.FieldData{}
	}
	return &StructGenerator{
		data: *data,
	}
}

func (s *StructGenerator) GenerateCodeIntoFile(jsonData []byte, file *ast.File, structName string) (map[string]*fieldData.FieldData, error) {
	s.file = file
	var JsonData interface{}
	err := json.Unmarshal(jsonData, &JsonData)
	if err != nil {
		return nil, err
	}

	fields, err := s.codeGen(JsonData, structName)
	if err != nil {
		return nil, err
	}

	err = s.packType(fields, structName)
	if err != nil {
		return nil, err
	}

	AstUtils.UnnestStruct(nil, file)
	s.renamePaths()
	s.attachJsonTags()
	return s.data, nil
}

func (s *StructGenerator) packType(fields []*ast.Field, structName string) error {
	expr, levelOfArrays, err := utils.WalkExpressionsWhitArrayCount(&fields[0].Type)
	if err != nil {
		return err
	}
	if reflect.TypeOf(*expr) == reflect.TypeOf(&ast.StructType{}) && levelOfArrays > 0 {
		s.file.Decls = append(s.file.Decls, &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{
				&ast.TypeSpec{
					Name: &ast.Ident{
						Name: structName + "AnonymousArray",
					},
					Type: &ast.StructType{
						Fields: &ast.FieldList{
							List: (*expr).(*ast.StructType).Fields.List,
						},
					},
				},
			},
		})

		s.file.Decls = append(s.file.Decls, &ast.GenDecl{
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
						(*ie).(*ast.ArrayType).Elt = &ast.StarExpr{
							X: &ast.Ident{
								Name: structName + "AnonymousArray",
							},
						}
						return oe
					}(),
				},
			},
		})

		for path, tag := range s.data {
			t := strings.Split(path, ".")
			var r string
			for _, s2 := range t {
				if r != "" {
					r += "."
				}
				if s2 == structName {
					r += structName + "AnonymousArray"
				} else {
					r += s2
				}
			}
			s.data[r] = tag
			if r != path {
				delete(s.data, path)
			}

		}
	} else {
		s.file.Decls = append(s.file.Decls, &ast.GenDecl{
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
	}
	return nil
}

func (s *StructGenerator) codeGen(jsonData interface{}, path string) ([]*ast.Field, error) {
	var fields []*ast.Field

	switch result := jsonData.(type) {
	case map[string]interface{}:
		str, err := s.processStruct(result, path)
		if err != nil {
			return nil, err
		}
		fields = append(fields, &ast.Field{
			Names: utils.GetFieldIdentFromPath(path),
			Type:  str,
		})
	case []interface{}:
		slice, err := s.processSlice(result, path)
		if err != nil {
			return nil, err
		}
		fields = append(fields, &ast.Field{
			Names: utils.GetFieldIdentFromPath(path),
			Type:  slice,
		})
	default:
		field, err := s.processField(result, path)
		if err != nil {
			return nil, err
		}
		fields = append(fields, &ast.Field{
			Names: utils.GetFieldIdentFromPath(path),
			Type:  field,
		})
	}

	return fields, nil
}

// Processes JSON-Struct elements
func (s *StructGenerator) processStruct(structData map[string]interface{}, path string) (*ast.StructType, error) {
	err := fieldData.SetOrCombineFieldData(&fieldData.FieldData{StructType: true, LastSeenTimestamp: time.Now().Unix()}, s.data, path)
	if err != nil {
		return nil, err
	}
	var localFields []*ast.Field
	for fieldName, field := range structData {
		err := fieldData.SetOrCombineFieldData(&fieldData.FieldData{JsonFieldName: &fieldName, LastSeenTimestamp: time.Now().Unix()}, s.data, path+"."+strcase.ToCamel(fieldName))
		if err != nil {
			return nil, err
		}
		f, err := s.codeGen(field, path+"."+strcase.ToCamel(fieldName))
		if err != nil {
			return nil, err
		}
		localFields = append(localFields, f...)
	}
	return &ast.StructType{Fields: &ast.FieldList{List: localFields}}, nil
}

// Processes JSON-Array elements
func (s *StructGenerator) processSlice(sliceData []interface{}, path string) (ast.Expr, error) {
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
			slice, err := s.processSlice(v, path)
			if err != nil {
				return nil, err
			}
			expressionList = append(expressionList, &ast.ArrayType{Elt: slice})
		case map[string]interface{}:
			str, err := s.processStruct(v, path)
			if err != nil {
				return nil, err
			}
			expressionList = append(expressionList, &ast.ArrayType{Elt: str})
		case interface{}:
			ident, err := s.processField(v, path)
			if err != nil {
				return nil, err
			}
			expressionList = append(expressionList, &ast.ArrayType{Elt: ident})
		}
	}
	// Add an empty []Interface{}, in case of an empty array
	if len(expressionList) == 0 {
		expressionList = append(expressionList, &ast.ArrayType{Elt: &ast.InterfaceType{
			Methods: &ast.FieldList{}},
		})
		err = fieldData.SetOrCombineFieldData(&fieldData.FieldData{EmptyValuePresent: true}, s.data, path)
		if err != nil {
			return nil, err
		}
	}
	f := expressionList[0]
	for i := 1; i < len(expressionList); i++ {
		f, err = s.combineFields(f, expressionList[i], path)
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}

func (s *StructGenerator) processField(field interface{}, path string) (*ast.Ident, error) {
	fData, err := fieldData.NewTagFromFieldData(field)
	if err != nil {
		return nil, err
	}
	if _, ok := s.data[path]; ok {
		combine, err := s.data[path].Combine(fData)
		if err != nil {
			return nil, err
		}
		s.data[path] = combine
	} else {
		s.data[path] = fData
	}

	return &ast.Ident{
		Name: reflect.TypeOf(field).String(),
	}, nil
}
