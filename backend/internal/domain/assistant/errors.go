package assistant

import "errors"

var (
	ErrInvalidQuestion = errors.New("assistant: a question is 1 to 500 plain characters")
	// ErrUnavailable: the model could not be reached, or did not answer in time.
	ErrUnavailable = errors.New("assistant: the assistant cannot answer right now")
)
