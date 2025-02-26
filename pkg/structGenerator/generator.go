//go:build test
// +build test

package structGenerator

import (
	"errors"
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"strconv"
)

import (
	"bytes"
	"encoding/json"
	"github.com/Lemonn/AstUtils"
	"github.com/Lemonn/JSON2Go/pkg/jsonMarshallerGenerators/marshaller"
	"github.com/Lemonn/JSON2Go/pkg/jsonMarshallerGenerators/unmarshaller"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment/buildin"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"math"
	"os"
	"strings"
	"time"
)

type StructGenerator struct {
	seenTypes         map[string]*fieldData.PathData
	file              *ast.File
	startTime         time.Time
	files             map[string]*FileData
	stackedMarshaller map[string]*ast.File
}

type FileData struct {
	file *ast.File
	fSet *token.FileSet
}

func (f *FileData) handleImports(Package string) error {
	var foundNodes []*AstUtils.FoundNodes
	var completed bool
	pathElements := strings.Split(Package, ".")
	AstUtils.SearchNodes(f.file, &foundNodes, []*ast.Node{}, func(n *ast.Node, parents []*ast.Node, completed *bool) bool {
		if v, ok := (*n).(*ast.StarExpr); ok {
			output := bytes.NewBuffer([]byte{})
			fset := token.NewFileSet()
			if err := printer.Fprint(output, fset, v); err != nil {

			}
			if vi, ok := v.X.(*ast.Ident); ok && vi.Name == pathElements[len(pathElements)-1] {
				return true
			}
		}
		return false
	}, &completed)

	return nil
}

func NewCodeGenerator() *StructGenerator {
	return &StructGenerator{
		startTime:         time.Now(),
		seenTypes:         make(map[string]*fieldData.PathData),
		files:             make(map[string]*FileData),
		stackedMarshaller: map[string]*ast.File{},
	}
}

func (s *StructGenerator) IsStruct(path string) bool {
	if len(s.seenTypes[path].Types) == 1 || s.emptySubtype(path) {
		Type := s.getType(path)
		if len(s.seenTypes[path].Types[Type]) == 1 {
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

func (s *StructGenerator) GenerateIntoDir(jsonData []byte, packageName string, path string, structName string) error {
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

	if s.IsStruct(structName) {
		var levelOfArrays int
		for levelOfArrays, _ = range s.seenTypes[structName].Types["field"] {
			break
		}
		if levelOfArrays > 0 {
			s.seenTypes[structName+".AnonymousArray"] = s.seenTypes[structName]

			s.seenTypes[structName] = &fieldData.PathData{
				Types:                   map[fieldData.Type]map[int]map[string]*fieldData.ValueDetails{},
				JsonFieldName:           "",
				Package:                 nil,
				Omitempty:               false,
				TypeAdjusterData:        nil,
				Error:                   nil,
				RequiredField:           false,
				ForceSourceType:         nil,
				DirectToForceSourceType: false,
			}
			s.seenTypes[structName].Types[fieldData.Field] = map[int]map[string]*fieldData.ValueDetails{}
			s.seenTypes[structName].Types[fieldData.Field][levelOfArrays] = make(map[string]*fieldData.ValueDetails)
			s.seenTypes[structName].Types[fieldData.Field][levelOfArrays][structName+".AnonymousArray"] = &fieldData.ValueDetails{
				Count:              0,
				FirstSeenTimestamp: 0,
				LastSeenTimestamp:  0,
			}
		}

	}

	err = s.testNew(path)
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

func (s *StructGenerator) getType(path string) fieldData.Type {
	for Type, _ := range s.seenTypes[path].Types {
		if len(s.seenTypes[path].Types) == 1 {
			return Type
		} else if Type != fieldData.EmptyArray && Type != fieldData.EmptyStruct {
			return Type
		}
	}
	return fieldData.Unsupported
}

func (s *StructGenerator) emptySubtype(path string) bool {
	var EmptySubtype bool
	var Level int
	if len(s.seenTypes[path].Types) == 2 {
		if _, ok := s.seenTypes[path].Types[fieldData.Field]; ok {
			if v, ok := s.seenTypes[path].Types[fieldData.EmptyArray]; ok {
				for Level, _ = range s.seenTypes[path].Types[fieldData.Field] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					EmptySubtype = true
				}

			} else if _, ok := s.seenTypes[path].Types[fieldData.EmptyStruct]; ok {
				for Level, _ = range s.seenTypes[path].Types[fieldData.EmptyStruct] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					EmptySubtype = true
				}
			}
		} else if _, ok := s.seenTypes[path].Types[fieldData.String]; ok {
			if v, ok := s.seenTypes[path].Types[fieldData.EmptyArray]; ok {
				for Level, _ = range s.seenTypes[path].Types[fieldData.String] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					EmptySubtype = true
				}
			}
		} else if _, ok := s.seenTypes[path].Types[fieldData.Float64]; ok {
			if v, ok := s.seenTypes[path].Types[fieldData.EmptyArray]; ok {
				for Level, _ = range s.seenTypes[path].Types[fieldData.Float64] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					EmptySubtype = true
				}
			}
		} else if _, ok := s.seenTypes[path].Types[fieldData.Bool]; ok {
			if v, ok := s.seenTypes[path].Types[fieldData.EmptyArray]; ok {
				for Level, _ = range s.seenTypes[path].Types[fieldData.Bool] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					EmptySubtype = true
				}
			}
		}
	}
	return EmptySubtype
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

