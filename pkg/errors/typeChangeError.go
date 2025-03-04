package errors

import "time"

// TypeChangeError is emitted whenever a filed-type changed in a way that's incompatible to the old type.
// If seen, this is equal of a major version change, and should therefore be handled as such.
type TypeChangeError struct {
	// Timestamp Unix timestamp of the combination time.
	Timestamp     int64
	OldType       string
	NewType       string
	ResultingType string
	// InterfaceTypeReplacement is set, whenever a previously unknown datatype becomes a known one
	InterfaceTypeReplacement bool
	// WasCustomType is set whenever the previously used type was a custom one
	WasCustomType bool
	// IsCustomType is set, if the currently used type is a custom one
	IsCustomType bool
}

func (e *TypeChangeError) SetTimestamp(t time.Time) {
	e.Timestamp = t.UnixNano()
}

func (e *TypeChangeError) GetTimestamp() time.Time {
	return time.Unix(e.Timestamp, 0)
}

func (e *TypeChangeError) IsFileError() {}

func (e *TypeChangeError) Error() string {
	//TODO implement me
	panic("implement me")
}
