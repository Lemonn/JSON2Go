package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type MethodInfo struct {
	Name    string
	Params  int
	Results int
}

type TypeInfo struct {
	Name        string
	PackageName string
	Methods     map[string]*MethodInfo
}

func getGoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func collectTypesAndMethods(files []string) (map[string]*TypeInfo, error) {
	typeInfos := make(map[string]*TypeInfo)
	for _, file := range files {
		fset := token.NewFileSet()
		src, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		parsedFile, err := decorator.ParseFile(fset, file, src, parser.AllErrors)
		if err != nil {
			return nil, err
		}

		dir := filepath.Dir(file)
		packageName := filepath.Base(dir)

		dst.Inspect(parsedFile, func(n dst.Node) bool {
			switch x := n.(type) {
			case *dst.TypeSpec:
				typeName := x.Name.Name
				key := fmt.Sprintf("%s.%s", packageName, typeName)
				if _, exists := typeInfos[key]; !exists {
					typeInfos[key] = &TypeInfo{
						Name:        typeName,
						PackageName: packageName,
						Methods:     make(map[string]*MethodInfo),
					}
				}
			case *dst.FuncDecl:
				if x.Recv != nil && len(x.Recv.List) > 0 {
					var typeName string
					switch recvType := x.Recv.List[0].Type.(type) {
					case *dst.StarExpr:
						if ident, ok := recvType.X.(*dst.Ident); ok {
							typeName = ident.Name
						}
					case *dst.Ident:
						typeName = recvType.Name
					}

					methodName := x.Name.Name

					numParams := 0
					if x.Type.Params != nil {
						for _, param := range x.Type.Params.List {
							numParams += len(param.Names)
						}
					}

					numResults := 0
					if x.Type.Results != nil {
						for _, result := range x.Type.Results.List {
							if result.Names != nil {
								numResults += len(result.Names)
							} else {
								numResults++
							}
						}
					}

					if typeName != "" {
						key := fmt.Sprintf("%s.%s", packageName, typeName)
						if _, exists := typeInfos[key]; !exists {
							typeInfos[key] = &TypeInfo{
								Name:        typeName,
								PackageName: packageName,
								Methods:     make(map[string]*MethodInfo),
							}
						}
						typeInfos[key].Methods[methodName] = &MethodInfo{
							Name:    methodName,
							Params:  numParams,
							Results: numResults,
						}
					}
				}
			}
			return true
		})
	}
	return typeInfos, nil
}

func generateUnmarshalJSON(fileErrorTypes []*TypeInfo, receiverType string) (string, error) {
	tmpl := "func (e *{{.ReceiverType}}) UnmarshalJSON(bytes []byte) error {\n" +
		"    localType := struct {\n" +
		"        Type string `json:\"type\"`\n" +
		"        Err  json.RawMessage `json:\"error\"`\n" +
		"    }{}\n" +
		"    if err := json.Unmarshal(bytes, &localType); err != nil {\n" +
		"        return err\n" +
		"    }\n" +
		"    switch localType.Type {\n" +
		"{{- range .FileErrorTypes}}\n" +
		"    case reflect.TypeOf(&{{.PackageName}}.{{.Name}}{}).String():\n" +
		"        var te {{.PackageName}}.{{.Name}}\n" +
		"        if unmarshallError := json.Unmarshal(localType.Err, &te); unmarshallError != nil {\n" +
		"            return unmarshallError\n" +
		"        }\n" +
		"        e.Err = &te\n" +
		"        e.Type = localType.Type\n" +
		"{{- end}}\n" +
		"    default:\n" +
		"        return &UnknownStoreTypeError{TypeName: localType.Type}\n" +
		"    }\n" +
		"    return nil\n" +
		"}\n"

	data := map[string]interface{}{
		"ReceiverType":   receiverType,
		"FileErrorTypes": fileErrorTypes,
	}
	t, err := template.New("unmarshalJSON").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

/*
func insertOrReplaceMethod(file *dst.File, receiverType, methodName, generatedCode string) error {
	var methodFound bool
	var newDecls []dst.Decl

	genFile, err := decorator.ParseFile(token.NewFileSet(), "", "package main \n"+generatedCode, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse generated code: %w", err)
	}

	var generatedMethod *dst.FuncDecl
	for _, decl := range genFile.Decls {
		if funcDecl, ok := decl.(*dst.FuncDecl); ok {
			if funcDecl.Name.Name == methodName {
				generatedMethod = funcDecl
				break
			}
		}
	}

	if generatedMethod == nil {
		return fmt.Errorf("generated method %s not found in parsed code", methodName)
	}

	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*dst.FuncDecl); ok {
			if funcDecl.Name.Name == methodName && funcDecl.Recv != nil {
				recvType := ""
				switch recv := funcDecl.Recv.List[0].Type.(type) {
				case *dst.StarExpr:
					if ident, ok := recv.X.(*dst.Ident); ok {
						recvType = ident.Name
					}
				case *dst.Ident:
					recvType = recv.Name
				}
				if recvType == receiverType {
					methodFound = true
					newDecls = append(newDecls, generatedMethod)
					continue
				}
			}
		}
		newDecls = append(newDecls, decl)
	}

	if !methodFound {
		newDecls = append(newDecls, generatedMethod)
	}

	file.Decls = newDecls
	return nil
}
*/

func insertOrReplaceMethod(file *dst.File, receiverType, methodName, generatedCode string) error {
	var methodFound bool
	var newDecls []dst.Decl

	genFile, err := decorator.ParseFile(token.NewFileSet(), "", "package main \n"+generatedCode, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse generated code: %w", err)
	}

	var generatedMethod *dst.FuncDecl
	for _, decl := range genFile.Decls {
		if funcDecl, ok := decl.(*dst.FuncDecl); ok {
			if funcDecl.Name.Name == methodName {
				generatedMethod = funcDecl
				break
			}
		}
	}

	if generatedMethod == nil {
		return fmt.Errorf("generated method %s not found in parsed code", methodName)
	}

	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*dst.FuncDecl); ok {
			if funcDecl.Name.Name == methodName && funcDecl.Recv != nil {
				recvType := ""
				switch recv := funcDecl.Recv.List[0].Type.(type) {
				case *dst.StarExpr:
					if ident, ok := recv.X.(*dst.Ident); ok {
						recvType = ident.Name
					}
				case *dst.Ident:
					recvType = recv.Name
				}
				if recvType == receiverType {
					methodFound = true
					// Preserve existing comments
					generatedMethod.Decs = funcDecl.Decs
					newDecls = append(newDecls, generatedMethod)
					continue
				}
			}
		}
		newDecls = append(newDecls, decl)
	}

	if !methodFound {
		newDecls = append(newDecls, generatedMethod)
	}

	file.Decls = newDecls
	return nil
}

