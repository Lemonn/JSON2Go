package fieldData

import (
	"errors"
	"fmt"
	"reflect"
)

type Type string

func (t Type) String() string {
	return string(t)
}

const (
	String      Type = "string"
	Float64     Type = "float64"
	Bool        Type = "bool"
	Field       Type = "field"
	EmptyArray  Type = "emptyArray"
	EmptyStruct Type = "emptyStruct"
	Null        Type = "null"
	Unsupported Type = "unsupported"
)

func TypeFromAny(a any) (Type, error) {
	t := reflect.TypeOf(a)
	if t == reflect.TypeOf(nil) {
		return Null, nil
	} else {
		return NewType(reflect.TypeOf(a).String())
	}
}

func NewType(s string) (Type, error) {
	switch s {
	case "string":
		return String, nil
	case "float64":
		return Float64, nil
	case "bool":
		return Bool, nil
	case "field":
		return Field, nil
	case "emptyArray":
		return EmptyArray, nil
	case "emptyStruct":
		return EmptyStruct, nil
	case "null":
		return Null, nil

	default:
		return Unsupported, errors.New(fmt.Sprintf("encountred an unsupported type: %s", s))
	}
}
