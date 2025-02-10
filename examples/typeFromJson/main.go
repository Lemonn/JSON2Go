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
	var tags map[string]*fieldData.FieldData

	//Read input JSON-File
	readFile, err := os.ReadFile("examples/typeFromJson/input.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	//Read go file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "examples/typeFromJson/output.go", nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Generate code. Include build-in TypeAdjusters and generate Marshall/Un-Marshall code
	strGen := structGenerator.NewCodeGenerator(nil)
	tags, err = strGen.GenerateCodeIntoFile(readFile, file, "Example", []typeAdjustment.TypeDeterminationFunction{buildin.NewTimeTypeChecker(true), &buildin.UUIDTypeChecker{}, &buildin.IntTypeChecker{}}, true)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Write JSON whit field data
	marshal, err := json.Marshal(tags)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = os.WriteFile("examples/typeFromJson/fieldData.json", marshal, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Write generated Golang code
	output := bytes.NewBuffer([]byte{})
	if err = printer.Fprint(output, fset, file); err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("examples/typeFromJson/output.go", output.Bytes(), 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
}
