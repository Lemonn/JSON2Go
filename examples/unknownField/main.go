package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"github.com/Lemonn/JSON2Go/pkg/structGenerator"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment"
	"github.com/Lemonn/JSON2Go/pkg/typeAdjustment/buildin"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
)

func main() {

	// Empty output.go except package definition before rerun
	//generate()

	//Unmarshal test obj
	run()

}

func run() {
	var example Example
	file, err := os.ReadFile("examples/unknownField/files/input.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(file, &example)

	// Handle hard unmarshalling errors
	e := CheckForFirstErrorNotOfTypeT(&AdditionalElementsError{}, err)
	if e != nil {
		panic("hard error " + e.Error())
	}

	// Handle soft errors
	elementErrors, err := GetAllErrorsOfType(&AdditionalElementsError{}, err)
	if err != nil {
		return
	} else {
		for _, elementError := range elementErrors {
			fmt.Println(elementError.String())
		}
	}
}

func generate() {
	var metadata *fieldData.Metadata

	//Read input JSON-File
	readFile, err := os.ReadFile("examples/unknownField/files/generate.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	//Read go file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "examples/unknownField/output.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Generate code. Include build-in TypeAdjusters and generate Marshall/Un-Marshall code
	strGen := structGenerator.NewCodeGenerator(nil)
	metadata, err = strGen.GenerateCodeIntoFile(readFile, file, "Example", []typeAdjustment.TypeDeterminationFunction{buildin.NewTimeTypeChecker(true), &buildin.UUIDTypeChecker{}, &buildin.IntTypeChecker{}}, true)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Write JSON whit field data
	marshal, err := json.Marshal(metadata)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = os.WriteFile("examples/unknownField/fieldData.json", marshal, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Write generated Golang code
	output := bytes.NewBuffer([]byte{})
	if err = printer.Fprint(output, fset, file); err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("examples/unknownField/output.go", output.Bytes(), 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
}
