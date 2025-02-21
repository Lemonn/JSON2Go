package structGenerator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type StructGenerator struct {
	seenTypes map[string]*fieldData.PathData
	file      *ast.File
	startTime time.Time
}

func NewCodeGenerator(data map[string]*fieldData.FieldData) *StructGenerator {
	if data == nil {
		data = make(map[string]*fieldData.FieldData)
	}
	return &StructGenerator{
		startTime: time.Now(),
	}
}

func (s *StructGenerator) GenerateIntoDir(jsonData []byte, packageName string, path string, structName string, typeAdjusters []typeAdjustment.TypeDeterminationFunction) error {
	var JsonData interface{}
	var err error
	err = json.Unmarshal(jsonData, &JsonData)
	if err != nil {
		return err
	}
	err = s.codeGen(JsonData, structName, 0)
	if err != nil {
		return err
	}

	err = s.testNew(path, typeAdjusters)
	if err != nil {
		return err
	}
	temp, err := json.Marshal(s.seenTypes)
	if err != nil {
		return err
	}
	err = os.WriteFile(path+"/seenTypes.json", temp, 0666)
	if err != nil {
		return err
	}

	return nil
}

func (s *StructGenerator) getStartPath() (string, error) {
	for path, _ := range s.seenTypes {
		pathElements := strings.Split(path, ".")
		if len(pathElements) == 1 {
			return pathElements[0], nil
		}
	}
	return "", errors.New("start path not found")
}

func (s *StructGenerator) getFieldType(path string) (expr ast.Expr, structType bool) {
	levelOfArrays := math.MaxInt32
	//TODO error on path not found
	if len(s.seenTypes[path].Types) == 1 {
		var Type fieldData.Type
		for Type, _ = range s.seenTypes[path].Types {
			break
		}
		if len(s.seenTypes[path].Types[Type]) == 1 {
			for levelOfArrays, _ = range s.seenTypes[path].Types[Type] {
				break
			}
			pathElements := strings.Split(path, ".")
			if Type == fieldData.Field {
				structType = true
				expr = utils.GeneratedNestedArray(levelOfArrays, &ast.StarExpr{X: &ast.Ident{Name: pathElements[len(pathElements)-1]}})
			} else {
				expr = utils.GeneratedNestedArray(levelOfArrays, &ast.Ident{Name: string(Type)})
			}
		} else {
			for i, _ := range s.seenTypes[path].Types[Type] {
				if levelOfArrays > i {
					levelOfArrays = i
				}
			}
			expr = utils.GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
		}
	} else {
		for _, m := range s.seenTypes[path].Types {
			for i, _ := range m {
				if levelOfArrays > i {
					levelOfArrays = i
				}
			}
		}
		expr = utils.GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
	}
	return expr, structType
}

func reversePath(path string, reversePath map[string]int) {
	pathElements := strings.Split(path, ".")
	for i, _ := range pathElements {
		var p string
		for j := range i + 1 {
			if p == "" {
				p = pathElements[len(pathElements)-j-1]
			} else {
				p = p + "." + pathElements[len(pathElements)-j-1]
			}
		}
		if _, ok := reversePath[p]; ok {
			reversePath[p]++
		} else {
			reversePath[p] = 1
		}
	}
}

func (s *StructGenerator) markNamingConflicts() {
	rPath := make(map[string]int)

	for path, data := range s.seenTypes {
		if _, ok := data.Types[fieldData.Field]; ok {
			reversePath(path, rPath)
		}
	}

	for path, data := range s.seenTypes {
		if _, ok := data.Types[fieldData.Field]; ok {
			pathElements := strings.Split(path, ".")
			p := pathElements[len(pathElements)-1]
			for i := 1; i < len(pathElements); i++ {
				if rPath[p] == 1 && i == 1 {
					break
				}
				if rPath[p] == 1 {
					slices.Reverse([]byte(p))
					s.seenTypes[path].Package = &p
					break
				}
				p += "." + pathElements[len(pathElements)-i-1]
			}
		}
	}
}

