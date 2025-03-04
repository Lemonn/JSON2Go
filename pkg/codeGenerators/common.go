package codeGenerators

import (
	"fmt"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"go/parser"
	"math"
	"strings"
)

type Common struct {
	fileData fieldData.FileData
}

func NewCommon(fileData fieldData.FileData) *Common {
	return &Common{
		fileData: fileData,
	}
}

func (b *Common) SetActiveTypeFile(fileData fieldData.FileData) {
	b.fileData = fileData
}

func (b *Common) IsPointer(path string) bool {
	var PointerType bool
	var Level int
	if len(b.fileData[path].Types) == 2 || (len(b.fileData[path].Types) == 3 && b.ContainsEmptyType(path)) {
		if _, ok := b.fileData[path].Types[fieldData.Field]; ok {
			if v, ok := b.fileData[path].Types[fieldData.Null]; ok {
				for Level, _ = range b.fileData[path].Types[fieldData.Field] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					PointerType = true
				}

			}
		} else if _, ok := b.fileData[path].Types[fieldData.String]; ok {
			if v, ok := b.fileData[path].Types[fieldData.Null]; ok {
				for Level, _ = range b.fileData[path].Types[fieldData.String] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					PointerType = true
				}
			}
		} else if _, ok := b.fileData[path].Types[fieldData.Float64]; ok {
			if v, ok := b.fileData[path].Types[fieldData.Null]; ok {
				for Level, _ = range b.fileData[path].Types[fieldData.Float64] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					PointerType = true
				}
			}
		} else if _, ok := b.fileData[path].Types[fieldData.Bool]; ok {
			if v, ok := b.fileData[path].Types[fieldData.Null]; ok {
				for Level, _ = range b.fileData[path].Types[fieldData.Bool] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					PointerType = true
				}
			}
		}
	}
	return PointerType
}

func (b *Common) EmptySubtype(path string) bool {
	var EmptySubtype bool
	var Level int
	pathData := b.fileData[path]
	if len(pathData.Types) == 2 || (len(pathData.Types) == 3 && b.ContainsPointerType(path)) {
		if _, ok := pathData.Types[fieldData.Field]; ok {
			if v, ok := pathData.Types[fieldData.EmptyArray]; ok {
				for Level, _ = range pathData.Types[fieldData.Field] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					EmptySubtype = true
				}

			} else if _, ok := pathData.Types[fieldData.EmptyStruct]; ok {
				for Level, _ = range pathData.Types[fieldData.EmptyStruct] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					EmptySubtype = true
				}
			}
		} else if _, ok := pathData.Types[fieldData.String]; ok {
			if v, ok := pathData.Types[fieldData.EmptyArray]; ok {
				for Level, _ = range pathData.Types[fieldData.String] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					EmptySubtype = true
				}
			}
		} else if _, ok := pathData.Types[fieldData.Float64]; ok {
			if v, ok := pathData.Types[fieldData.EmptyArray]; ok {
				for Level, _ = range pathData.Types[fieldData.Float64] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					EmptySubtype = true
				}
			}
		} else if _, ok := pathData.Types[fieldData.Bool]; ok {
			if v, ok := pathData.Types[fieldData.EmptyArray]; ok {
				for Level, _ = range pathData.Types[fieldData.Bool] {
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

func (b *Common) GetFieldType(path string, withoutArray bool) (expr ast.Expr) {
	pathData := b.fileData[path]
	levelOfArrays := math.MaxInt32
	if withoutArray {
		levelOfArrays = 0
	}
	//TODO error on path not found
	if pathData.ForceSourceType != nil {
		parseExpr, err := parser.ParseExpr(*pathData.ForceSourceType)
		if err != nil {
			return nil
		}
		return parseExpr
	}

	if len(pathData.Types) == 1 || b.EmptySubtype(path) || b.IsPointer(path) {
		Type := b.GetType(path)
		if len(pathData.Types[Type]) == 1 {
			if !withoutArray {
				for levelOfArrays, _ = range b.fileData[path].Types[Type] {
					break
				}
			}

			if Type == fieldData.Field {
				expr = utils.GeneratedNestedArray(levelOfArrays, &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: utils.GetFieldName(path)}, Sel: &ast.Ident{Name: utils.GetFieldName(path)}}})
			} else if Type == fieldData.EmptyArray {
				expr = utils.GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
			} else if Type == fieldData.EmptyStruct {
				expr = utils.GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
			} else if Type == fieldData.Null {
				expr = utils.GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
			} else {
				expr = utils.GeneratedNestedArray(levelOfArrays, &ast.Ident{Name: string(Type)})
				if b.IsPointer(path) {
					expr = &ast.StarExpr{X: expr}
				}
			}
		} else {
			if !withoutArray {
				for i, _ := range pathData.Types[Type] {
					if levelOfArrays > i {
						levelOfArrays = i
					}
				}
			}

			expr = utils.GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
		}
	} else {
		if !withoutArray {
			for _, m := range pathData.Types {
				for i, _ := range m {
					if levelOfArrays > i {
						levelOfArrays = i
					}
				}
			}
		}
		expr = utils.GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
	}
	return expr
}

