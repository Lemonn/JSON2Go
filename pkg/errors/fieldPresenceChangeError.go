package errors

type FieldPresenceChangeError struct {
	NewState  string
	Timestamp int64
}

func (p FieldPresenceChangeError) Error() string {
	//TODO implement me
	panic("implement me")
}
