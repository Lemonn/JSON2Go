package structGenerator

import (
	"fmt"
	"github.com/Lemonn/AstUtils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"go/token"
)

func (s *StructGenerator) attachJsonTags() {
	var foundNodes []*AstUtils.FoundNodes
	var completed bool
	s.tagOmittedFields()
	AstUtils.SearchNodes(s.file, &foundNodes, []*ast.Node{}, func(n *ast.Node, parents []*ast.Node, completed *bool) bool {
		if _, ok := (*n).(*ast.StructType); ok && len(parents) > 0 {
			return true
		}
		return false
	}, &completed)

	for i, node := range foundNodes {
		var path string
		for _, parent := range node.Parents {
			if v, ok := (*parent).(*ast.TypeSpec); ok {
				path = v.Name.Name
				break
			}
		}
		if s.data[path] == nil {
			continue
		}
		for ii, field := range (*node.Node).(*ast.StructType).Fields.List {
			var fData *fieldData.FieldData
			var ok bool
			if fData, ok = s.data[path+"."+field.Names[0].Name]; !ok {
				continue
			}
			var jsonTag string
			if fData.Omitempty {
				jsonTag = fmt.Sprintf("`json:\"%s,omitempty\"`", *fData.JsonFieldName)
			} else {
				jsonTag = fmt.Sprintf("`json:\"%s\"`", *fData.JsonFieldName)
			}
			(*foundNodes[i].Node).(*ast.StructType).Fields.List[ii].Tag = &ast.BasicLit{Kind: token.STRING, Value: jsonTag}
		}
	}
}

func (s *StructGenerator) tagOmittedFields() {
	for path, data := range s.data {
		if data.LastSeenTimestamp < s.startTime.Unix() {
			s.data[path].Omitempty = true
		}
	}
}
