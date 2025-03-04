package errors

import (
	"fmt"
	"time"
)

type ConflictingJsonFieldNameError struct {
	Timestamp    int64
	OldFieldName string `json:"oldFieldName"`
	NewFieldName string `json:"newFieldName"`
}

func (e *ConflictingJsonFieldNameError) SetTimestamp(t time.Time) {
	e.Timestamp = t.UnixNano()
}

func (e *ConflictingJsonFieldNameError) GetTimestamp() time.Time {
	return time.Unix(e.Timestamp, 0)
}

func (e *ConflictingJsonFieldNameError) IsFileError() {}

func (e *ConflictingJsonFieldNameError) Error() string {
	return fmt.Sprintf("conflicting JSON fieldnames, oldName: %s, newName: %s", e.OldFieldName, e.NewFieldName)
}
