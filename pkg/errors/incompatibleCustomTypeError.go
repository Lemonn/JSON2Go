package errors

// IncompatibleCustomTypeError is emitted whenever a custom type stays the same but the underlying data make
// it incompatible. A prime example for this is the time.Time TypeChecker. The time type would stay the same,
// but the parse strings could become incompatible
type IncompatibleCustomTypeError struct {
	Timestamp int64
	Err       error
}

func (i *IncompatibleCustomTypeError) Error() string {
	//TODO implement me
	panic("implement me")
}
