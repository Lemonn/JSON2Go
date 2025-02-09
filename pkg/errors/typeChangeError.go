package errors

// TypeChangeError is emitted whenever a filed-type changed in  a way that incompatible to the old type
type TypeChangeError struct {
	Path    string
	OldType string
	NewType string
	// InterfaceTypeReplacement is set, whenever a previously unknown datatype becomes a known one
	InterfaceTypeReplacement bool
	Err                      error
}

func (t TypeChangeError) Error() string {
	//TODO implement me
	panic("implement me")
}
