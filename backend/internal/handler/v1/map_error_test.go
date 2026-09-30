package v1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	domainaccount "github.com/boms/backend/internal/domain/account"
	domaincart "github.com/boms/backend/internal/domain/cart"
	domaincategory "github.com/boms/backend/internal/domain/category"
	domaincombo "github.com/boms/backend/internal/domain/combo"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainpolicy "github.com/boms/backend/internal/domain/policy"
	domainproduct "github.com/boms/backend/internal/domain/product"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWriteMapUsecaseError_mapsKnownErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "not_found", err: apperrors.ErrNotFound, wantStatus: 404, wantCode: "not_found"},
		{name: "invalid_credentials", err: apperrors.ErrInvalidCredentials, wantStatus: 401, wantCode: "invalid_credentials"},
		{name: "session_revoked", err: apperrors.ErrSessionRevoked, wantStatus: 401, wantCode: "session_revoked"},
		{name: "forbidden", err: apperrors.ErrForbidden, wantStatus: 403, wantCode: "forbidden"},
		{name: "cannot_modify_self", err: domainuser.ErrCannotModifySelf, wantStatus: 403, wantCode: "cannot_modify_self"},
		{name: "cannot_modify_admin", err: domainuser.ErrCannotModifyAdmin, wantStatus: 403, wantCode: "cannot_modify_admin"},
		{name: "self_delete_customer_only", err: domainuser.ErrSelfDeleteCustomerOnly, wantStatus: 403, wantCode: "forbidden"},
		{name: "employee_code_exists", err: domainuser.ErrEmployeeCodeExists, wantStatus: 409, wantCode: "employee_code_exists"},
		{name: "phone_exists", err: domainuser.ErrPhoneExists, wantStatus: 409, wantCode: "phone_exists"},
		{name: "email_exists", err: usecase.ErrEmailExists, wantStatus: 409, wantCode: "email_exists"},
		{name: "category_has_products", err: domaincategory.ErrHasProducts, wantStatus: 422, wantCode: "category_has_products"},
		{name: "category_slug_exists", err: domaincategory.ErrSlugExists, wantStatus: 409, wantCode: "slug_exists"},
		{name: "product_not_found", err: domainproduct.ErrNotFound, wantStatus: 404, wantCode: "not_found"},
		{name: "combo_not_found", err: domaincombo.ErrNotFound, wantStatus: 404, wantCode: "not_found"},
		{name: "combo_slug_exists", err: domaincombo.ErrSlugExists, wantStatus: 409, wantCode: "slug_exists"},
		{name: "discount_code_not_found", err: domaindiscount.ErrNotFound, wantStatus: 404, wantCode: "not_found"},
		{name: "discount_code_exists", err: domaindiscount.ErrCodeExists, wantStatus: 409, wantCode: "code_exists"},
		{name: "discount_inactive", err: domaindiscount.ErrInactive, wantStatus: 422, wantCode: "discount_inactive"},
		{name: "discount_expired", err: domaindiscount.ErrExpired, wantStatus: 422, wantCode: "discount_expired"},
		{name: "cart_empty", err: domaincart.ErrEmpty, wantStatus: 422, wantCode: "cart_empty"},
		{name: "product_unavailable", err: domaincart.ErrProductUnavailable, wantStatus: 422, wantCode: "product_unavailable"},
		{name: "invalid_order_status_transition", err: domainorder.ErrInvalidStatusTransition, wantStatus: 422, wantCode: "invalid_order_status_transition"},
		{name: "invalid_pickup_at", err: domainorder.ErrInvalidPickupAt, wantStatus: 400, wantCode: "validation_error"},
		{name: "pickup_too_soon", err: domainorder.ErrPickupTooSoon, wantStatus: 422, wantCode: "pickup_too_soon"},
		{name: "pickup_too_far", err: domainorder.ErrPickupTooFar, wantStatus: 422, wantCode: "pickup_too_far"},
		{name: "pickup_closed_day", err: domainorder.ErrPickupClosedDay, wantStatus: 422, wantCode: "pickup_closed_day"},
		{name: "pickup_outside_hours", err: domainorder.ErrPickupOutsideHours, wantStatus: 422, wantCode: "pickup_outside_hours"},
		{name: "pickup_off_slot", err: domainorder.ErrPickupOffSlot, wantStatus: 422, wantCode: "pickup_off_slot"},
		{name: "pickup_slot_full", err: domainorder.ErrPickupSlotFull, wantStatus: 409, wantCode: "pickup_slot_full"},
		{name: "ticket_not_found", err: domainorder.ErrTicketNotFound, wantStatus: 404, wantCode: "not_found"},
		{name: "invalid_ticket_transition", err: domainorder.ErrInvalidTicketTransition, wantStatus: 422, wantCode: "invalid_ticket_transition"},
		{name: "ticket_order_not_active", err: domainorder.ErrTicketOrderNotActive, wantStatus: 422, wantCode: "ticket_order_not_active"},
		{name: "ticket_not_movable", err: domainorder.ErrTicketNotMovable, wantStatus: 422, wantCode: "ticket_not_movable"},
		{name: "ticket_station_taken", err: domainorder.ErrTicketStationTaken, wantStatus: 409, wantCode: "ticket_station_taken"},
		{name: "pickup_day_limit", err: domainorder.ErrPickupDayLimit, wantStatus: 422, wantCode: "pickup_day_limit"},
		{name: "terms_not_accepted", err: domainpolicy.ErrTermsNotAccepted, wantStatus: 422, wantCode: "terms_not_accepted"},
		{name: "account_has_open_orders", err: domainuser.ErrAccountHasOpenOrders, wantStatus: 422, wantCode: "account_has_open_orders"},
		{name: "account_erased", err: domainuser.ErrAccountErased, wantStatus: 422, wantCode: "account_erased"},
		{name: "email_not_verified", err: domainuser.ErrEmailNotVerified, wantStatus: 422, wantCode: "email_not_verified"},
		{name: "email_already_verified", err: domainuser.ErrEmailAlreadyVerified, wantStatus: 409, wantCode: "email_already_verified"},
		{name: "invalid_link", err: domainaccount.ErrLinkInvalid, wantStatus: 422, wantCode: "invalid_link"},
		{name: "store_invalid_hours", err: domainstore.ErrInvalidHours, wantStatus: 400, wantCode: "validation_error"},
		{name: "store_invalid_lead_time", err: domainstore.ErrInvalidLeadTime, wantStatus: 400, wantCode: "validation_error"},
		{name: "store_invalid_advance_days", err: domainstore.ErrInvalidAdvanceDays, wantStatus: 400, wantCode: "validation_error"},
		{name: "store_invalid_slot_length", err: domainstore.ErrInvalidSlotLength, wantStatus: 400, wantCode: "validation_error"},
		{name: "store_invalid_slot_capacity", err: domainstore.ErrInvalidSlotCapacity, wantStatus: 400, wantCode: "validation_error"},
		{name: "store_invalid_instant_prep", err: domainstore.ErrInvalidInstantPrep, wantStatus: 400, wantCode: "validation_error"},
		{name: "closed_date_out_of_range", err: domainstore.ErrClosedDateOutOfRange, wantStatus: 400, wantCode: "validation_error"},
		{name: "closed_date_reason", err: domainstore.ErrInvalidClosedDateReason, wantStatus: 400, wantCode: "validation_error"},
		{name: "closed_date_exists", err: domainstore.ErrClosedDateExists, wantStatus: 409, wantCode: "closed_date_exists"},
		{name: "closed_date_not_found", err: domainstore.ErrClosedDateNotFound, wantStatus: 404, wantCode: "not_found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			status, code := invokeMapError(t, tt.err)
			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, tt.wantCode, code)
		})
	}
}

