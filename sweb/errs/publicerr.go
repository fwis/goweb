package errs

import (
	"fmt"
)

type PubError struct {
	error
	Message string
}

func (e *PubError) String() string { return e.Message }

func (e *PubError) Error() string {
	return e.Message
}

func NewPubError(msg string) *PubError { return &PubError{Message: msg} }

func PubErrorf(format string, a ...interface{}) *PubError {
	return NewPubError(fmt.Sprintf(format, a...))
}
