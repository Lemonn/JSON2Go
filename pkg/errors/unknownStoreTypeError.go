package errors

import "fmt"

type UnknownStoreTypeError struct {
	TypeName string
}

func (u *UnknownStoreTypeError) Error() string {
	return fmt.Sprintf("unknown store type %s. Please add all types that should be stores to the StoreType ummarshall function!", u.TypeName)
}
