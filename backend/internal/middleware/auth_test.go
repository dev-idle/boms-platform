package middleware_test

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

type mockTokenSigner struct{ mock.Mock }

func (m *mockTokenSigner) SignAccess(in port.AccessTokenClaims) (string, error) {
	args := m.Called(in)
	return args.String(0), args.Error(1)
}
func (m *mockTokenSigner) SignRefresh(in port.RefreshTokenClaims) (string, error) {
	args := m.Called(in)
	return args.String(0), args.Error(1)
}
func (m *mockTokenSigner) ParseAccess(token string) (port.AccessTokenClaims, error) {
	args := m.Called(token)
	claims, _ := args.Get(0).(port.AccessTokenClaims)
	return claims, args.Error(1)
}
func (m *mockTokenSigner) ParseRefresh(token string) (port.RefreshTokenClaims, error) {
	args := m.Called(token)
	claims, _ := args.Get(0).(port.RefreshTokenClaims)
	return claims, args.Error(1)
}

type mockSessionStore struct{ mock.Mock }

func (m *mockSessionStore) Create(ctx context.Context, userID, sessionID string, meta domainsession.SessionMeta) error {
	return m.Called(ctx, userID, sessionID, meta).Error(0)
}
func (m *mockSessionStore) Get(ctx context.Context, userID, sessionID string) (domainsession.SessionMeta, error) {
	args := m.Called(ctx, userID, sessionID)
	meta, _ := args.Get(0).(domainsession.SessionMeta)
	return meta, args.Error(1)
}
func (m *mockSessionStore) Delete(ctx context.Context, userID, sessionID string) error {
	return m.Called(ctx, userID, sessionID).Error(0)
}
func (m *mockSessionStore) DeleteAllForUser(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockSessionStore) Rotate(ctx context.Context, userID, oldSessionID, newSessionID, expectedRefreshJTI string, meta domainsession.SessionMeta) error {
	return m.Called(ctx, userID, oldSessionID, newSessionID, expectedRefreshJTI, meta).Error(0)
}

func testGet(t *testing.T, path string) *http.Request {
	t.Helper()
	return httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
}

func withTestResponse(t *testing.T, app *fiber.App, req *http.Request, fn func(t *testing.T, resp *http.Response)) {
	t.Helper()
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer func() { _ = resp.Body.Close() }()
	fn(t, resp)
}

