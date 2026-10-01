package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

// AppError is an application-level error with HTTP status and stable code.
type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Err        error
	Details    map[string]string
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetail returns a copy with an extra detail entry.
func (e *AppError) WithDetail(key, value string) *AppError {
	cp := *e
	if cp.Details == nil {
		cp.Details = map[string]string{}
	} else {
		cp.Details = mapsClone(e.Details)
	}
	cp.Details[key] = value
	return &cp
}

func mapsClone(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// New creates an AppError.
func New(status int, code, message string) *AppError {
	return &AppError{StatusCode: status, Code: code, Message: message}
}

// Wrap adds an underlying error.
func Wrap(status int, code, message string, err error) *AppError {
	return &AppError{StatusCode: status, Code: code, Message: message, Err: err}
}

// Common constructors for handlers and adapters.
// Codes are snake_case stable identifiers; clients switch on Code, not Message.
var (
	ErrNotFound                     = New(http.StatusNotFound, "not_found", "Resource not found")
	ErrProfileNotFound              = New(http.StatusNotFound, "profile_not_found", "Profile not found")
	ErrConflict                     = New(http.StatusConflict, "conflict", "Resource conflict")
	ErrValidation                   = New(http.StatusBadRequest, "validation_error", "Validation failed")
	ErrUnauthorized                 = New(http.StatusUnauthorized, "unauthorized", "Authentication required")
	ErrTokenExpired                 = New(http.StatusUnauthorized, "token_expired", "Access token expired")
	ErrInvalidCredentials           = New(http.StatusUnauthorized, "invalid_credentials", "Invalid credentials")
	ErrInvalidRefreshToken          = New(http.StatusUnauthorized, "invalid_refresh_token", "Invalid refresh token")
	ErrMissingRefreshToken          = New(http.StatusUnauthorized, "missing_refresh_token", "Refresh token required")
	ErrSessionRevoked               = New(http.StatusUnauthorized, "session_revoked", "Session revoked")
	ErrForbidden                    = New(http.StatusForbidden, "forbidden", "Insufficient permissions")
	ErrSelfDeleteCustomerOnly       = New(http.StatusForbidden, "forbidden", "Only customers can self-delete")
	ErrPasswordChangeRequired       = New(http.StatusForbidden, "password_change_required", "Password change required before accessing this resource")
	ErrInternal                     = New(http.StatusInternalServerError, "internal_error", "Internal server error")
	ErrServiceUnavailable           = New(http.StatusServiceUnavailable, "service_unavailable", "Service unavailable")
	ErrTooManyRequests              = New(http.StatusTooManyRequests, "rate_limited", "Too many requests")
	ErrCannotModifySelf             = New(http.StatusForbidden, "cannot_modify_self", "Cannot modify your own account")
	ErrCannotModifyAdmin            = New(http.StatusForbidden, "cannot_modify_admin", "Cannot modify an admin account")
	ErrInvalidRoleTransition        = New(http.StatusUnprocessableEntity, "invalid_role_transition", "Invalid role transition")
	ErrEmployeeCodeExists           = New(http.StatusConflict, "employee_code_exists", "Employee code already exists")
	ErrPhoneExists                  = New(http.StatusConflict, "phone_exists", "Phone number is already in use")
	ErrCategoryHasProducts          = New(http.StatusUnprocessableEntity, "category_has_products", "Category still has products")
	ErrSlugExists                   = New(http.StatusConflict, "slug_exists", "Slug already exists")
	ErrCodeExists                   = New(http.StatusConflict, "code_exists", "Code already exists")
	ErrDiscountInactive             = New(http.StatusUnprocessableEntity, "discount_inactive", "Discount code is inactive")
	ErrDiscountExpired              = New(http.StatusUnprocessableEntity, "discount_expired", "Discount code has expired")
	ErrDiscountExhausted            = New(http.StatusUnprocessableEntity, "discount_exhausted", "Discount code has no uses remaining")
	ErrDiscountMinOrderNotMet       = New(http.StatusUnprocessableEntity, "discount_min_order_not_met", "Order subtotal does not meet the discount minimum")
	ErrDiscountUsedUp               = New(http.StatusUnprocessableEntity, "discount_used_up", "You have already used this discount code as often as it allows")
	ErrCartEmpty                    = New(http.StatusUnprocessableEntity, "cart_empty", "Cart is empty")
	ErrCartMaxItems                 = New(http.StatusUnprocessableEntity, "cart_max_items", "Cart cannot hold more items")
	ErrProductUnavailable           = New(http.StatusUnprocessableEntity, "product_unavailable", "Product is not available")
	ErrComboUnavailable             = New(http.StatusUnprocessableEntity, "combo_unavailable", "Combo is not available")
	ErrInvalidOrderStatusTransition = New(http.StatusUnprocessableEntity, "invalid_order_status_transition", "Invalid order status transition")
	ErrPickupTooSoon                = New(http.StatusUnprocessableEntity, "pickup_too_soon", "Pickup time is sooner than the bakery can prepare the order")
	ErrPickupTooFar                 = New(http.StatusUnprocessableEntity, "pickup_too_far", "Pickup time is further ahead than the bakery takes orders")
	ErrPickupClosedDay              = New(http.StatusUnprocessableEntity, "pickup_closed_day", "The bakery is closed on that day")
	ErrPickupOutsideHours           = New(http.StatusUnprocessableEntity, "pickup_outside_hours", "Pickup time is outside opening hours")
	ErrPickupOffSlot                = New(http.StatusUnprocessableEntity, "pickup_off_slot", "Pickup time is not one of the pickup slots")
	ErrPickupSlotFull               = New(http.StatusConflict, "pickup_slot_full", "That pickup slot is full")
	ErrPickupSoldOut                = New(http.StatusUnprocessableEntity, "pickup_sold_out", "An item in the order is sold out that day; choose another day")
	ErrInvalidTicketTransition      = New(http.StatusUnprocessableEntity, "invalid_ticket_transition", "That ticket cannot move to that status")
	ErrTicketOrderNotActive         = New(http.StatusUnprocessableEntity, "ticket_order_not_active", "The order is not being made")
	ErrTicketNotMovable             = New(http.StatusUnprocessableEntity, "ticket_not_movable", "Only a ticket nobody has started can move")
	ErrTicketStationTaken           = New(http.StatusConflict, "ticket_station_taken", "The order already has a ticket at that station")
	ErrPickupDayLimit               = New(http.StatusUnprocessableEntity, "pickup_day_limit", "You already have as many orders as one customer can book for that day")
	ErrClosedDateExists             = New(http.StatusConflict, "closed_date_exists", "That day is already closed")
	ErrTermsNotAccepted             = New(http.StatusUnprocessableEntity, "terms_not_accepted", "Please accept the current terms, privacy policy and refund policy")
	ErrAccountHasOpenOrders         = New(http.StatusUnprocessableEntity, "account_has_open_orders", "You have an order that is not collected or cancelled yet")
	ErrAccountErased                = New(http.StatusUnprocessableEntity, "account_erased", "This account was erased at its owner's request and cannot be restored")
	ErrEmailNotVerified             = New(http.StatusUnprocessableEntity, "email_not_verified", "Confirm your email address before placing an order")
	ErrEmailAlreadyVerified         = New(http.StatusConflict, "email_already_verified", "Your email address is already confirmed")
	ErrInvalidLink                  = New(http.StatusUnprocessableEntity, "invalid_link", "This link has expired or has already been used")
	ErrOrderNotPayable              = New(http.StatusUnprocessableEntity, "order_not_payable", "This order can no longer be paid")
	ErrPaymentNotCompleted          = New(http.StatusUnprocessableEntity, "payment_not_completed", "The payment was not completed")
	ErrPaymentUnderReview           = New(http.StatusConflict, "payment_under_review", "PayPal is still reviewing your payment; try again once it clears")
	ErrPickupCodeInvalid            = New(http.StatusUnprocessableEntity, "pickup_code_invalid", "That pickup code does not match this order")
	ErrPickupCodeLocked             = New(http.StatusTooManyRequests, "pickup_code_locked", "Too many wrong pickup codes for this order; try again later")
	ErrMessagingUnavailable         = New(http.StatusUnprocessableEntity, "messaging_unavailable", "Messages are not available for this order")
	ErrWebhookInvalid               = New(http.StatusBadRequest, "webhook_invalid", "The notice is not signed by the payment provider")
)

// ToErrorBody projects an AppError into the HTTP response error body shape.
// Returns nil-safe Code/Message; use this from handlers to keep the wire contract identical to sentinel definitions.
func (e *AppError) ToErrorBody() (code, message string, details map[string]string) {
	if e == nil {
		return ErrInternal.Code, ErrInternal.Message, nil
	}
	return e.Code, e.Message, e.Details
}

// FromFiberError maps *fiber.Error to AppError.
func FromFiberError(err *fiber.Error) *AppError {
	if err == nil {
		return nil
	}
	return New(err.Code, httpStatusText(err.Code), err.Message)
}

func httpStatusText(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusTooManyRequests:
		return "rate_limited"
	default:
		return "http_error"
	}
}

// AsAppError unwraps err into *AppError when possible.
func AsAppError(err error) (*AppError, bool) {
	var ae *AppError
	if stderrors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

// Errorf wraps fmt.Errorf as an internal AppError.
func Errorf(format string, args ...any) *AppError {
	return Wrap(http.StatusInternalServerError, "internal_error", "Internal server error", fmt.Errorf(format, args...))
}
