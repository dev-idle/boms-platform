package discount

import "time"

// ValidateRedeemable checks whether a code can be applied at the given time and subtotal.
func ValidateRedeemable(code *Code, now time.Time, subtotalCents int64) error {
	if code == nil {
		return ErrNotFound
	}
	if code.DeletedAt != nil {
		return ErrNotFound
	}
	if !code.IsActive {
		return ErrInactive
	}
	if now.Before(code.StartsAt) || !now.Before(code.EndsAt) {
		return ErrExpired
	}
	if code.MaxUses != nil && code.UsedCount >= *code.MaxUses {
		return ErrExhausted
	}
	if code.MinOrderCents != nil && subtotalCents < *code.MinOrderCents {
		return ErrMinOrderNotMet
	}
	return nil
}

// ComputeDiscountCents returns the discount amount for a redeemable code and subtotal.
// Caller must run ValidateRedeemable first.
func ComputeDiscountCents(code *Code, subtotalCents int64) int64 {
	switch code.DiscountType {
	case TypePercent:
		return subtotalCents * code.Value / 100
	case TypeFixedCents:
		if code.Value >= subtotalCents {
			return subtotalCents
		}
		return code.Value
	default:
		return 0
	}
}
