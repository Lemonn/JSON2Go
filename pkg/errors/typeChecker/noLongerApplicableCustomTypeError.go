package typeChecker

import (
	"time"
)

type NoLongerApplicableCustomTypeError struct {
	Timestamp         int64  `json:"timestamp"`
	ActiveTypeChecker string `json:"activeTypeChecker"`
}

func (e *NoLongerApplicableCustomTypeError) SetTimestamp(t time.Time) {
	e.Timestamp = t.Unix()
}

func (e *NoLongerApplicableCustomTypeError) GetTimestamp() time.Time {
	return time.Unix(e.Timestamp, 0)
}

func (e *NoLongerApplicableCustomTypeError) IsFileError() {}

func (e *NoLongerApplicableCustomTypeError) Error() string {
	return ""
}
