package errors

import "fmt"

// ActiveAdjusterNotFoundError Is emitted whenever a type was previously replaced by a TypeCheckers that is no longer
// present in the current list of TypeCheckers
type ActiveAdjusterNotFoundError struct {
	NotFoundCheckerName string
	ActiveCheckerNames  []string
}

func (e *ActiveAdjusterNotFoundError) Error() string {
	return fmt.Sprintf("type was previsoly replaced by %s and needs a recheck. But the used Typechekcer "+
		"is no longer active. Currently active TypeCheckers: %v", e.NotFoundCheckerName, e.ActiveCheckerNames)
}
