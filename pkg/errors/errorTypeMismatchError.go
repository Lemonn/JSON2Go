package errors

import "fmt"

type ErrorTypeMismatchError struct {
	JsonType  string
	ErrorType string
}

func (e *ErrorTypeMismatchError) Error() string {
	return fmt.Sprintf("type mismatch jsonType: %s, errorType: %s", e.JsonType, e.ErrorType)
}
