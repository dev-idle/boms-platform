package user

import "errors"

var (
	ErrProfileNotFound        = errors.New("profile not found")
	ErrInvalidRoleTransition  = errors.New("invalid role transition")
	ErrEmployeeCodeExists     = errors.New("employee code already exists")
	ErrPhoneExists            = errors.New("phone already exists")
	ErrCannotModifySelf       = errors.New("cannot modify self")
	ErrCannotModifyAdmin      = errors.New("cannot modify admin")
	ErrSelfDeleteCustomerOnly = errors.New("self delete customer only")
	// ErrAccountHasOpenOrders refuses to erase an account while the bakery still
	// has one of its orders to make or hand over.
	ErrAccountHasOpenOrders = errors.New("account has open orders")
	// ErrAccountErased refuses to restore an account erased at its owner's request.
	ErrAccountErased = errors.New("account erased")
	// ErrEmailNotVerified refuses an order from an account whose address is not
	// confirmed: order updates go to that address, and it must be the customer's.
	ErrEmailNotVerified = errors.New("email not verified")
	// ErrEmailAlreadyVerified answers a request for a new confirmation link
	// from an account whose address is confirmed, often from another device.
	ErrEmailAlreadyVerified = errors.New("email already verified")
)