func (s *StructGenerator) getFieldType(path string) (expr ast.Expr) {
	levelOfArrays := math.MaxInt32
	//TODO error on path not found
	if s.seenTypes[path].ForceSourceType != nil {
		parseExpr, err := parser.ParseExpr(*s.seenTypes[path].ForceSourceType)
		if err != nil {
			return nil
		}
		return parseExpr
	}

	if len(s.seenTypes[path].Types) == 1 || s.emptySubtype(path) {
		Type := s.getType(path)
		if len(s.seenTypes[path].Types[Type]) == 1 {
			for levelOfArrays, _ = range s.seenTypes[path].Types[Type] {
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

	return expr
}

func (s *StructGenerator) testNew(filePath string) error {
	//s.markNamingConflicts()
	startPath, err := s.getStartPath()
	if err != nil {
		return err
	}

	pathsToProcess := []string{startPath}

	for {
		for _, path := range pathsToProcess {
			var file *ast.File
			if _, ok := s.files[path]; !ok {
				fset := token.NewFileSet()
				pathElements := strings.Split(path, ".")
				file, err = parser.ParseFile(fset, "", "package "+pathElements[len(pathElements)-1], parser.ParseComments)
				if err != nil {
					return err
				}
				s.files[path] = &FileData{
					file: file,
					fSet: fset,
				}
				file = s.files[path].file
			}

			file = s.files[path].file
			if s.IsStruct(path) {
				var fields []*ast.Field
				var levelOfArrays int
				for levelOfArrays, _ = range s.seenTypes[path].Types["field"] {
					break
				}
				s.stackedMarshaller[path] = file
				for fieldPath, _ := range s.seenTypes[path].Types["field"][levelOfArrays] {
					//Adjust Type
					ta := typeAdjustment.NewTypeAdjuster(s.seenTypes, []typeAdjustment.TypeDeterminationFunction{&buildin.UUIDTypeChecker{}}, s.startTime)
					err := ta.AdjustTypesNew(fieldPath)
					if err != nil {
						return err
					}
					expr := s.getFieldType(fieldPath)
					if s.IsStruct(fieldPath) {
						pathsToProcess = append(pathsToProcess, fieldPath)
						AstUtils.AddMissingImports(file, []string{strings.ReplaceAll("out/"+fieldPath, ".", "/")})
					}

					pathElements := strings.Split(fieldPath, ".")
					fields = append(fields, &ast.Field{
						Names: []*ast.Ident{{Name: pathElements[len(pathElements)-1]}},
						Type:  expr,
						//TODO set path correctly. Do not set a path if JsonFieldName == "". Write a function, that checks if omitempty should be set
						Tag: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("`json:\"%s,omitempty\"`", s.seenTypes[fieldPath].JsonFieldName)},
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
							Type: func() ast.Expr {
								return &ast.StructType{
									Fields: &ast.FieldList{List: fields},
								}
							}(),
						},
					},
				})
			} else {
				file.Decls = append(file.Decls, &ast.GenDecl{
					Tok: token.TYPE,
					Specs: []ast.Spec{
						&ast.TypeSpec{
							Name: &ast.Ident{
								Name: path,
							},
							Type: s.getFieldType(path),
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

	for path, file := range s.stackedMarshaller {
		fmt.Println(path)
		uGen := unmarshaller.NewGenerator(s.seenTypes)
		generate, _, err := uGen.Generate(path)
		fmt.Println(generate)
		if err != nil {
			return err
		}
		file.Decls = append(file.Decls, generate...)

		mGen := marshaller.NewGenerator(s.seenTypes)
		gen, _, err := mGen.Generate(path)
		fmt.Println(generate)
		if err != nil {
			return err
		}
		file.Decls = append(file.Decls, gen...)
	}

	for path, file := range s.files {
		/*
			fmt.Println(path)
			err := file.handleImports(path)
			if err != nil {
				return err
			}

		*/

		err = os.MkdirAll(strings.ReplaceAll(filePath+"/"+path, ".", "/"), os.ModePerm)
		if err != nil {
			return err
		}
		output := bytes.NewBuffer([]byte{})
		if err := printer.Fprint(output, file.fSet, file.file); err != nil {
			return err
		}
		c, err := format.Source(output.Bytes())
		if err != nil {
			return err
		}
		pathElements := strings.Split(path, ".")
		err = os.WriteFile(strings.ReplaceAll(filePath+"/"+path, ".", "/")+"/"+pathElements[len(pathElements)-1]+".go", c, 0666)
		if err != nil {
			return err
		}

	}
	return nil
}

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
	if len(structData) == 0 {
		s.setTypeAtLevel(path, fieldData.EmptyStruct, depth, "{}")
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
	if len(sliceData) == 0 {
		s.setTypeAtLevel(path, fieldData.EmptyArray, depth, "[]")
	}
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
