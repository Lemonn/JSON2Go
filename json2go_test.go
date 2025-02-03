package JSON2Go

import (
	"fmt"
	"github.com/Lemonn/AstUtils"
	"github.com/go-toolsmith/astequal"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"testing"
)

func Test_combineStructFields(t *testing.T) {

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", "package main\n\ntype Test struct {\n\tListOfUuids []interface {\n\t} `json2go:\"eyJzZWVuVmFsdWVzIjp7IiI6ImludGVyZmFjZXt9IiwiOWU0ZjExYzItZDA1OS00MTcyLWE2YjgtMWZiM2RjYjc3OTE2Ijoic3RyaW5nIiwiYzQ2ZjliNWMtMTNiOS00ZGI4LTkyNzItZmM1NDliMjZlOTBiIjoic3RyaW5nIn0sIm1peGVkVHlwZXMiOnRydWUsImxhc3RTZWVuVGltZXN0YW1wIjoxNzM4NTk3OTQ3fQ==\" json:\"ListOfUUIDs\" `\n}", parser.ParseComments)
	if err != nil {
		panic(err)
	}

	fset1 := token.NewFileSet()
	file1, err := parser.ParseFile(fset1, "", "package main\n\ntype Test struct {\n\tListOfUuids []uuid.UUID `json2go:\"eyJzZWVuVmFsdWVzIjp7IjllNGYxMWMyLWQwNTktNDE3Mi1hNmI4LTFmYjNkY2I3NzkxNiI6InN0cmluZyIsImM0NmY5YjVjLTEzYjktNGRiOC05MjcyLWZjNTQ5YjI2ZTkwYiI6InN0cmluZyJ9LCJwYXJzZUZ1bmN0aW9ucyI6eyJmcm9tVHlwZVBhcnNlRnVuY3Rpb24iOiJmcm9tTGlzdE9mVXVpZHNSUlIiLCJ0b1R5cGVQYXJzZUZ1bmN0aW9uIjoidG9MaXN0T2ZVdWlkc1JSUiJ9LCJiYXNlVHlwZSI6InN0cmluZyIsImxhc3RTZWVuVGltZXN0YW1wIjoxNzM4NTk4OTgzfQ==\" json:\"ListOfUUIDs\" `\n}", parser.ParseComments)
	if err != nil {
		panic(err)
	}

	type args struct {
		oldElement *ast.StructType
		newElement *ast.StructType
	}
	tests := []struct {
		name    string
		args    args
		want    []*ast.Field
		wantErr bool
	}{
		{name: "StructWhitConflictingArrayFieldsAndCustomType", args: struct {
			oldElement *ast.StructType
			newElement *ast.StructType
		}{oldElement: file.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType), newElement: file1.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)}, want: []*ast.Field{
			{
				Names: []*ast.Ident{
					{
						Name: "ListOfUuids",
					},
				},
				Type: &ast.ArrayType{
					Elt: &ast.InterfaceType{
						Methods: &ast.FieldList{},
					},
				},
				Tag: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "`json:\"ListOfUUIDs\" json2go:\"eyJzZWVuVmFsdWVzIjp7IiI6ImludGVyZmFjZXt9IiwiOWU0ZjExYzItZDA1OS00MTcyLWE2YjgtMWZiM2RjYjc3OTE2Ijoic3RyaW5nIiwiYzQ2ZjliNWMtMTNiOS00ZGI4LTkyNzItZmM1NDliMjZlOTBiIjoic3RyaW5nIn0sIm1peGVkVHlwZXMiOnRydWUsImxhc3RTZWVuVGltZXN0YW1wIjoxNzM4NTk4OTgzfQ==\"`",
				},
			},
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := combineStructFields(tt.args.oldElement, tt.args.newElement)
			fmt.Println(got)
			for _, field := range got {
				fmt.Println(field.Names[0].Name)
				fmt.Println(field.Tag.Value)
			}
			file.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List = got

			if err = printer.Fprint(os.Stdout, fset, file); err != nil {
				log.Fatal(err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("combineStructFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !astequal.Node(&ast.FieldList{List: got}, &ast.FieldList{List: tt.want}) {
				t.Errorf("combineStructFields()1 got = %v, want %v", got, tt.want)
			}
			for i, field := range got {
				if !AstUtils.TagsEqual(field.Tag, tt.want[i].Tag) {
					t.Errorf("combineStructFields() got = %v, want %v", field.Tag.Value, tt.want[i].Tag.Value)
				}
			}
		})
	}
}