func (b *Common) GetAdjustedFieldType(path string) (expr ast.Expr, err error) {
	if b.fileData[path].TypeAdjusterData != nil && b.fileData[path].ActiveType != nil {
		expr, err = parser.ParseExpr(*b.fileData[path].ActiveType)
		if err != nil {
			return nil, err
		}
		return utils.GeneratedNestedArray(b.GetLevelOfArrays(path), expr), nil
	} else {
		return b.GetFieldType(path, false), nil
	}
}

func (b *Common) IsStruct(path string) bool {
	if len(b.fileData[path].Types) == 1 || b.EmptySubtype(path) || b.IsPointer(path) {
		Type := b.GetType(path)
		if len(b.fileData[path].Types[Type]) == 1 {
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

func (b *Common) GetLevelOfArrays(path string) int {
	levelOfArrays := math.MaxInt32
	for Type, _ := range b.fileData[path].Types {
		for i, _ := range b.fileData[path].Types[Type] {
			if levelOfArrays > i {
				levelOfArrays = i
			}
		}
	}
	return levelOfArrays
}

func (b *Common) ContainsEmptyType(path string) bool {
	if _, ok := b.fileData[path].Types[fieldData.EmptyStruct]; ok {
		return true
	} else if _, ok := b.fileData[path].Types[fieldData.EmptyArray]; ok {
		return true
	}
	return false
}

func (b *Common) GetTypeByPath(path string) fieldData.Type {
	for Type, _ := range b.fileData[path].Types {
		if len(b.fileData[path].Types) == 1 {
			return Type
		} else if Type != fieldData.EmptyArray && Type != fieldData.EmptyStruct && Type != fieldData.Null {
			return Type
		}
	}
	return fieldData.Unsupported
}

func (b *Common) GetType(path string) fieldData.Type {
	pathData := b.fileData[path]
	for Type, _ := range pathData.Types {
		if len(pathData.Types) == 1 {
			return Type
		} else if Type != fieldData.EmptyArray && Type != fieldData.EmptyStruct && Type != fieldData.Null {
			return Type
		}
	}
	return fieldData.Unsupported
}

func (b *Common) ContainsPointerType(path string) bool {
	if _, ok := b.fileData[path].Types[fieldData.Null]; ok {
		return true
	}
	return false
}

func (b *Common) GetParentPath(path string) string {
	pathElements := strings.Split(path, ".")
	if len(pathElements) > 1 {
		return strings.Join(pathElements[:len(pathElements)-1], ".")
	} else {
		return path
	}
}

func (b *Common) IsRootPath(path string) bool {
	if len(strings.Split(path, ".")) == 1 {
		return true
	}
	return false
}

func (b *Common) Omitempty(path string) bool {
	if b.IsRootPath(path) {
		return false
	} else {
		parentPath := b.GetParentPath(path)
		LevelOfArrays := b.GetLevelOfArrays(path)
		if _, ok := b.fileData[parentPath].Types[fieldData.EmptyStruct]; ok {
			if _, ok := b.fileData[parentPath].Types[fieldData.EmptyStruct][LevelOfArrays]; ok {
				return false
			}
		} else if _, ok := b.fileData[path].Types[fieldData.EmptyArray]; ok {
			if _, ok := b.fileData[path].Types[fieldData.EmptyArray][LevelOfArrays]; ok {
				return false
			}
		} else if _, ok := b.fileData[path].Types[fieldData.Null]; ok {
			if _, ok := b.fileData[path].Types[fieldData.Null][LevelOfArrays]; ok {
				return false
			}
		}

		if b.fileData[path].SeenCounter+b.fileData[path].IntroductionCount < b.fileData[parentPath].SeenCounter {
			return true
		} else {
			return false
		}

	}
}

func (b *Common) GetJsonTag(path string) string {
	return fmt.Sprintf("`json:\"%s%s\"`", b.fileData[path].JsonFieldName, func() string {
		if b.Omitempty(path) {
			return ",omitempty"
		} else {
			return ""
		}
	}())
}
