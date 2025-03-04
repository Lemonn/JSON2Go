package errors

// TypeConflictError Is emitted whenever a type with no custom type assigned changes it's representation.
type TypeConflictError struct {
	OldType string
	NewType string
}

func (e *TypeConflictError) Error() string {
	return "type conflict old:new: " + e.OldType + ":" + e.NewType
}
