package store

import "errors"

var (
	ErrInvalidHours            = errors.New("opening time must be before closing time, within one day, in whole minutes")
	ErrInvalidLeadTime         = errors.New("pre-order lead time must be whole minutes, at most seven days, and shorter than the booking window")
	ErrInvalidAdvanceDays      = errors.New("booking window must be between 1 and 90 days")
	ErrClosedDateOutOfRange    = errors.New("closed date must be between today and a year ahead")
	ErrInvalidClosedDateReason = errors.New("closed date reason must be 1 to 200 characters")
	ErrClosedDateExists        = errors.New("that day is already closed")
	ErrClosedDateNotFound      = errors.New("closed date not found")
)
