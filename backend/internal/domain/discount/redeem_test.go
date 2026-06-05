package discount_test

import (
	"testing"
	"time"

	domaindiscount "github.com/boms/backend/internal/domain/discount"
)

func TestComputeDiscountCents_percent(t *testing.T) {
	t.Parallel()
	code := &domaindiscount.Code{
		DiscountType: domaindiscount.TypePercent,
		Value:        10,
	}
	if got := domaindiscount.ComputeDiscountCents(code, 2500); got != 250 {
		t.Fatalf("ComputeDiscountCents() = %d, want 250", got)
	}
}

func TestComputeDiscountCents_fixedCapsAtSubtotal(t *testing.T) {
	t.Parallel()
	code := &domaindiscount.Code{
		DiscountType: domaindiscount.TypeFixedCents,
		Value:        500,
	}
	if got := domaindiscount.ComputeDiscountCents(code, 300); got != 300 {
		t.Fatalf("ComputeDiscountCents() = %d, want 300", got)
	}
}

func TestValidateRedeemable_minOrder(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)
	min := int64(1000)
	code := &domaindiscount.Code{
		IsActive:      true,
		StartsAt:      now.Add(-time.Hour),
		EndsAt:        now.Add(time.Hour),
		MinOrderCents: &min,
	}
	if err := domaindiscount.ValidateRedeemable(code, now, 500); err != domaindiscount.ErrMinOrderNotMet {
		t.Fatalf("ValidateRedeemable() = %v, want ErrMinOrderNotMet", err)
	}
}
