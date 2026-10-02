package promotion

import "errors"

var (
	ErrInvalidSubject          = errors.New("promotion: a subject is 1 to 120 plain characters on one line")
	ErrInvalidBody             = errors.New("promotion: a message is 1 to 5000 plain characters")
	ErrInvalidUnsubscribeToken = errors.New("promotion: the unsubscribe link is not one we sent")
)
