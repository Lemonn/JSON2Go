package errors

import (
	"encoding/json"
	"github.com/Lemonn/JSON2Go/internal/utils"
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

func (e *StoreType) UnmarshalJSON(bytes []byte) error {
	localType := struct {
		Type string          `json:"type"`
		Err  json.RawMessage `json:"error"`
	}{}

	if err := json.Unmarshal(bytes, &localType); err != nil {
		return err
	}

	//IDEA we could write a generator that searches all relevant errors and adds them. This could be done in an extra
	// library, which does provide error storage code gen functions.
	//All error types that are stored inside the JSON-File need to be added here!
	switch localType.Type {
	case reflect.TypeOf(&typeChecker.NoLongerApplicableCustomTypeError{}).String():
		var te typeChecker.NoLongerApplicableCustomTypeError
		if unmarshallError := json.Unmarshal(localType.Err, &te); unmarshallError != nil {
			return unmarshallError
		}
		e.Err = &te
		e.Type = localType.Type
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
