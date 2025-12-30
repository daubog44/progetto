package resiliency

import "errors"

// PermanentError represents an error that should not be retried.
type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string {
	return e.Err.Error()
}

func (e *PermanentError) Unwrap() error {
	return e.Err
}

// NewPermanentError wraps an error as a PermanentError.
func NewPermanentError(err error) error {
	return &PermanentError{Err: err}
}

// IsPermanentError checks if the error is of type PermanentError.
func IsPermanentError(err error) bool {
	var target *PermanentError
	return errors.As(err, &target)
}