func TestRequireAuth(t *testing.T) {
	t.Parallel()
	uid := uuid.New()
	sid := uuid.New()

	t.Run("valid token passes and sets locals", func(t *testing.T) {
		t.Parallel()
		signer := new(mockTokenSigner)
		signer.On("ParseAccess", "good-token").Return(port.AccessTokenClaims{
			Subject: uid.String(), Role: "customer", SessionID: sid.String(), JTI: "j1",
		}, nil)

		var gotUser, gotSession uuid.UUID
		var gotRole domainuser.Role
		app := fiber.New()
		app.Get("/", middleware.RequireAuth(signer), func(c *fiber.Ctx) error {
			var ok bool
			gotUser, ok = middleware.GetUserID(c)
			require.True(t, ok)
			gotRole, ok = middleware.GetRole(c)
			require.True(t, ok)
			gotSession, ok = middleware.GetSessionID(c)
			require.True(t, ok)
			return c.SendStatus(fiber.StatusOK)
		})

		req := testGet(t, "/")
		req.Header.Set(fiber.HeaderAuthorization, "Bearer good-token")
		withTestResponse(t, app, req, func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		assert.Equal(t, uid, gotUser)
		assert.Equal(t, sid, gotSession)
		assert.Equal(t, domainuser.RoleCustomer, gotRole)
	})

	t.Run("missing token rejects", func(t *testing.T) {
		t.Parallel()
		app := fiber.New()
		app.Get("/", middleware.RequireAuth(new(mockTokenSigner)), func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		withTestResponse(t, app, testGet(t, "/"), func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			assertErrorCode(t, resp, "unauthorized")
		})
	})

	t.Run("expired token rejects", func(t *testing.T) {
		t.Parallel()
		signer := new(mockTokenSigner)
		signer.On("ParseAccess", "expired").Return(port.AccessTokenClaims{}, apperrors.ErrUnauthorized)

		app := fiber.New()
		app.Get("/", middleware.RequireAuth(signer), func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		req := testGet(t, "/")
		req.Header.Set(fiber.HeaderAuthorization, "Bearer expired")
		withTestResponse(t, app, req, func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})

	t.Run("tampered token rejects", func(t *testing.T) {
		t.Parallel()
		signer := new(mockTokenSigner)
		signer.On("ParseAccess", "bad").Return(port.AccessTokenClaims{}, apperrors.ErrUnauthorized)

		app := fiber.New()
		app.Get("/", middleware.RequireAuth(signer), func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		req := testGet(t, "/")
		req.Header.Set(fiber.HeaderAuthorization, "Bearer bad")
		withTestResponse(t, app, req, func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})
}

func TestRequireAuthWithSession(t *testing.T) {
	t.Parallel()
	uid := uuid.New()
	sid := uuid.New()

	t.Run("valid token and session passes", func(t *testing.T) {
		t.Parallel()
		signer := new(mockTokenSigner)
		sessions := new(mockSessionStore)
		signer.On("ParseAccess", "tok").Return(port.AccessTokenClaims{
			Subject: uid.String(), Role: "customer", SessionID: sid.String(), JTI: "j",
		}, nil)
		sessions.On("Get", mock.Anything, uid.String(), sid.String()).Return(domainsession.SessionMeta{RefreshJTI: "j"}, nil)

		app := fiber.New()
		app.Get("/", middleware.RequireAuthWithSession(signer, sessions), func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		req := testGet(t, "/")
		req.Header.Set(fiber.HeaderAuthorization, "Bearer tok")
		withTestResponse(t, app, req, func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	})

	t.Run("deleted session rejects session_revoked", func(t *testing.T) {
		t.Parallel()
		signer := new(mockTokenSigner)
		sessions := new(mockSessionStore)
		signer.On("ParseAccess", "tok").Return(port.AccessTokenClaims{
			Subject: uid.String(), Role: "customer", SessionID: sid.String(), JTI: "j",
		}, nil)
		sessions.On("Get", mock.Anything, uid.String(), sid.String()).Return(domainsession.SessionMeta{}, apperrors.ErrNotFound)

		app := fiber.New()
		app.Get("/", middleware.RequireAuthWithSession(signer, sessions), func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		req := testGet(t, "/")
		req.Header.Set(fiber.HeaderAuthorization, "Bearer tok")
		withTestResponse(t, app, req, func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			assertErrorCode(t, resp, "session_revoked")
		})
	})

	t.Run("redis error returns 500", func(t *testing.T) {
		t.Parallel()
		signer := new(mockTokenSigner)
		sessions := new(mockSessionStore)
		signer.On("ParseAccess", "tok").Return(port.AccessTokenClaims{
			Subject: uid.String(), Role: "customer", SessionID: sid.String(), JTI: "j",
		}, nil)
		sessions.On("Get", mock.Anything, uid.String(), sid.String()).Return(domainsession.SessionMeta{}, apperrors.Errorf("redis down"))

		app := fiber.New()
		app.Get("/", middleware.RequireAuthWithSession(signer, sessions), func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		req := testGet(t, "/")
		req.Header.Set(fiber.HeaderAuthorization, "Bearer tok")
		withTestResponse(t, app, req, func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		})
	})
}

func TestOptionalAuth(t *testing.T) {
	t.Parallel()
	uid := uuid.New()
	sid := uuid.New()

	t.Run("valid token sets locals", func(t *testing.T) {
		t.Parallel()
		signer := new(mockTokenSigner)
		signer.On("ParseAccess", "tok").Return(port.AccessTokenClaims{
			Subject: uid.String(), Role: "customer", SessionID: sid.String(), JTI: "j",
		}, nil)

		var got bool
		app := fiber.New()
		app.Get("/", middleware.OptionalAuth(signer), func(c *fiber.Ctx) error {
			_, got = middleware.GetUserID(c)
			return c.SendStatus(fiber.StatusOK)
		})
		req := testGet(t, "/")
		req.Header.Set(fiber.HeaderAuthorization, "Bearer tok")
		withTestResponse(t, app, req, func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.True(t, got)
		})
	})

	t.Run("missing token continues without locals", func(t *testing.T) {
		t.Parallel()
		var got bool
		app := fiber.New()
		app.Get("/", middleware.OptionalAuth(new(mockTokenSigner)), func(c *fiber.Ctx) error {
			_, got = middleware.GetUserID(c)
			return c.SendStatus(fiber.StatusOK)
		})
		withTestResponse(t, app, testGet(t, "/"), func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.False(t, got)
		})
	})

	t.Run("invalid token continues without locals", func(t *testing.T) {
		t.Parallel()
		signer := new(mockTokenSigner)
		signer.On("ParseAccess", "bad").Return(port.AccessTokenClaims{}, apperrors.ErrUnauthorized)

		var got bool
		app := fiber.New()
		app.Get("/", middleware.OptionalAuth(signer), func(c *fiber.Ctx) error {
			_, got = middleware.GetUserID(c)
			return c.SendStatus(fiber.StatusOK)
		})
		req := testGet(t, "/")
		req.Header.Set(fiber.HeaderAuthorization, "Bearer bad")
		withTestResponse(t, app, req, func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.False(t, got)
		})
	})
}

func TestRequireRole(t *testing.T) {
	t.Parallel()

	t.Run("allowed role passes", func(t *testing.T) {
		t.Parallel()
		app := fiber.New()
		app.Get("/", func(c *fiber.Ctx) error {
			c.Locals("auth_role", domainuser.RoleAdmin)
			return c.Next()
		}, middleware.RequireRole(domainuser.RoleAdmin), func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		withTestResponse(t, app, testGet(t, "/"), func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	})

	t.Run("wrong role returns 403", func(t *testing.T) {
		t.Parallel()
		app := fiber.New()
		app.Get("/", func(c *fiber.Ctx) error {
			c.Locals("auth_role", domainuser.RoleCustomer)
			return c.Next()
		}, middleware.RequireRole(domainuser.RoleAdmin), func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		withTestResponse(t, app, testGet(t, "/"), func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusForbidden, resp.StatusCode)
			assertErrorCode(t, resp, "forbidden")
		})
	})
}

func TestRequirePasswordChanged(t *testing.T) {
	t.Parallel()
	userID := uuid.New()
	sessionID := uuid.New()

	t.Run("uses cached session meta (locals) without store lookup", func(t *testing.T) {
		t.Parallel()
		store := new(mockSessionStore)

		app := fiber.New()
		app.Get("/", func(c *fiber.Ctx) error {
			c.Locals("auth_user_id", userID)
			c.Locals("auth_session_id", sessionID)
			c.Locals("auth_session_meta", domainsession.SessionMeta{MustChangePassword: false})
			return c.Next()
		}, middleware.RequirePasswordChanged(store), func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		withTestResponse(t, app, testGet(t, "/"), func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		store.AssertNotCalled(t, "Get", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("blocks when cached meta says must_change_password=true", func(t *testing.T) {
		t.Parallel()
		store := new(mockSessionStore)

		app := fiber.New()
		app.Get("/", func(c *fiber.Ctx) error {
			c.Locals("auth_user_id", userID)
			c.Locals("auth_session_id", sessionID)
			c.Locals("auth_session_meta", domainsession.SessionMeta{MustChangePassword: true})
			return c.Next()
		}, middleware.RequirePasswordChanged(store), func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		withTestResponse(t, app, testGet(t, "/"), func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusForbidden, resp.StatusCode)
			assertErrorCode(t, resp, "password_change_required")
		})
	})

	t.Run("falls back to session store when meta not cached", func(t *testing.T) {
		t.Parallel()
		store := new(mockSessionStore)
		store.On("Get", mock.Anything, userID.String(), sessionID.String()).
			Return(domainsession.SessionMeta{MustChangePassword: true}, nil)

		app := fiber.New()
		app.Get("/", func(c *fiber.Ctx) error {
			c.Locals("auth_user_id", userID)
			c.Locals("auth_session_id", sessionID)
			return c.Next()
		}, middleware.RequirePasswordChanged(store), func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		withTestResponse(t, app, testGet(t, "/"), func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusForbidden, resp.StatusCode)
			assertErrorCode(t, resp, "password_change_required")
		})
		store.AssertCalled(t, "Get", mock.Anything, userID.String(), sessionID.String())
	})

	t.Run("returns session_revoked when fallback lookup misses", func(t *testing.T) {
		t.Parallel()
		store := new(mockSessionStore)
		store.On("Get", mock.Anything, userID.String(), sessionID.String()).
			Return(domainsession.SessionMeta{}, apperrors.ErrNotFound)

		app := fiber.New()
		app.Get("/", func(c *fiber.Ctx) error {
			c.Locals("auth_user_id", userID)
			c.Locals("auth_session_id", sessionID)
			return c.Next()
		}, middleware.RequirePasswordChanged(store), func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		withTestResponse(t, app, testGet(t, "/"), func(t *testing.T, resp *http.Response) {
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			assertErrorCode(t, resp, "session_revoked")
		})
	})
}

func assertErrorCode(t *testing.T, resp *http.Response, code string) {
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
