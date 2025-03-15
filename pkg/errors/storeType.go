package errors

import (
	"encoding/json"
	"github.com/Lemonn/JSON2Go/internal/utils"
	"github.com/Lemonn/JSON2Go/pkg/errors/combiner"
	"github.com/Lemonn/JSON2Go/pkg/errors/typeChecker"
	"reflect"
)

type StoreType struct {
	Type string `json:"type"`
	Err  error  `json:"error"`
}

func NewStoreType(err error) *StoreType {
	return &StoreType{
		Type: reflect.TypeOf(err).String(),
		Err:  err,
	}
}

func NewStoreTypeArray(err error) []*StoreType {
	var st []*StoreType
	es := utils.GetAllWrappedErrors(err)
	for _, innerErr := range es {
		st = append(st, NewStoreType(innerErr))
	}
	return st
}

//go:generate go run ../../internal/aiGen/fileErrorGenerator/generator.go
func (e *StoreType) UnmarshalJSON(bytes []byte) error {
	localType :=
		struct {
			Type string          `json:"type"`
			Err  json.RawMessage `json:"error"`
		}{}
	if err := json.
		Unmarshal(bytes, &localType); err != nil {
		return err
	}
	switch localType.
		Type {
	case reflect.TypeOf(&combiner.
		ConflictingJsonFieldNameError{}).String():
		var te combiner.
			ConflictingJsonFieldNameError

		if unmarshallError := json.Unmarshal(localType.Err, &te); unmarshallError !=
			nil {
			return unmarshallError
		}
		e.Err = &te
		e.Type = localType.Type
	case reflect.TypeOf(&typeChecker.IncompatibleCustomTypeError{}).String():
		var te typeChecker.IncompatibleCustomTypeError

		if unmarshallError := json.Unmarshal(localType.Err, &te); unmarshallError !=
			nil {
			return unmarshallError
		}
		e.
			Err = &te

		e.Type = localType.Type
	case reflect.TypeOf(&combiner.AmbiguousFieldPresenceError{}).String():
		var te combiner.AmbiguousFieldPresenceError
		if unmarshallError := json.Unmarshal(localType.Err, &te); unmarshallError != nil {
			return unmarshallError

		}
		e.Err = &te
		e.Type = localType.Type
	case reflect.TypeOf(&combiner.FieldPresenceChangeError{}).String():
		var te combiner.
			FieldPresenceChangeError
		if unmarshallError := json.Unmarshal(localType.Err,
			&te); unmarshallError != nil {
			return unmarshallError
		}
		e.Err = &te
		e.Type = localType.Type
	case reflect.TypeOf(&combiner.TypeChangeError{}).String():
		var te combiner.TypeChangeError
		if unmarshallError := json.Unmarshal(localType.
			Err, &te,
		); unmarshallError != nil {
			return unmarshallError
		}
		e.Err = &te
		e.Type = localType.Type
	case reflect.TypeOf(
		&typeChecker.TypeExpansionError{}).String():
		var te typeChecker.TypeExpansionError
		if unmarshallError := json.Unmarshal(localType.Err, &te); unmarshallError != nil {
			return unmarshallError
		}
		e.Err = &te
		e.Type = localType.Type
	case reflect.TypeOf(&combiner.ConflictingCustomTypesError{}).String():
		var te combiner.
			ConflictingCustomTypesError
		if unmarshallError := json.Unmarshal(localType.Err, &te); unmarshallError !=
			nil {
			return unmarshallError
		}
		e.Err = &te
		e.Type = localType.Type
	case reflect.TypeOf(&typeChecker.NoLongerApplicableCustomTypeError{}).String():
		var te typeChecker.NoLongerApplicableCustomTypeError
		if unmarshallError := json.Unmarshal(localType.Err, &te); unmarshallError != nil {
			return unmarshallError
		}
		e.Err = &te
		e.Type =
			localType.
				Type
	default:
		return &UnknownStoreTypeError{TypeName: localType.Type}
	}
	return nil
}

func (e *StoreType) MarshalJSON() ([]byte, error) {
	localType := struct {
		Type string `json:"type"`
		Err  error  `json:"error"`
	}{
		Type: reflect.TypeOf(e.Err).String(),
		Err:  e.Err,
	}
	return json.Marshal(localType)
}
