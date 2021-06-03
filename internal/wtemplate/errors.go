package wtemplate

// OperationAbortedError is the error for the template operation chain failure.
type OperationAbortedError struct {
	message string
}

// Error returns the error description.
func (e OperationAbortedError) Error() string {
	return e.message
}

