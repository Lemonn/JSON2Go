package fieldData

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type Metadata struct {
	TotalSampleCount int                   `json:"totalSampleCount"`
	LastRunTimestamp int64                 `json:"lastRunTimestamp"`
	Data             map[string]*FieldData `json:"data"`
}
type FieldData struct {
	// SeenValues Holds all seen field values and their corresponding type
	SeenValues map[string]string `json:"seenValues,omitempty"`
	// CheckedNonMatchingTypes Is used to store a map of non-matching types referencing the time of storage as unix timestamp.
	// This is useful when working with a lot of input data, and the seen values becomes to big
	CheckedNonMatchingTypes map[string]int64 `json:"checkedNonMatchingTypes,omitempty"`
	// ParseFunctions Holds the names of the parse functions generated by TypeDeterminationFunction
	ParseFunctions *ParseFunctions `json:"parseFunctions,omitempty"`
	// BaseType The type before it was modified from a TypeDeterminationFunction
	BaseType *string `json:"baseType,omitempty"`
	// CurrentCustomType Set whenever a custom type is in use. For example is set to time.Time for the time type
	CurrentCustomType *string `json:"currentCustomType,omitempty"`
	// MixedTypes is set, whenever an array whit mixed types is encountered
	MixedTypes bool `json:"mixedTypes,omitempty"`
	// LastSeenTimestamp Unix timestamp. Is updated whenever a value is seen
	LastSeenTimestamp int64 `json:"lastSeenTimestamp"`
	// EmptyValuePresent Is the opposite of omitempty, and is set if the field contains empty values
	EmptyValuePresent bool `json:"emptyValuePresent,omitempty"`
	// Contains the name, as found in the JSON-File
	JsonFieldName *string `json:"jsonFieldName,omitempty"`
	// NameOfActiveTypeAdjuster Name of the TypeAdjuster that replaced the type
	NameOfActiveTypeAdjuster *string `json:"nameOfActiveTypeAdjuster,omitempty"`
	// TypeAdjusterData Data stored by the currently in used TypeAdjuster. This is used to emmit the
	// IncompatibleCustomType error, for example if a type such as time.Time stays the same but the underling
	// time strings are incompatible
	TypeAdjusterData json.RawMessage `json:"typeAdjusterData,omitempty"`
	// StructType Set whenever only data of a struct is stored.
	StructType bool `json:"structType,omitempty"`
	// RequiredField set to true, if the unmarshall generator should make this a required field.
	RequiredField bool `json:"requiredField,omitempty"`
}

// ParseFunctions Holds the names of the parse functions
type ParseFunctions struct {
	// FromTypeParseFunction Holds the function name, which converts from json to custom type
	FromTypeParseFunction string `json:"fromTypeParseFunction,omitempty"`
	// ToTypeParseFunction Holds the function name, which converts from custom to json type
	ToTypeParseFunction string `json:"toTypeParseFunction,omitempty"`
}

func (j *FieldData) Combine(j1 *FieldData) (*FieldData, error) {
	var jNew FieldData
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
	values := make(map[string]string)
	if j.SeenValues != nil {
		for value, FieldType := range j.SeenValues {
			values[value] = FieldType
		}
	}
	if j1.SeenValues != nil {
		for value, FieldType := range j1.SeenValues {
			values[value] = FieldType
		}
	}
	jNew.SeenValues = values

	//Combine NonMatchingTypes
	NonMatchingTypes := make(map[string]int64)
	if j.CheckedNonMatchingTypes != nil {
		for key, value := range j.CheckedNonMatchingTypes {
			NonMatchingTypes[key] = value
		}
	}
	if j1.CheckedNonMatchingTypes != nil {
		for key, value := range j1.CheckedNonMatchingTypes {
			NonMatchingTypes[key] = value
		}
	}
	jNew.CheckedNonMatchingTypes = NonMatchingTypes

	//Combine LastSeen
	if j.LastSeenTimestamp > j1.LastSeenTimestamp {
		jNew.LastSeenTimestamp = j.LastSeenTimestamp
	} else {
		jNew.LastSeenTimestamp = j1.LastSeenTimestamp
	}

	//Combine MixedTypes
	if j.MixedTypes || j1.MixedTypes {
		jNew.MixedTypes = true
	} else {
		jNew.MixedTypes = false
	}

	// Combine EmptyValuePresent
	if j.EmptyValuePresent || j1.EmptyValuePresent {
		jNew.EmptyValuePresent = true
	} else {
		jNew.EmptyValuePresent = false
	}

	//Combine JsonFieldName
	if j.JsonFieldName != nil {
		jNew.JsonFieldName = j.JsonFieldName
	} else {
		jNew.JsonFieldName = j1.JsonFieldName
	}

	//Combine StructType
	if j.StructType || j1.StructType {
		jNew.StructType = true
	} else {
		jNew.StructType = false
	}

	//TODO combine the missing types

	return &jNew, nil
}

func NewTagFromFieldData(fieldData interface{}) (*FieldData, error) {
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
	default:
		return nil, errors.New(fmt.Sprintf("unsupported type of field data: %T", fieldData))
	}

	return &FieldData{
		SeenValues:        map[string]string{fieldValue: reflect.TypeOf(fieldData).String()},
		LastSeenTimestamp: time.Now().Unix(),
	}, nil
}

func SetOrCombineFieldData(data *FieldData, tags map[string]*FieldData, path string) error {
	if v, ok := tags[path]; ok {
		combine, err := data.Combine(v)
		if err != nil {
			return err
		}
		tags[path] = combine
	} else {
		tags[path] = data
	}
	return nil
}
