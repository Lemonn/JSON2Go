package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
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
		parsedFile, err := parser.ParseFile(fset, file, src, parser.AllErrors)
		if err != nil {
			return nil, err
		}

		// Get package name from file's directory
		dir := filepath.Dir(file)
		packageName := filepath.Base(dir)

		// Traverse the AST to collect types and their methods
		ast.Inspect(parsedFile, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:
				typeName := x.Name.Name
				key := fmt.Sprintf("%s.%s", packageName, typeName)
				if _, exists := typeInfos[key]; !exists {
					typeInfos[key] = &TypeInfo{
						Name:        typeName,
						PackageName: packageName,
						Methods:     make(map[string]*MethodInfo),
					}
				}
			case *ast.FuncDecl:
				if x.Recv != nil && len(x.Recv.List) > 0 {
					// Get receiver type
					var typeName string
					switch recvType := x.Recv.List[0].Type.(type) {
					case *ast.StarExpr:
						// Pointer receiver
						if ident, ok := recvType.X.(*ast.Ident); ok {
							typeName = ident.Name
						}
					case *ast.Ident:
						// Value receiver
						typeName = recvType.Name
					}
					methodName := x.Name.Name

					// Count number of parameters and results
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
								// Anonymous result (no name)
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

func insertOrReplaceMethod(file *ast.File, receiverType, methodName, generatedCode string) error {
	var methodFound bool
	var newDecls []ast.Decl

	// Parse the generated method code to get its AST node
	genFile, err := parser.ParseFile(token.NewFileSet(), "", "package main \n"+generatedCode, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse generated code: %w", err)
	}
	var generatedMethod *ast.FuncDecl
	for _, decl := range genFile.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == methodName {
				generatedMethod = funcDecl
				break
			}
		}
	}
	if generatedMethod == nil {
		return fmt.Errorf("generated method %s not found in parsed code", methodName)
	}

	// Iterate over declarations in the existing file
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			// Check if this is the method we want to replace
			if funcDecl.Name.Name == methodName && funcDecl.Recv != nil {
				recvType := ""
				switch recv := funcDecl.Recv.List[0].Type.(type) {
				case *ast.StarExpr:
					if ident, ok := recv.X.(*ast.Ident); ok {
						recvType = ident.Name
					}
				case *ast.Ident:
					recvType = recv.Name
				}
				if recvType == receiverType {
					// Replace this method with the generated one
					methodFound = true
					newDecls = append(newDecls, generatedMethod)
					continue // Skip adding the old method
				}
			}
		}
		newDecls = append(newDecls, decl)
	}
	if !methodFound {
		// Method not found; append the generated method
		newDecls = append(newDecls, generatedMethod)
	}

	// Update the file declarations
	file.Decls = newDecls
	return nil
}

func main() {
	// Command-line flags
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

	// Step 1: Collect types that implement FileError interface
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

	// Identify types implementing FileError interface
	var fileErrorTypes []*TypeInfo
	for _, info := range typeInfos {
		if methodInfo, exists := info.Methods["IsFileError"]; exists {
			if methodInfo.Params == 0 && methodInfo.Results == 0 {
				// Method signature matches: no parameters, no results
				fileErrorTypes = append(fileErrorTypes, info)
			}
		}
	}

	// Step 2: Generate UnmarshalJSON method using text/template
	methodCode, err := generateUnmarshalJSON(fileErrorTypes, receiverType)
	if err != nil {
		fmt.Printf("Error generating UnmarshalJSON method: %v\n", err)
		return
	}

	// Step 3: Read and parse the target Go file
	fset := token.NewFileSet()
	src, err := ioutil.ReadFile(targetFile)
	if err != nil {
		fmt.Printf("Error reading target Go file: %v\n", err)
		return
	}
	fileAST, err := parser.ParseFile(fset, targetFile, src, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing target Go file: %v\n", err)
		return
	}

	// Step 4: Insert or replace UnmarshalJSON method in the AST
	err = insertOrReplaceMethod(fileAST, receiverType, "UnmarshalJSON", methodCode)
	if err != nil {
		fmt.Printf("Error inserting/replacing method: %v\n", err)
		return
	}

	// Step 5: Write the modified AST back to the file
	var buf bytes.Buffer
	err = formatNode(&buf, fset, fileAST)
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

// Helper function to format and print AST node
func formatNode(w io.Writer, fset *token.FileSet, node interface{}) error {
	cfg := &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	return cfg.Fprint(w, fset, node)
}
