package errors

type incompatibleCheckerStatesError struct{}

func (i *incompatibleCheckerStatesError) Error() string {
	return "Non-Combinable Checker States Error"
}