/*
func (s *StructGenerator) setOmitempty() {
	for s2, data := range s.seenTypes {
		for t, m := range data.Types {
			for i, m2 := range m {

			}
		}
	}
}

*/

func (s *StructGenerator) testNew(filePath string, typeAdjusters []typeAdjustment.TypeDeterminationFunction) error {
	s.markNamingConflicts()
	startPath, err := s.getStartPath()
	if err != nil {
		return err
	}

	pathsToProcess := []string{startPath}
	files := make(map[string]*ast.File)
	fset := token.NewFileSet()
	defaultFile, err := parser.ParseFile(fset, "", "package test", parser.ParseComments)
	if err != nil {
		panic(err)
	}

	for {
		for _, path := range pathsToProcess {
			var file *ast.File

			if s.seenTypes[path].Package != nil {
				if _, ok := files[*s.seenTypes[path].Package]; ok {
					file = files[*s.seenTypes[path].Package]
				} else {
					fset := token.NewFileSet()
					file, err = parser.ParseFile(fset, "", "package "+strings.ReplaceAll(*s.seenTypes[path].Package, ".", ""), parser.ParseComments)
					if err != nil {
						return err
					}
					files[*s.seenTypes[path].Package] = file
				}
			} else {
				file = defaultFile
			}
			expr, structType := s.getFieldType(path)
			if structType {
				var fields []*ast.Field
				var levelOfArrays int
				for levelOfArrays, _ = range s.seenTypes[path].Types["field"] {
					break
				}
				for fieldPath, _ := range s.seenTypes[path].Types["field"][levelOfArrays] {
					expr, structType = s.getFieldType(fieldPath)
					if structType {
						pathsToProcess = append(pathsToProcess, fieldPath)
					}
					pathElements := strings.Split(fieldPath, ".")
					fields = append(fields, &ast.Field{
						Names: []*ast.Ident{{Name: pathElements[len(pathElements)-1]}},
						Type:  expr,
						Tag:   &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("`json:\"%s,omitempty\"`", s.seenTypes[fieldPath].JsonFieldName)},
					})
				}
				pathElements := strings.Split(path, ".")
				file.Decls = append(file.Decls, &ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: pathElements[len(pathElements)-1],
							},
							Type: &ast.StructType{
								Fields: &ast.FieldList{List: fields},
							},
						},
					},
				})
			} else {
				fmt.Println(path)
				file.Decls = append(file.Decls, &ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: path,
							},
							Type: expr,
						},
					},
				})
			}
			pathsToProcess = pathsToProcess[1:]
		}
		if len(pathsToProcess) == 0 {
			break
		}
	}

	for path, file := range files {
		err = os.MkdirAll(strings.ReplaceAll(filePath+"/"+path, ".", "/"), os.ModePerm)
		output := bytes.NewBuffer([]byte{})
		fset := token.NewFileSet()
		if err := printer.Fprint(output, fset, file); err != nil {
			return err
		}
		c, err := format.Source(output.Bytes())
		err = os.WriteFile(strings.ReplaceAll(filePath+"/"+path, ".", "/")+"/testNew.go", c, 0666)
		if err != nil {
			return err
		}

	}

	output := bytes.NewBuffer([]byte{})
	fset = token.NewFileSet()
	if err := printer.Fprint(output, fset, defaultFile); err != nil {
		return err
	}
	c, err := format.Source(output.Bytes())
	err = os.WriteFile(filePath+"/test.go", c, 0666)
	if err != nil {
		return err
	}
	return nil
}

/*
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
					Type: utils.GeneratedNestedArray(levelOfArrays, &ast.StarExpr{
						X: &ast.Ident{
							Name: structName + "AnonymousArray",
						},
					}),
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
					Type: utils.GeneratedNestedArray(levelOfArrays, *expr),
				},
			},
		})
	}
	return nil
}
*/

