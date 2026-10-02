package v1

import (
	"errors"

	domainaccount "github.com/boms/backend/internal/domain/account"
	domaincart "github.com/boms/backend/internal/domain/cart"
	domaincategory "github.com/boms/backend/internal/domain/category"
	domaincombo "github.com/boms/backend/internal/domain/combo"
	domainconversation "github.com/boms/backend/internal/domain/conversation"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainpayment "github.com/boms/backend/internal/domain/payment"
	domainpolicy "github.com/boms/backend/internal/domain/policy"
	domainproduct "github.com/boms/backend/internal/domain/product"
	domainpromotion "github.com/boms/backend/internal/domain/promotion"
	domainreview "github.com/boms/backend/internal/domain/review"
	domainsaved "github.com/boms/backend/internal/domain/saved"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v3"
)

// writeMapUsecaseError maps known domain and AppError values to HTTP responses.
// Unmapped errors propagate to fiber.ErrorHandler as internal_error with logging.
func writeMapUsecaseError(c fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, usecase.ErrEmailExists):
		return writeAppError(c, usecase.ErrEmailExists)
	case errors.Is(err, usecase.ErrMeNotFound):
		return writeAppError(c, usecase.ErrMeNotFound)
	case errors.Is(err, usecase.ErrUserNotFound):
		return writeAppError(c, usecase.ErrUserNotFound)
	case errors.Is(err, apperrors.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, apperrors.ErrConflict):
		return writeAppError(c, apperrors.ErrConflict)
	case errors.Is(err, apperrors.ErrInvalidCredentials):
		return writeAppError(c, apperrors.ErrInvalidCredentials)
	case errors.Is(err, apperrors.ErrInvalidRefreshToken):
		return writeAppError(c, apperrors.ErrInvalidRefreshToken)
	case errors.Is(err, apperrors.ErrMissingRefreshToken):
		return writeAppError(c, apperrors.ErrMissingRefreshToken)
	case errors.Is(err, apperrors.ErrSessionRevoked):
		return writeAppError(c, apperrors.ErrSessionRevoked)
	case errors.Is(err, apperrors.ErrTokenExpired):
		return writeAppError(c, apperrors.ErrTokenExpired)
	case errors.Is(err, apperrors.ErrForbidden):
		return writeAppError(c, apperrors.ErrForbidden)
	case errors.Is(err, domainuser.ErrSelfDeleteCustomerOnly):
		return writeAppError(c, apperrors.ErrSelfDeleteCustomerOnly)
	case errors.Is(err, domainuser.ErrAccountHasOpenOrders):
		return writeAppError(c, apperrors.ErrAccountHasOpenOrders)
	case errors.Is(err, domainuser.ErrAccountErased):
		return writeAppError(c, apperrors.ErrAccountErased)
	case errors.Is(err, domainuser.ErrEmailNotVerified):
		return writeAppError(c, apperrors.ErrEmailNotVerified)
	case errors.Is(err, domainuser.ErrEmailAlreadyVerified):
		return writeAppError(c, apperrors.ErrEmailAlreadyVerified)
	case errors.Is(err, domainaccount.ErrLinkInvalid), errors.Is(err, domainpromotion.ErrInvalidUnsubscribeToken):
		return writeAppError(c, apperrors.ErrInvalidLink)
	case errors.Is(err, domainuser.ErrProfileNotFound):
		return writeAppError(c, apperrors.ErrProfileNotFound)
	case errors.Is(err, domainuser.ErrEmployeeCodeExists):
		return writeAppError(c, apperrors.ErrEmployeeCodeExists)
	case errors.Is(err, domainuser.ErrPhoneExists):
		return writeAppError(c, apperrors.ErrPhoneExists)
	case errors.Is(err, domainuser.ErrCannotModifySelf):
		return writeAppError(c, apperrors.ErrCannotModifySelf)
	case errors.Is(err, domainuser.ErrCannotModifyAdmin):
		return writeAppError(c, apperrors.ErrCannotModifyAdmin)
	case errors.Is(err, domainuser.ErrInvalidRoleTransition):
		return writeAppError(c, apperrors.ErrInvalidRoleTransition)
	case errors.Is(err, domaincategory.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domaincategory.ErrHasProducts):
		return writeAppError(c, apperrors.ErrCategoryHasProducts)
	case errors.Is(err, domaincategory.ErrSlugExists):
		return writeAppError(c, apperrors.ErrSlugExists)
	case errors.Is(err, domaincategory.ErrInactive):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("category_id", "category is inactive"))
	case errors.Is(err, domainproduct.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domainproduct.ErrSlugExists):
		return writeAppError(c, apperrors.ErrSlugExists)
	case errors.Is(err, domainproduct.ErrInvalidOption):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("options", "group, a label of 1-60 characters of plain text, and an added price of $0-$1,000"))
	case errors.Is(err, domainproduct.ErrInvalidChoice):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("option_ids", "choose one option from each group"))
	case errors.Is(err, domainproduct.ErrInvalidMessage):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("message", "at most 60 characters of plain text"))
	case errors.Is(err, domainproduct.ErrCustomization):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("customization", "configure a customizable product, and only that"))
	case errors.Is(err, domaincombo.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domaincombo.ErrSlugExists):
		return writeAppError(c, apperrors.ErrSlugExists)
	case errors.Is(err, domaindiscount.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domaindiscount.ErrCodeExists):
		return writeAppError(c, apperrors.ErrCodeExists)
	case errors.Is(err, domaindiscount.ErrInactive):
		return writeAppError(c, apperrors.ErrDiscountInactive)
	case errors.Is(err, domaindiscount.ErrExpired):
		return writeAppError(c, apperrors.ErrDiscountExpired)
	case errors.Is(err, domaindiscount.ErrExhausted):
		return writeAppError(c, apperrors.ErrDiscountExhausted)
	case errors.Is(err, domaindiscount.ErrMinOrderNotMet):
		return writeAppError(c, apperrors.ErrDiscountMinOrderNotMet)
	case errors.Is(err, domaindiscount.ErrUsedUpByCustomer):
		return writeAppError(c, apperrors.ErrDiscountUsedUp)
	case errors.Is(err, domainpayment.ErrNotPayable):
		return writeAppError(c, apperrors.ErrOrderNotPayable)
	case errors.Is(err, domainpayment.ErrNotCompleted):
		return writeAppError(c, apperrors.ErrPaymentNotCompleted)
	case errors.Is(err, domainpayment.ErrUnderReview):
		return writeAppError(c, apperrors.ErrPaymentUnderReview)
	case errors.Is(err, domainpayment.ErrWebhookInvalid):
		return writeAppError(c, apperrors.ErrWebhookInvalid)
	case errors.Is(err, domaincart.ErrItemNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domaincart.ErrEmpty):
		return writeAppError(c, apperrors.ErrCartEmpty)
	case errors.Is(err, domaincart.ErrProductUnavailable):
		return writeAppError(c, apperrors.ErrProductUnavailable)
	case errors.Is(err, domaincart.ErrComboUnavailable):
		return writeAppError(c, apperrors.ErrComboUnavailable)
	case errors.Is(err, domaincart.ErrMaxItemsReached):
		return writeAppError(c, apperrors.ErrCartMaxItems)
	case errors.Is(err, domaincart.ErrQuantityOutOfRange):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("quantity", "must be between 1 and 99"))
	case errors.Is(err, domaincart.ErrInvalidLine):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("line", "provide exactly one of product_id or combo_id"))
	case errors.Is(err, domainorder.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domainorder.ErrInvalidStatusTransition):
		return writeAppError(c, apperrors.ErrInvalidOrderStatusTransition)
	case errors.Is(err, domainorder.ErrInvalidPickupAt):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("pickup_at", "pickup time is required"))
	case errors.Is(err, domainorder.ErrPickupTooSoon):
		return writeAppError(c, apperrors.ErrPickupTooSoon)
	case errors.Is(err, domainorder.ErrPickupTooFar):
		return writeAppError(c, apperrors.ErrPickupTooFar)
	case errors.Is(err, domainorder.ErrPickupClosedDay):
		return writeAppError(c, apperrors.ErrPickupClosedDay)
	case errors.Is(err, domainorder.ErrPickupOutsideHours):
		return writeAppError(c, apperrors.ErrPickupOutsideHours)
	case errors.Is(err, domainorder.ErrPickupOffSlot):
		return writeAppError(c, apperrors.ErrPickupOffSlot)
	case errors.Is(err, domainorder.ErrPickupSlotFull):
		return writeAppError(c, apperrors.ErrPickupSlotFull)
	case errors.Is(err, domainorder.ErrPickupSoldOut):
		return writeAppError(c, apperrors.ErrPickupSoldOut)
	case errors.Is(err, domainorder.ErrInvalidGuestName):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("guest.name", "a name of 1 to 100 plain characters"))
	case errors.Is(err, domainorder.ErrTicketNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domainorder.ErrInvalidTicketTransition):
		return writeAppError(c, apperrors.ErrInvalidTicketTransition)
	case errors.Is(err, domainorder.ErrTicketOrderNotActive):
		return writeAppError(c, apperrors.ErrTicketOrderNotActive)
	case errors.Is(err, domainorder.ErrTicketNotMovable):
		return writeAppError(c, apperrors.ErrTicketNotMovable)
	case errors.Is(err, domainorder.ErrTicketStationTaken):
		return writeAppError(c, apperrors.ErrTicketStationTaken)
	case errors.Is(err, domainorder.ErrPickupDayLimit):
		return writeAppError(c, apperrors.ErrPickupDayLimit)
	case errors.Is(err, domainpolicy.ErrTermsNotAccepted):
		return writeAppError(c, apperrors.ErrTermsNotAccepted)
	case errors.Is(err, domainstore.ErrInvalidHours):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("closes_at", "must be after the opening time, within the day"))
	case errors.Is(err, domainstore.ErrInvalidLeadTime):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("preorder_min_lead_minutes", "must be 0 to 10080 minutes and shorter than the booking window"))
	case errors.Is(err, domainstore.ErrInvalidAdvanceDays):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("max_advance_days", "must be between 1 and 90"))
	case errors.Is(err, domainstore.ErrInvalidSlotLength):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("slot_minutes", "must be 10, 15, 20, 30 or 60 and fit in the opening hours"))
	case errors.Is(err, domainstore.ErrInvalidSlotCapacity):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("slot_capacity", "must be between 1 and 200"))
	case errors.Is(err, domainstore.ErrInvalidInstantPrep):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("instant_prep_minutes", "must be 0 to 240 minutes"))
	case errors.Is(err, domainstore.ErrInvalidPaymentHold):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("payment_hold_minutes", "must be 5 to 120 minutes"))
	case errors.Is(err, domainstore.ErrClosedDateOutOfRange):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("date", "must be between today and a year ahead"))
	case errors.Is(err, domainorder.ErrPickupCodeInvalid):
		return writeAppError(c, apperrors.ErrPickupCodeInvalid)
	case errors.Is(err, domainorder.ErrPickupCodeLocked):
		return writeAppError(c, apperrors.ErrPickupCodeLocked)
	case errors.Is(err, domainconversation.ErrInvalidBody):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "a message of 1 to 2000 plain characters"))
	case errors.Is(err, domainconversation.ErrUnavailable):
		return writeAppError(c, apperrors.ErrMessagingUnavailable)
	case errors.Is(err, domainconversation.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domainsaved.ErrInvalidList):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("list", "must be favorite or wishlist"))
	case errors.Is(err, domainsaved.ErrListFull):
		return writeAppError(c, apperrors.ErrSavedListFull)
	case errors.Is(err, domainreview.ErrInvalidRating):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("rating", "must be 1 to 5 stars"))
	case errors.Is(err, domainreview.ErrInvalidComment):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("comment", "up to 1000 plain characters"))
	case errors.Is(err, domainreview.ErrInvalidStatus):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("status", "must be published or hidden"))
	case errors.Is(err, domainreview.ErrNotReviewable):
		return writeAppError(c, apperrors.ErrReviewUnavailable)
	case errors.Is(err, domainreview.ErrExists):
		return writeAppError(c, apperrors.ErrReviewExists)
	case errors.Is(err, domainreview.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domainpromotion.ErrInvalidSubject):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("subject", "1 to 120 plain characters on one line"))
	case errors.Is(err, domainpromotion.ErrInvalidBody):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "1 to 5000 plain characters"))
	case errors.Is(err, domainorder.ErrInvalidCancelReason):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("reason", "must be 1 to 200 plain characters"))
	case errors.Is(err, domainstore.ErrInvalidClosedDateReason):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("reason", "must be 1 to 200 characters"))
	case errors.Is(err, domainstore.ErrClosedDateExists):
		return writeAppError(c, apperrors.ErrClosedDateExists)
	case errors.Is(err, domainstore.ErrClosedDateNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	}
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		return writeAppError(c, appErr)
	}
	return err
}

// mustChangePasswordPtr returns a JSON pointer when the flag is true (omitempty otherwise).
func mustChangePasswordPtr(v bool) *bool {
	if !v {
		return nil
	}
	t := true
	return &t
}
