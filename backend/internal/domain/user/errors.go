package user

import "errors"

var (
	ErrProfileNotFound        = errors.New("profile not found")
	ErrInvalidRoleTransition  = errors.New("invalid role transition")
	ErrEmployeeCodeExists     = errors.New("employee code already exists")
	ErrCannotModifySelf       = errors.New("cannot modify self")
	ErrSelfDeleteCustomerOnly = errors.New("self delete customer only")
)