func (s *StructGenerator) codeGen(jsonData interface{}, path string, depth int) error {
	switch result := jsonData.(type) {
	case map[string]interface{}:
		err := s.processStruct(result, path, depth)
		if err != nil {
			return err
		}
	case []interface{}:
		err := s.processSlice(result, path, depth)
		if err != nil {
			return err
		}
	default:
		err := s.processField(result, path, depth)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StructGenerator) setTypeAtLevel(path string, Type fieldData.Type, Depth int, value string) {
	valueDetails := &fieldData.ValueDetails{
		Count:              1,
		FirstSeenTimestamp: s.startTime.Unix(),
		LastSeenTimestamp:  s.startTime.Unix(),
	}
	if s.seenTypes == nil {
		s.seenTypes = map[string]*fieldData.PathData{}
	}
	if v, ok := s.seenTypes[path]; !ok {
		s.seenTypes[path] = &fieldData.PathData{
			Types: map[fieldData.Type]map[int]map[string]*fieldData.ValueDetails{},
		}
	} else if v.Types == nil {
		s.seenTypes[path].Types = map[fieldData.Type]map[int]map[string]*fieldData.ValueDetails{}
	}
	if _, ok := s.seenTypes[path].Types[Type]; !ok {
		s.seenTypes[path].Types[Type] = map[int]map[string]*fieldData.ValueDetails{}
	}
	if _, ok := s.seenTypes[path].Types[Type][Depth]; !ok {
		s.seenTypes[path].Types[Type][Depth] = map[string]*fieldData.ValueDetails{}
	}
	if _, ok := s.seenTypes[path].Types[Type][Depth][value]; !ok {
		s.seenTypes[path].Types[Type][Depth][value] = valueDetails
	} else {
		s.seenTypes[path].Types[Type][Depth][value] = s.seenTypes[path].Types[Type][Depth][value].Combine(valueDetails)
	}
}

// Processes JSON-Struct elements
func (s *StructGenerator) processStruct(structData map[string]interface{}, path string, depth int) error {
	for fieldName, field := range structData {
		s.setTypeAtLevel(path, fieldData.Field, depth, path+"."+utils.JsonNameToGoName(fieldName))
		if _, ok := s.seenTypes[path+"."+utils.JsonNameToGoName(fieldName)]; !ok {
			s.seenTypes[path+"."+utils.JsonNameToGoName(fieldName)] = &fieldData.PathData{JsonFieldName: fieldName}
		} else {
			s.seenTypes[path+"."+utils.JsonNameToGoName(fieldName)].JsonFieldName = fieldName
		}
		err := s.codeGen(field, path+"."+utils.JsonNameToGoName(fieldName), 0)
		if err != nil {
			return err
		}
	}
	return nil
}

// Processes JSON-Array elements
func (s *StructGenerator) processSlice(sliceData []interface{}, path string, depth int) error {
	var err error
	depth++
	for _, i := range sliceData {
		switch v := i.(type) {
		case []interface{}:
			err = s.processSlice(v, path, depth)
			if err != nil {
				return err
			}
		case map[string]interface{}:
			err = s.processStruct(v, path, depth)
			if err != nil {
				return err
			}
		case interface{}:
			err = s.processField(v, path, depth)
			if err != nil {
				return err
			}
		}
	}

	//TODO we need to handle empty [] array

	return nil
}

func (s *StructGenerator) processField(field interface{}, path string, depth int) error {
	var fieldValue string
	switch t := field.(type) {
	case float64:
		fieldValue = strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		if t {
			fieldValue = "true"
		}
		fieldValue = "false"
	case string:
		fieldValue = field.(string)
	default:
		return errors.New(fmt.Sprintf("unsupported type of field data: %T", field))
	}
	Type, err := fieldData.TypeFromAny(field)
	if err != nil {
		return err
	}
	s.setTypeAtLevel(path, Type, depth, fieldValue)
	return nil
}
