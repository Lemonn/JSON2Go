package utils

import (
	"fmt"
	"github.com/Lemonn/JSON2Go/pkg/fieldData"
	"go/ast"
	"go/parser"
	"math"
	"strings"
)

type SeenTypeUtils struct {
	seenTypes map[string]*fieldData.PathData
}

func NewSeenTypeUtils(seenTypes map[string]*fieldData.PathData) *SeenTypeUtils {
	return &SeenTypeUtils{
		seenTypes: seenTypes,
	}
}

func (s *SeenTypeUtils) GetAdjustedFieldType(path string) (expr ast.Expr, err error) {
	if s.seenTypes[path].TypeAdjusterData != nil && s.seenTypes[path].TypeAdjusterData.ActiveType != nil {
		expr, err = parser.ParseExpr(*s.seenTypes[path].TypeAdjusterData.ActiveType)
		if err != nil {
			return nil, err
		}
		return GeneratedNestedArray(s.GetLevelOfArrays(path), expr), nil
	} else {
		return s.GetFieldType(path, false), nil
	}
}

func (s *SeenTypeUtils) GetFieldType(path string, withoutArray bool) (expr ast.Expr) {
	levelOfArrays := math.MaxInt32
	if withoutArray {
		levelOfArrays = 0
	}
	//TODO error on path not found
	if s.seenTypes[path].ForceSourceType != nil {
		parseExpr, err := parser.ParseExpr(*s.seenTypes[path].ForceSourceType)
		if err != nil {
			return nil
		}
		return parseExpr
	}

	if len(s.seenTypes[path].Types) == 1 || s.EmptySubtype(path) || s.PointerType(path) {
		Type := s.GetType(path)
		if len(s.seenTypes[path].Types[Type]) == 1 {
			if !withoutArray {
				for levelOfArrays, _ = range s.seenTypes[path].Types[Type] {
					break
				}
			}
			pathElements := strings.Split(path, ".")
			if Type == fieldData.Field {
				expr = GeneratedNestedArray(levelOfArrays, &ast.StarExpr{X: &ast.SelectorExpr{X: &ast.Ident{Name: pathElements[len(pathElements)-1]}, Sel: &ast.Ident{Name: pathElements[len(pathElements)-1]}}})
			} else if Type == fieldData.EmptyArray {
				expr = GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
			} else if Type == fieldData.EmptyStruct {
				expr = GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
			} else if Type == fieldData.Null {
				expr = GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
			} else {
				expr = GeneratedNestedArray(levelOfArrays, &ast.Ident{Name: string(Type)})
				if s.PointerType(path) {
					expr = &ast.StarExpr{X: expr}
				}
			}
		} else {
			if !withoutArray {
				for i, _ := range s.seenTypes[path].Types[Type] {
					if levelOfArrays > i {
						levelOfArrays = i
					}
				}
			}

			expr = GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
		}
	} else {
		if !withoutArray {
			for _, m := range s.seenTypes[path].Types {
				for i, _ := range m {
					if levelOfArrays > i {
						levelOfArrays = i
					}
				}
			}
		}
		expr = GeneratedNestedArray(levelOfArrays, &ast.InterfaceType{Methods: &ast.FieldList{}})
	}
	return expr
}

func (s *SeenTypeUtils) IsBasicType(path string) bool {
	b, _, _ := s.IsBasicTypeWhitDetails(path)
	return b
}

func (s *SeenTypeUtils) IsBasicTypeWhitDetails(path string) (bool, int, fieldData.Type) {
	var setType *fieldData.Type
	var setLevel *int
	if s.seenTypes[path].Types[fieldData.EmptyStruct] != nil || s.seenTypes[path].Types[fieldData.Field] != nil {
		return false, 0, fieldData.Unsupported
	}
	for Type, levels := range s.seenTypes[path].Types {
		if !(Type == fieldData.Null || Type == fieldData.EmptyArray) && setType != nil && *setType != Type {
			return false, 0, fieldData.Unsupported
		} else if !(Type == fieldData.Null || Type == fieldData.EmptyArray) && setType == nil {
			setType = &Type
		}
		for level, _ := range levels {
			if setLevel == nil {
				setLevel = &level
			} else if *setLevel != level {
				return false, 0, fieldData.Unsupported
			}
		}

	}
	if setType != nil && setLevel != nil {
		return true, *setLevel, *setType
	}
	return false, 0, fieldData.Unsupported
}

func (s *SeenTypeUtils) GetType(path string) fieldData.Type {
	for Type, _ := range s.seenTypes[path].Types {
		if len(s.seenTypes[path].Types) == 1 {
			return Type
		} else if Type != fieldData.EmptyArray && Type != fieldData.EmptyStruct && Type != fieldData.Null {
			return Type
		}
	}
	return fieldData.Unsupported
}

