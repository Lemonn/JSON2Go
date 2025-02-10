package structGenerator

import (
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"reflect"
	"strings"
)

func (s *StructGenerator) combineStructFields(oldElement, newElement *ast.StructType, path string) ([]*ast.Field, error) {
	var combinedFields []*ast.Field
	fields := map[string][]*ast.Field{}

	for _, oldField := range oldElement.Fields.List {
		if _, ok := fields[oldField.Names[0].Name]; !ok {
			fields[oldField.Names[0].Name] = []*ast.Field{}
		}
		fields[oldField.Names[0].Name] = append(fields[oldField.Names[0].Name], oldField)
	}

	for _, newField := range newElement.Fields.List {
		if _, ok := fields[newField.Names[0].Name]; !ok {
			fields[newField.Names[0].Name] = []*ast.Field{}
		}
		fields[newField.Names[0].Name] = append(fields[newField.Names[0].Name], newField)
	}

	for fieldName, exprs := range fields {
		if len(exprs) == 1 {
			s.data[path+"."+fieldName].Omitempty = true
			combinedFields = append(combinedFields, exprs[0])
			continue
		}
		combinedExpr, err := s.combineFields(exprs[0].Type, exprs[1].Type, path+"."+exprs[0].Names[0].Name)
		if err != nil {
			return nil, err
		}
		combinedFields = append(combinedFields, &ast.Field{
			Doc:     exprs[0].Doc,
			Names:   exprs[0].Names,
			Type:    combinedExpr,
			Tag:     exprs[0].Tag,
			Comment: exprs[0].Comment,
		})
	}
	return combinedFields, nil
}

// TODO set conflicting field json2go tag value
func (s *StructGenerator) combineFields(expr0, expr1 ast.Expr, path string) (ast.Expr, error) {
	//Get Tags of both fields
	fData := s.data[path]

	//Traverse as long as both fields are of type *ast.Array
	var level int
	for reflect.TypeOf(expr0) == reflect.TypeOf(expr1) {
		if _, ok := expr0.(*ast.ArrayType); ok {
			if _, ok := expr1.(*ast.ArrayType); ok {
				level++
				expr0 = expr0.(*ast.ArrayType).Elt
				expr1 = expr1.(*ast.ArrayType).Elt
			} else {
				break
			}
		} else {
			break
		}
	}

	var finalExpr ast.Expr
	// As we traversed down all potential array layers, the only possible cases are now
	// 1. Both types are equal, in this case it's either a StructType, which needs to be combined.
	// Or an equal field type such as InterfaceType which does not need special treatment.

	// 2. One field is of interface type. In this case we look if it's a mixed type. If that's the case, we set it
	// to InterfaceType. Otherwise, it's case of a previously empty field, for which we're seen values now. This means
	// we set the field to the new values

	// 3 Fields of mixed types, that could not be combined into one, in this case set field to InterfaceType
	if reflect.TypeOf(expr0) == reflect.TypeOf(expr1) {
		finalExpr = expr0
		if _, ok := expr0.(*ast.StructType); ok {
			fields, err := s.combineStructFields(expr0.(*ast.StructType), expr1.(*ast.StructType), path)
			if err != nil {
				return nil, err
			}
			finalExpr = &ast.StructType{Fields: &ast.FieldList{List: fields}}
		} else if v, ok := expr0.(*ast.Ident); ok && v.Name != expr1.(*ast.Ident).Name {
			s.data[path].MixedTypes = true
			finalExpr = &ast.InterfaceType{Methods: &ast.FieldList{}}
		}

		for range level {
			finalExpr = &ast.ArrayType{Elt: finalExpr}
		}
		return finalExpr, nil
	} else if _, ok := expr0.(*ast.InterfaceType); ok {
		if fData != nil && fData.MixedTypes {
			finalExpr = &ast.InterfaceType{Methods: &ast.FieldList{}}
			for range level {
				finalExpr = &ast.ArrayType{Elt: finalExpr}
			}
			return finalExpr, nil
		} else {
			finalExpr = expr1
			for range level {
				finalExpr = &ast.ArrayType{Elt: finalExpr}
			}
			return finalExpr, nil
		}

	} else if _, ok := expr1.(*ast.InterfaceType); ok {
		if fData != nil && fData.MixedTypes {
			finalExpr = &ast.InterfaceType{Methods: &ast.FieldList{}}
			for range level {
				finalExpr = &ast.ArrayType{Elt: finalExpr}
			}
			return finalExpr, nil
		} else {
			finalExpr = expr0
			for range level {
				finalExpr = &ast.ArrayType{Elt: finalExpr}
			}
			return finalExpr, nil
		}
	} else {
		finalExpr = &ast.InterfaceType{Methods: &ast.FieldList{}}
		for range level {
			finalExpr = &ast.ArrayType{Elt: finalExpr}
		}
		s.data[path].MixedTypes = true
		return finalExpr, nil
	}
}

