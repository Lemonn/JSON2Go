package errors

// FieldPresenceChangeError is set whenever a field changes from/to omitempty
type FieldPresenceChangeError struct {
	NewState  string
	Timestamp int64
}

func (p FieldPresenceChangeError) Error() string {
	//TODO implement me
	panic("implement me")
}