func main() {
	var (
		rootDir      string
		targetFile   string
		receiverType string
	)
	flag.StringVar(&rootDir, "root", ".", "Root directory to search for Go files")
	flag.StringVar(&targetFile, "file", "", "Path to the Go file to modify")
	flag.StringVar(&receiverType, "receiver", "StoreType", "Receiver type name for UnmarshalJSON method")
	flag.Parse()

	basePath, err := os.Getwd()
	if err != nil {
		return
	}
	if targetFile == "" {
		targetFile = filepath.Join(basePath, "storeType.go")
	}
	if rootDir == "" {
		rootDir = filepath.Join(basePath, "/pkg/errors")
	}
	if receiverType == "" {
		receiverType = "StoreType"
	}

	goFiles, err := getGoFiles(rootDir)
	if err != nil {
		fmt.Printf("Error reading Go files: %v\n", err)
		return
	}
	typeInfos, err := collectTypesAndMethods(goFiles)
	if err != nil {
		fmt.Printf("Error collecting types and methods: %v\n", err)
		return
	}

	var fileErrorTypes []*TypeInfo
	for _, info := range typeInfos {
		if methodInfo, exists := info.Methods["IsFileError"]; exists {
			if methodInfo.Params == 0 && methodInfo.Results == 0 {
				fileErrorTypes = append(fileErrorTypes, info)
			}
		}
	}

	methodCode, err := generateUnmarshalJSON(fileErrorTypes, receiverType)
	if err != nil {
		fmt.Printf("Error generating UnmarshalJSON method: %v\n", err)
		return
	}

	fset := token.NewFileSet()
	src, err := ioutil.ReadFile(targetFile)
	if err != nil {
		fmt.Printf("Error reading target Go file: %v\n", err)
		return
	}
	fileAST, err := decorator.ParseFile(fset, targetFile, src, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing target Go file: %v\n", err)
		return
	}

	err = insertOrReplaceMethod(fileAST, receiverType, "UnmarshalJSON", methodCode)
	if err != nil {
		fmt.Printf("Error inserting/replacing method: %v\n", err)
		return
	}

	var buf bytes.Buffer
	err = formatNode(&buf, fileAST)
	if err != nil {
		fmt.Printf("Error formatting modified AST: %v\n", err)
		return
	}
	err = ioutil.WriteFile(targetFile, buf.Bytes(), 0644)
	if err != nil {
		fmt.Printf("Error writing modified Go file: %v\n", err)
		return
	}
	fmt.Println("Method successfully inserted/replaced in the target Go file.")
}

func formatNode(w io.Writer, node *dst.File) error {
	return decorator.Fprint(w, node)
}
