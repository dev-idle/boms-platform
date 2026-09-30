package policy

import "errors"

// ErrTermsNotAccepted refuses a registration or an order made without
// accepting the current policies.
var ErrTermsNotAccepted = errors.New("the current terms were not accepted")
