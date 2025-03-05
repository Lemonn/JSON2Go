package typeChecker

type ActiveCheckerExcludedError struct {
}

func (e ActiveCheckerExcludedError) Error() string {
	return "ActiveCheckerExcludedError"
}
