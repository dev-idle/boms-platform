package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	domainsession "github.com/boms/backend/internal/domain/session"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/middleware"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type routeTestTokenSigner struct{ mock.Mock }

func (m *routeTestTokenSigner) SignAccess(in port.AccessTokenClaims) (string, error) {
	args := m.Called(in)
	return args.String(0), args.Error(1)
}
func (m *routeTestTokenSigner) SignRefresh(in port.RefreshTokenClaims) (string, error) {
	args := m.Called(in)
	return args.String(0), args.Error(1)
}
func (m *routeTestTokenSigner) ParseAccess(token string) (port.AccessTokenClaims, error) {
	args := m.Called(token)
	claims, _ := args.Get(0).(port.AccessTokenClaims)
	return claims, args.Error(1)
}
func (m *routeTestTokenSigner) ParseRefresh(token string) (port.RefreshTokenClaims, error) {
	args := m.Called(token)
	claims, _ := args.Get(0).(port.RefreshTokenClaims)
	return claims, args.Error(1)
}

type routeTestSessionStore struct{ mock.Mock }

func (m *routeTestSessionStore) Create(ctx context.Context, userID, sessionID string, meta domainsession.SessionMeta) error {
	return m.Called(ctx, userID, sessionID, meta).Error(0)
}
func (m *routeTestSessionStore) Get(ctx context.Context, userID, sessionID string) (domainsession.SessionMeta, error) {
	args := m.Called(ctx, userID, sessionID)
	meta, _ := args.Get(0).(domainsession.SessionMeta)
	return meta, args.Error(1)
}
func (m *routeTestSessionStore) Delete(ctx context.Context, userID, sessionID string) error {
	return m.Called(ctx, userID, sessionID).Error(0)
}
func (m *routeTestSessionStore) DeleteAllForUser(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *routeTestSessionStore) Rotate(ctx context.Context, userID, oldSessionID, newSessionID, expectedRefreshJTI string, meta domainsession.SessionMeta) error {
	return m.Called(ctx, userID, oldSessionID, newSessionID, expectedRefreshJTI, meta).Error(0)
}

func TestCatalogReadIsPublic(t *testing.T) {
	t.Parallel()

	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) }

	app := fiber.New()
	apiV1 := app.Group("/api/v1")
	catalogRead := apiV1.Group("/catalog")
	catalogRead.Get("/products", okHandler)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/api/v1/catalog/products",
		nil,
	)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCustomerSessionGroupDoesNotLeakRoleMiddleware(t *testing.T) {
	t.Parallel()

	uid := uuid.New()
	sid := uuid.New()
	signer := new(routeTestTokenSigner)
	sessions := new(routeTestSessionStore)
	signer.On("ParseAccess", "manager-token").Return(port.AccessTokenClaims{
		Subject: uid.String(), Role: string(domainuser.RoleManager), SessionID: sid.String(), JTI: "j",
	}, nil)
	sessions.On("Get", mock.Anything, uid.String(), sid.String()).
		Return(domainsession.SessionMeta{RefreshJTI: "j"}, nil)

	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) }
	passwordChanged := middleware.RequirePasswordChanged(sessions)

	app := fiber.New()
	apiV1 := app.Group("/api/v1")

	customerCart := customerSessionGroup(apiV1, "/cart", signer, sessions, passwordChanged)
	customerCart.Get("", okHandler)

	customerOrders := customerSessionGroup(apiV1, "/orders", signer, sessions, passwordChanged)
	customerOrders.Get("", okHandler)

	managerRead := apiV1.Group(
		"/manager",
		middleware.RequireAuthWithSession(signer, sessions),
		middleware.RequireRole(domainuser.RoleManager),
		passwordChanged,
	)
	managerRead.Get("/categories", okHandler)

	t.Run("manager_reaches_manager_route", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/manager/categories", nil)
		req.Header.Set(fiber.HeaderAuthorization, "Bearer manager-token")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("manager_blocked_on_customer_cart", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/cart", nil)
		req.Header.Set(fiber.HeaderAuthorization, "Bearer manager-token")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		assertRouteErrorCode(t, resp, apperrors.ErrForbidden.Code)
	})
}

func assertRouteErrorCode(t *testing.T, resp *http.Response, code string) {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var env struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(body, &env))
	assert.False(t, env.Success)
	assert.Equal(t, code, env.Error.Code)
}
