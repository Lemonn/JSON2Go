package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

func (e *Example) UnmarshalJSON(bytes []byte) error {
	var data map[string]json.RawMessage
	var joinedErrors error
	err := json.Unmarshal(bytes, &data)
	if err != nil {
		return err
	}
	if value, ok := data["Name"]; ok {
		err = json.Unmarshal(value, &e.Name)
		if err != nil {
			var additionalElementsError *AdditionalElementsError
			if errors.As(err, &additionalElementsError) {
				joinedErrors = errors.Join(joinedErrors, additionalElementsError)
			} else {
				return err
			}
		}
		delete(data, "Name")
	}
	if value, ok := data["Year"]; ok {
		var unmarshalledValue float64
		err = json.Unmarshal(value, &unmarshalledValue)
		if err != nil {
			return err
		}
		e.Year, err = fromExampleYear(unmarshalledValue)
		if err != nil {
			return errors.Join(joinedErrors, err)
		}
		delete(data, "Year")
	}
	if value, ok := data["Pet"]; ok {
		err = json.Unmarshal(value, &e.Pet)
		if err != nil {
			var additionalElementsError *AdditionalElementsError
			if errors.As(err, &additionalElementsError) {
				joinedErrors = errors.Join(joinedErrors, additionalElementsError)
			} else {
				return err
			}
		}
		delete(data, "Pet")
	}
	if len(data) != 0 {
		joinedErrors = errors.Join(joinedErrors, &AdditionalElementsError{ParsedObj: "Example", Elements: data})
	}
	return joinedErrors
}
func CheckForFirstErrorNotOfTypeT[T error](errType T, e error) error {
UNWRAP:
	switch err := e.(type) {
	case interface {
		Unwrap() []error
	}:
		if !errors.As(err.Unwrap()[len(err.Unwrap())-1], &errType) {
			return err.Unwrap()[len(err.Unwrap())-1]
		}
		if len(err.Unwrap()) > 0 {
			e = err.Unwrap()[0]
			goto UNWRAP
		} else {
			return nil
		}
	default:
		if !errors.As(err, &errType) {
			return err
		} else {
			return nil
		}
	}
}

type Example struct {
	Name string `json:"Name"`
	Year int    `json:"Year"`
	Pet  string `json:"Pet"`
}

func fromExampleYear(baseValue float64) (int, error) {
	return int(baseValue), nil
}
func toExampleYear(baseValue int) (float64, error) {
	return float64(baseValue), nil
}
func (e *Example) MarshalJSON() ([]byte, error) {
	type localType struct {
		Name string  `json:"Name"`
		Year float64 `json:"Year"`
		Pet  string  `json:"Pet"`
	}
	var lt localType
	var err error
	lt.Name = e.Name
	lt.Year, err = toExampleYear(e.Year)
	if err != nil {
		return nil, err
	}
	lt.Pet = e.Pet
	return json.Marshal(lt)
}

type AdditionalElementsError struct {
	ParsedObj string
	Elements  map[string]json.RawMessage
}

func (j *AdditionalElementsError) String() string {
	return j.Error()
}
func (j *AdditionalElementsError) Error() string {
	m := "the following unexpected additional elements were found: "
	for s, e := range j.Elements {
		m += fmt.Sprintf("[(%s) RawJsonString {\"%s\": %s}]", s, s, e)
	}
	m += " whilst parsing " + j.ParsedObj
	return m
}
func GetAllErrorsOfType[T error](errType T, e error) ([]T, error) {
	var result []T
UNWRAP:
	switch err := e.(type) {
	case interface {
		Unwrap() []error
	}:
		if errors.As(err.Unwrap()[len(err.Unwrap())-1], &errType) {
			result = append([]T{errType}, result...)
		}
		if len(err.Unwrap()) > 0 {
			e = err.Unwrap()[0]
			goto UNWRAP
		} else {
			return result, nil
		}
	default:
		if len(result) > 0 {
			return result, nil
		} else if len(result) == 0 && errors.As(err, &errType) {
			result = append(result, errType)
			return result, nil
		} else {
			return nil, errors.New("error is not of joinError type and also error is not of searched type")
		}
	}
}