func TestWriteMapUsecaseError_unmappedReturnsInternal(t *testing.T) {
	t.Parallel()
	status, code := invokeMapErrorViaErrorHandler(t, assert.AnError)
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Equal(t, apperrors.ErrInternal.Code, code)
}

func TestMustChangePasswordPtr(t *testing.T) {
	t.Parallel()
	assert.Nil(t, mustChangePasswordPtr(false))
	require.NotNil(t, mustChangePasswordPtr(true))
	assert.True(t, *mustChangePasswordPtr(true))
}

func invokeMapError(t *testing.T, err error) (status int, code string) {
	t.Helper()
	app := fiber.New()
	var gotCode string
	app.Get("/", func(c fiber.Ctx) error {
		mapErr := writeMapUsecaseError(c, err)
		status = c.Response().StatusCode()
		if mapErr != nil {
			return mapErr
		}
		var env response.Envelope
		_ = json.Unmarshal(c.Response().Body(), &env)
		if env.Error != nil {
			gotCode = env.Error.Code
		}
		return nil
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	resp, reqErr := app.Test(req)
	require.NoError(t, reqErr)
	defer func() { _ = resp.Body.Close() }()
	if gotCode != "" {
		return status, gotCode
	}
	return resp.StatusCode, decodeResponseCode(t, resp)
}

func invokeMapErrorViaErrorHandler(t *testing.T, err error) (status int, code string) {
	t.Helper()
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zap.NewNop())})
	app.Get("/", func(c fiber.Ctx) error {
		return writeMapUsecaseError(c, err)
	})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	resp, reqErr := app.Test(req)
	require.NoError(t, reqErr)
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode, decodeResponseCode(t, resp)
}

func decodeResponseCode(t *testing.T, resp *http.Response) string {
	t.Helper()
	raw, readErr := io.ReadAll(resp.Body)
	require.NoError(t, readErr)
	var env response.Envelope
	require.NoError(t, json.Unmarshal(raw, &env))
	require.NotNil(t, env.Error)
	return env.Error.Code
}
