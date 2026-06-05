package v1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	domaincategory "github.com/boms/backend/internal/domain/category"
	domaincombo "github.com/boms/backend/internal/domain/combo"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	domainproduct "github.com/boms/backend/internal/domain/product"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
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
		{name: "employee_code_exists", err: domainuser.ErrEmployeeCodeExists, wantStatus: 409, wantCode: "employee_code_exists"},
		{name: "email_exists", err: usecase.ErrEmailExists, wantStatus: 409, wantCode: "email_exists"},
		{name: "category_has_products", err: domaincategory.ErrHasProducts, wantStatus: 422, wantCode: "category_has_products"},
		{name: "category_slug_exists", err: domaincategory.ErrSlugExists, wantStatus: 409, wantCode: "slug_exists"},
		{name: "product_not_found", err: domainproduct.ErrNotFound, wantStatus: 404, wantCode: "not_found"},
		{name: "combo_not_found", err: domaincombo.ErrNotFound, wantStatus: 404, wantCode: "not_found"},
		{name: "combo_slug_exists", err: domaincombo.ErrSlugExists, wantStatus: 409, wantCode: "slug_exists"},
		{name: "discount_code_not_found", err: domaindiscount.ErrNotFound, wantStatus: 404, wantCode: "not_found"},
		{name: "discount_code_exists", err: domaindiscount.ErrCodeExists, wantStatus: 409, wantCode: "code_exists"},
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
	app.Get("/", func(c *fiber.Ctx) error {
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
	app.Get("/", func(c *fiber.Ctx) error {
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
