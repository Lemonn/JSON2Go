package errors

import "time"

// FieldPresenceChangeError is set whenever a field changes from/to omitempty
type FieldPresenceChangeError struct {
	Timestamp int64
	NewState  string
}

func (e *FieldPresenceChangeError) SetTimestamp(t time.Time) {
	e.Timestamp = t.UnixNano()
}

func (e *FieldPresenceChangeError) GetTimestamp() time.Time {
	return time.Unix(e.Timestamp, 0)
}

func (e *FieldPresenceChangeError) IsFileError() {}

func (e *FieldPresenceChangeError) Error() string {
	//TODO implement me
	panic("implement me")
}
