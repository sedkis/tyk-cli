package cli

// ExitError represents an error with a specific exit code
type ExitError struct {
	Code    int
	Message string
}

func (e *ExitError) Error() string {
	return e.Message
}