func resetToBaseType(expr *ast.Expr, json2go *fieldData.FieldData) {
	if json2go != nil && json2go.BaseType != nil {
		if *json2go.BaseType == "interface{}" {
			*expr = &ast.InterfaceType{Methods: &ast.FieldList{}}
		}
		*expr = &ast.Ident{
			Name: *json2go.BaseType,
		}
	}
}

func (s *StructGenerator) renamePaths() {
	for path, tag := range s.data {
		pathElements := strings.Split(path, ".")
		if len(pathElements) > 1 {
			if tag.StructType {
				s.data[pathElements[len(pathElements)-1]] = &fieldData.FieldData{LastSeenTimestamp: tag.LastSeenTimestamp, StructType: true}
			}
			tag.StructType = false
			s.data[pathElements[len(pathElements)-2]+"."+pathElements[len(pathElements)-1]] = tag
			if path != pathElements[len(pathElements)-2]+"."+pathElements[len(pathElements)-1] {
				delete(s.data, path)
			}
		}
	}
	return
}

/*
func CombineStructsOfFile(file, file1 *ast.File) (*ast.File, error) {
	var foundNodes []*AstUtils.FoundNodes
	var completed = false
	foundStructs := map[string][]*ast.StructType{}
	AstUtils.SearchNodes(file, &foundNodes, []*ast.Node{}, func(n *ast.Node, parents []*ast.Node, completed *bool) bool {
		if _, ok := (*n).(*ast.StructType); ok && len(parents) > 0 {
			if _, ok := (*parents[0]).(*ast.TypeSpec); ok {
				return true
			}
		}
		return false
	}, &completed)

	for _, node := range foundNodes {
		typeSpec := (*node.Parents[0]).(*ast.TypeSpec)
		ignoreStruct, err := shouldIgnoreStruct(typeSpec.Type.(*ast.StructType))
		if err != nil {
			return nil, err
		}
		if ignoreStruct {
			continue
		}
		foundStructs[typeSpec.Name.Name] = []*ast.StructType{typeSpec.Type.(*ast.StructType)}
	}

	var foundNodes1 []*AstUtils.FoundNodes
	var completed1 = false
	AstUtils.SearchNodes(file1, &foundNodes1, []*ast.Node{}, func(n *ast.Node, parents []*ast.Node, completed *bool) bool {
		if _, ok := (*n).(*ast.StructType); ok {
			return true
		}
		return false
	}, &completed1)

	for _, node := range foundNodes1 {
		typeSpec := (*node.Parents[0]).(*ast.TypeSpec)
		ignoreStruct, err := shouldIgnoreStruct(typeSpec.Type.(*ast.StructType))
		if err != nil {
			return nil, err
		}
		if ignoreStruct {
			continue
		}
		if _, ok := foundStructs[typeSpec.Name.Name]; !ok {
			foundStructs[typeSpec.Name.Name] = []*ast.StructType{typeSpec.Type.(*ast.StructType)}
		} else {
			foundStructs[typeSpec.Name.Name] = append(foundStructs[typeSpec.Name.Name], typeSpec.Type.(*ast.StructType))
		}
	}

	outFile, err := AstUtils.GetEmptyFile(file.Name.Name)
	if err != nil {
		return nil, err
	}

	for name, structs := range foundStructs {
		var fields []*ast.Field
		var err error
		if len(structs) == 1 {
			fields = structs[0].Fields.List
		} else {
			fields, err = combineStructFields(structs[0], structs[1], path)
			if err != nil {
				return nil, err
			}
		}

		outFile.Decls = append(outFile.Decls, &ast.GenDecl{
			Tok: token.TYPE,
			Specs: []ast.Spec{&ast.TypeSpec{
				Name: &ast.Ident{
					Name: name,
				},
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: fields,
					},
				},
			}},
		})
	}
	return outFile, nil
}

*/
