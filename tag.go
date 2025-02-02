package JSON2Go

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Lemonn/AstUtils"
	"go/ast"
	"go/token"
	"io"
	"strconv"
	"time"
)

type ParseFunctions struct {
	FromTypeParseFunction string `json:"fromTypeParseFunction,omitempty"`
	ToTypeParseFunction   string `json:"toTypeParseFunction,omitempty"`
}
type Tag struct {
	SeenValues              []string        `json:"seenValues,omitempty"`
	CheckedNonMatchingTypes []string        `json:"checkedNonMatchingTypes,omitempty"`
	ParseFunctions          *ParseFunctions `json:"parseFunctions,omitempty"`
	BaseType                *string         `json:"baseType,omitempty"`
	LastSeenTimestamp       int64           `json:"lastSeenTimestamp"`
}

func NewTagFromFieldData(fieldData interface{}) *Tag {
	var fieldValue string
	switch t := fieldData.(type) {
	case float64:
		fieldValue = strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		if t {
			fieldValue = "true"
		}
		fieldValue = "false"
	case string:
		fieldValue = fieldData.(string)
	}

	//TODO handle empty field value
	return &Tag{
		SeenValues:        []string{fieldValue},
		LastSeenTimestamp: time.Now().Unix(),
	}
}

func (j *Tag) ToTagString() (string, error) {
	value, err := j.ToTagValue()
	if err != nil {
		return "", err
	}
	return "json2go:\"" + value + "\" ", nil
}

func (j *Tag) ToTagValue() (string, error) {
	var err error
	var r []byte
	r, err = json.Marshal(j)
	if err != nil {
		return "", err
	}
	b64 := bytes.NewBuffer([]byte{})
	raw := base64.NewEncoder(base64.StdEncoding, b64)
	_, err = raw.Write(r)
	if err != nil {
		return "", err
	}
	err = raw.Close()
	if err != nil {
		return "", err
	}
	return b64.String(), nil
}

func (j *Tag) ToBasicLit() (*ast.BasicLit, error) {
	tagString, err := j.ToTagString()
	if err != nil {
		return nil, err
	}
	return &ast.BasicLit{
		Kind:  token.STRING,
		Value: tagString,
	}, nil
}

func (j *Tag) AppendToTag(tag *ast.BasicLit) (*ast.BasicLit, error) {
	lit, err := j.ToBasicLit()
	if err != nil {
		return nil, err
	}
	return combineTags(lit, tag)
}

func (j *Tag) Combine(j1 *Tag) (*Tag, error) {
	var jNew Tag

	//Combine BaseType
	if j.BaseType == nil && j1.BaseType == nil {
		jNew.BaseType = nil
	} else if j.BaseType != nil && j1.BaseType == nil {
		jNew.BaseType = j.BaseType
	} else if j.BaseType == nil && j1.BaseType != nil {
		jNew.BaseType = j1.BaseType
	} else if *j.BaseType != *j1.BaseType {
		return nil, errors.New(fmt.Sprintf("base type not equal %s:%s", *j.BaseType, *j1.BaseType))
	} else {
		jNew.BaseType = j1.BaseType
	}

	//Combine ParseFunction
	if j.ParseFunctions == nil && j1.ParseFunctions == nil {
		jNew.ParseFunctions = nil
	} else if j.ParseFunctions != nil && j1.ParseFunctions == nil {
		jNew.ParseFunctions = j.ParseFunctions
	} else if j.ParseFunctions == nil && j1.ParseFunctions != nil {
		jNew.ParseFunctions = j1.ParseFunctions
	} else if j.ParseFunctions.ToTypeParseFunction != j1.ParseFunctions.ToTypeParseFunction ||
		j.ParseFunctions.FromTypeParseFunction != j1.ParseFunctions.FromTypeParseFunction {
		return nil, errors.New("parse functions not equal")
	} else {
		jNew.ParseFunctions = j1.ParseFunctions
	}

	//Combine SeenValues
	values := make(map[string]struct{})
	jNew.SeenValues = []string{}
	if j.SeenValues != nil {
		for _, value := range j.SeenValues {
			values[value] = struct{}{}
		}
	}
	if j1.SeenValues != nil {
		for _, value := range j1.SeenValues {
			values[value] = struct{}{}
		}
	}
	for v, _ := range values {
		jNew.SeenValues = append(jNew.SeenValues, v)
	}

	//Combine NonMatchingTypes
	NonMatchingTypes := make(map[string]struct{})
	if j.CheckedNonMatchingTypes != nil {
		for _, nonMatchingType := range j.CheckedNonMatchingTypes {
			NonMatchingTypes[nonMatchingType] = struct{}{}
		}
	}
	if j1.CheckedNonMatchingTypes != nil {
		for _, nonMatchingType := range j1.CheckedNonMatchingTypes {
			NonMatchingTypes[nonMatchingType] = struct{}{}
		}
	}
	if NonMatchingTypes == nil {
		jNew.CheckedNonMatchingTypes = []string{}
	}
	for s, _ := range NonMatchingTypes {
		jNew.CheckedNonMatchingTypes = append(jNew.CheckedNonMatchingTypes, s)
	}

	//Combine LastSeen
	if j.LastSeenTimestamp > j1.LastSeenTimestamp {
		jNew.LastSeenTimestamp = j.LastSeenTimestamp
	} else {
		jNew.LastSeenTimestamp = j1.LastSeenTimestamp
	}

	return &jNew, nil
}

func GetJson2GoTag(tag string) (*Tag, error) {
	decoded := base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer([]byte(tag)))
	all, err := io.ReadAll(decoded)
	if err != nil {
		return nil, err
	}
	var json2GoTag Tag
	err = json.Unmarshal(all, &json2GoTag)
	if err != nil {
		return nil, err
	}
	return &json2GoTag, err
}

func GetJson2GoTagFromBasicLit(tag *ast.BasicLit) (*Tag, error) {
	var err error
	var json2GoTag *Tag
	keys := AstUtils.ExtractTagsByKey(tag)

	if v, ok := keys["json2go"]; !ok {
		return nil, nil
	} else {
		json2GoTag, err = GetJson2GoTag(v[0])
		if err != nil {
			return nil, err
		}
	}
	return json2GoTag, nil
}

type TagCombiner struct{}

func (j *TagCombiner) Combine(values []string) (string, error) {
	if len(values) == 0 {
		return "", nil
	} else if len(values) == 1 {
		return values[0], nil
	} else {
		tag0, err := GetJson2GoTag(values[0])
		if err != nil {
			return "", err
		}
		tag1, err := GetJson2GoTag(values[1])
		if err != nil {
			return "", err
		}
		combined, err := tag0.Combine(tag1)
		if err != nil {
			return "", err
		}
		return combined.ToTagValue()
	}
}

func combineTags(tag1, tag2 *ast.BasicLit) (*ast.BasicLit, error) {
	combiners := make(map[string]AstUtils.TagCombiner)
	combiners["json2go"] = &TagCombiner{}
	return AstUtils.CombineTags(tag1, tag2, combiners)
}

func resetToBasicType(field *ast.Field) (*ast.Field, error) {
	if field.Tag != nil {
		lit, err := GetJson2GoTagFromBasicLit(field.Tag)
		if err != nil {
			return nil, err
		}
		if lit != nil && lit.BaseType != nil {
			field.Type = &ast.Ident{
				Name: *lit.BaseType,
			}
		}
	}
	return field, nil
}
