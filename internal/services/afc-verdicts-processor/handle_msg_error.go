package afcverdictsprocessor

import "errors"

type handleMessageError struct {
	temporary bool
	err       error
}

func newHandleMessageError(err error, isTemporary bool) *handleMessageError {
	return &handleMessageError{
		temporary: isTemporary,
		err:       err,
	}
}

func asHandleError(err error) (*handleMessageError, bool) {
	e := &handleMessageError{}

	if errors.As(err, &e) {
		return e, true
	}

	return nil, false
}

func (e *handleMessageError) Error() string {
	return e.err.Error()
}

func (e *handleMessageError) IsTemporary() bool {
	return e.temporary
}

func (e *handleMessageError) Is(target error) bool {
	return errors.Is(e.err, target)
}
