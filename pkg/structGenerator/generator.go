package structGenerator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lemonn/AstUtils"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/jsonMarshallerGenerators/unmarshaller"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment/buildin"
	"slices"

	//"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"math"
	"os"
	"strconv"
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

/*
func (f *FileData) getFileByPackage(packageName string) *FileData {
	if v, ok := f[packageName]; ok {
		return v
	} else {
		fSet := token.NewFileSet()
		file, err := parser.ParseFile(fSet, "", "package test", parser.ParseComments)
		if err != nil {
			panic(err)
		}
		f[packageName] = &FileData{
			file: file,
			fSet: fSet,
		}
		return f[packageName]
	}
}

*/

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
						Name: "B",
					},
					Type: &ast.StructType{
						Fields: &ast.FieldList{
							List: []*ast.Field{
								&ast.Field{
									Names: []*ast.Ident{
										&ast.Ident{
											Name: "Inner",
										},
									},
									Type: &ast.StarExpr{
										X: &ast.SelectorExpr{
											X: &ast.Ident{
												Name: "BInner",
											},
											Sel: &ast.Ident{
												Name: "Inner",
											},
										},
									},
									Tag: &ast.BasicLit{
										Kind:  token.STRING,
										Value: "`json:\"Inner,omitempty\"`",
									},
								},
								&ast.Field{
									Names: []*ast.Ident{
										&ast.Ident{
											Name: "H",
										},
									},
									Type: &ast.StarExpr{
										X: &ast.Ident{
											Name: "H",
										},
									},
									Tag: &ast.BasicLit{
										Kind:  token.STRING,
										Value: "`json:\"H,omitempty\"`",
									},
								},
								&ast.Field{
									Names: []*ast.Ident{
										&ast.Ident{
											Name: "Duplicate",
										},
									},
									Type: &ast.StarExpr{
										X: &ast.SelectorExpr{
											X: &ast.Ident{
												Name: "BDuplicate",
											},
											Sel: &ast.Ident{
												Name: "Duplicate",
											},
										},
									},
									Tag: &ast.BasicLit{
										Kind:  token.STRING,
										Value: "`json:\"Duplicate,omitempty\"`",
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

func NewCodeGenerator() *StructGenerator {
	return &StructGenerator{
		startTime:         time.Now(),
		seenTypes:         make(map[string]*fieldData.PathData),
		files:             make(map[string]*FileData),
		stackedMarshaller: map[string]*ast.File{},
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
					f := strings.Split(p, ".")
					slices.Reverse(f)
					var k string
					for _, s2 := range f {
						if k == "" {
							k = s2
						} else {
							k = k + "." + s2
						}
					}
					s.seenTypes[path].Package = &k
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
			/*
				if s.seenTypes[path].Package != nil {
					if _, ok := s.files[*s.seenTypes[path].Package]; ok {
						file = s.files[*s.seenTypes[path].Package].file
					} else {
						fset := token.NewFileSet()
						file, err = parser.ParseFile(fset, "", "package "+strings.ReplaceAll(*s.seenTypes[path].Package, ".", ""), parser.ParseComments)
						if err != nil {
							return err
						}
						s.files[*s.seenTypes[path].Package] = &FileData{
							file: file,
							fSet: fset,
						}
					}
				} else {
					file = defaultFile
				}

			*/
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
			expr, structType := s.getFieldType(path)
			if structType {
				var fields []*ast.Field
				var levelOfArrays int
				for levelOfArrays, _ = range s.seenTypes[path].Types["field"] {
					break
				}
				s.stackedMarshaller[path] = file
				for fieldPath, _ := range s.seenTypes[path].Types["field"][levelOfArrays] {
					expr, structType = s.getFieldType(fieldPath)
					if structType {
						pathsToProcess = append(pathsToProcess, fieldPath)
						AstUtils.AddMissingImports(file, []string{strings.ReplaceAll("out/"+fieldPath, ".", "/")})
					}
					ta := typeAdjustment.NewTypeAdjuster(s.seenTypes, []typeAdjustment.TypeDeterminationFunction{&buildin.UUIDTypeChecker{}}, s.startTime)
					err := ta.AdjustTypesNew(fieldPath)
					if err != nil {
						return err
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

	for path, file := range s.stackedMarshaller {
		fmt.Println(path)
		uGen := unmarshaller.NewGenerator(s.seenTypes)
		generate, _, err := uGen.Generate(path)
		fmt.Println(generate)
		if err != nil {
			return err
		}
		file.Decls = append(file.Decls, generate...)
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
