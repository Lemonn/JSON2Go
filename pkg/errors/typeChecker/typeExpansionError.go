package typeChecker

import "time"

// TypeExpansionError Is emitted whenever a custom type gets expanded. An example is when an enum gets an additional field.
// This usually should be handled as a minor version change.
type TypeExpansionError struct {
	Timestamp int64
	Path      string
}

func (e *TypeExpansionError) SetTimestamp(t time.Time) {
	e.Timestamp = t.UnixNano()
}

func (e *TypeExpansionError) GetTimestamp() time.Time {
	return time.Unix(e.Timestamp, 0)
}

func (e *TypeExpansionError) IsFileError() {}

func (e *TypeExpansionError) Error() string {
	return "type expansion error"
}
