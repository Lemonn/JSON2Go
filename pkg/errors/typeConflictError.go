package errors

type TypeConflictError struct {
	OldType string
	NewType string
}

func (e *TypeConflictError) Error() string {
	return "type conflict old:new: " + e.OldType + ":" + e.NewType
}