func (s *SeenTypeUtils) ContainsEmptyType(path string) bool {
	if _, ok := s.seenTypes[path].Types[fieldData.EmptyStruct]; ok {
		return true
	} else if _, ok := s.seenTypes[path].Types[fieldData.EmptyArray]; ok {
		return true
	}
	return false
}

func (s *SeenTypeUtils) ContainsPointerType(path string) bool {
	if _, ok := s.seenTypes[path].Types[fieldData.Null]; ok {
		return true
	}
	return false
}

func (s *SeenTypeUtils) PointerType(path string) bool {
	var PointerType bool
	var Level int
	if len(s.seenTypes[path].Types) == 2 || (len(s.seenTypes[path].Types) == 3 && s.ContainsEmptyType(path)) {
		if _, ok := s.seenTypes[path].Types[fieldData.Field]; ok {
			if v, ok := s.seenTypes[path].Types[fieldData.Null]; ok {
				for Level, _ = range s.seenTypes[path].Types[fieldData.Field] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					PointerType = true
				}

			}
		} else if _, ok := s.seenTypes[path].Types[fieldData.String]; ok {
			if v, ok := s.seenTypes[path].Types[fieldData.Null]; ok {
				for Level, _ = range s.seenTypes[path].Types[fieldData.String] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					PointerType = true
				}
			}
		} else if _, ok := s.seenTypes[path].Types[fieldData.Float64]; ok {
			if v, ok := s.seenTypes[path].Types[fieldData.Null]; ok {
				for Level, _ = range s.seenTypes[path].Types[fieldData.Float64] {
					break
				}
				if _, ok := v[Level]; ok && len(v) == 1 {
					PointerType = true
				}
			}
		} else if _, ok := s.seenTypes[path].Types[fieldData.Bool]; ok {
			if v, ok := s.seenTypes[path].Types[fieldData.Null]; ok {
				for Level, _ = range s.seenTypes[path].Types[fieldData.Bool] {
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

func (s *SeenTypeUtils) EmptySubtype(path string) bool {
	var EmptySubtype bool
	var Level int
	if len(s.seenTypes[path].Types) == 2 || (len(s.seenTypes[path].Types) == 3 && s.ContainsPointerType(path)) {
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

func (s *SeenTypeUtils) IsStruct(path string) bool {
	if len(s.seenTypes[path].Types) == 1 || s.EmptySubtype(path) || s.PointerType(path) {
		Type := s.GetType(path)
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

func (s *SeenTypeUtils) GetLevelOfArrays(path string) int {
	levelOfArrays := math.MaxInt32
	for Type, _ := range s.seenTypes[path].Types {
		for i, _ := range s.seenTypes[path].Types[Type] {
			if levelOfArrays > i {
				levelOfArrays = i
			}
		}
	}
	return levelOfArrays
}

func (s *SeenTypeUtils) GetMaxFieldCount(path string) int {
	var highestValue int
	for _, levels := range s.seenTypes[path].Types {
		for _, values := range levels {
			for _, details := range values {
				if highestValue < details.Count {
					highestValue = details.Count

				}
			}
		}
	}
	return highestValue
}

func (s *SeenTypeUtils) GetParentPath(path string) string {
	pathElements := strings.Split(path, ".")
	if len(pathElements) > 1 {
		return strings.Join(pathElements[:len(pathElements)-1], ".")
	} else {
		return path
	}
}

func (s *SeenTypeUtils) IsRootPath(path string) bool {
	if len(strings.Split(path, ".")) == 1 {
		return true
	}
	return false
}

func (s *SeenTypeUtils) Omitempty(path string) bool {
	if s.IsRootPath(path) {
		return false
	} else {
		parentPath := s.GetParentPath(path)
		LevelOfArrays := s.GetLevelOfArrays(path)
		if _, ok := s.seenTypes[parentPath].Types[fieldData.EmptyStruct]; ok {
			if _, ok := s.seenTypes[parentPath].Types[fieldData.EmptyStruct][LevelOfArrays]; ok {
				return false
			}
		} else if _, ok := s.seenTypes[path].Types[fieldData.EmptyArray]; ok {
			if _, ok := s.seenTypes[path].Types[fieldData.EmptyArray][LevelOfArrays]; ok {
				return false
			}
		} else if _, ok := s.seenTypes[path].Types[fieldData.Null]; ok {
			if _, ok := s.seenTypes[path].Types[fieldData.Null][LevelOfArrays]; ok {
				return false
			}
		}

		if s.seenTypes[path].SeenCounter < s.seenTypes[parentPath].SeenCounter {
			return true
		} else {
			return false
		}

	}
}

func (s *SeenTypeUtils) GetJsonTag(path string) string {
	return fmt.Sprintf("`json:\"%s%s\"`", s.seenTypes[path].JsonFieldName, func() string {
		if s.Omitempty(path) {
			return ",omitempty"
		} else {
			return ""
		}
	}())
}
