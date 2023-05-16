package outbox

import (
	"errors"
	"fmt"
)

var ErrJobAlreadyExists = errors.New("job with provided name already registered")

type jobFailedError struct {
	reason string
}

func newJobFailedError(reason string) error {
	return &jobFailedError{reason: reason}
}

func (e *jobFailedError) Error() string {
	return fmt.Sprintf("job failed with reason=%s", e.reason)
}

func (e *jobFailedError) getReason() string {
	return e.reason
}
