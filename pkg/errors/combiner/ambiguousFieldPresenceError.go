package combiner

import (
	"fmt"
	"time"
)

// AmbiguousFieldPresenceError Is emitted whenever the omitempty status of a field is ambiguous.
// This is the case if [],{} or null values are present, but the field is omitted nevertheless.
// This is non-relevant for read only cases, it only matters if one wants to send generated data to an API.
// Then this error could indicate a complex field dependency (Field X determines the default value of field Y)
type AmbiguousFieldPresenceError struct {
	Timestamp int64
	Path      string
}

func (e *AmbiguousFieldPresenceError) SetTimestamp(t time.Time) {
	e.Timestamp = t.UnixNano()
}

func (e *AmbiguousFieldPresenceError) GetTimestamp() time.Time {
	return time.Unix(e.Timestamp, 0)
}

func (e *AmbiguousFieldPresenceError) IsFileError() {}

func (e *AmbiguousFieldPresenceError) Error() string {
	return fmt.Sprintf("the json omit state of the type at %s, is ambiguous. This is due to the fact, that "+
		"the type contains [], {} or null but also seems to ommited fields.", e.Path)
}
