package v1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainrealtime "github.com/boms/backend/internal/domain/realtime"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/middleware"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

// acceptingSigner reads every bearer token as the same session.
type acceptingSigner struct {
	port.TokenSigner
	claims port.AccessTokenClaims
}

func (s acceptingSigner) ParseAccess(string) (port.AccessTokenClaims, error) {
	return s.claims, nil
}

type recordingTicketStore struct {
	issued domainrealtime.Ticket
}

func (s *recordingTicketStore) Issue(_ context.Context, t domainrealtime.Ticket, _ time.Duration) (string, error) {
	s.issued = t
	return "ticket-token", nil
}

func (s *recordingTicketStore) Consume(context.Context, string) (domainrealtime.Ticket, error) {
	return domainrealtime.Ticket{}, domainrealtime.ErrTicketInvalid
}

func TestRealtimeHandler_IssueTicket(t *testing.T) {
	t.Parallel()

	post := func(t *testing.T, app *fiber.App) *http.Response {
		t.Helper()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/realtime/tickets", nil)
		req.Header.Set(fiber.HeaderAuthorization, "Bearer access-token")
		resp, err := app.Test(req)
		require.NoError(t, err)
		return resp
	}

	t.Run("refuses_a_request_without_a_session", func(t *testing.T) {
		t.Parallel()
		handler := NewRealtimeHandler(usecase.NewRealtimeUsecase(&recordingTicketStore{}, "ws://localhost:8081/ws", 30*time.Second))
		app := fiber.New()
		app.Post("/realtime/tickets", handler.IssueTicket)

		resp := post(t, app)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("issues_a_ticket_for_the_session_that_nothing_caches", func(t *testing.T) {
		t.Parallel()
		userID, sessionID := uuid.New(), uuid.New()
		store := &recordingTicketStore{}
		handler := NewRealtimeHandler(usecase.NewRealtimeUsecase(store, "ws://localhost:8081/ws", 30*time.Second))
		app := fiber.New()
		signer := acceptingSigner{claims: port.AccessTokenClaims{
			Subject: userID.String(), Role: string(domainuser.RoleStaff), SessionID: sessionID.String(),
		}}
		app.Post("/realtime/tickets", middleware.RequireAuth(signer), handler.IssueTicket)

		resp := post(t, app)
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, "no-store", resp.Header.Get(fiber.HeaderCacheControl))
		assert.Equal(t, domainrealtime.Ticket{UserID: userID, Role: domainuser.RoleStaff, SessionID: sessionID}, store.issued)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		var envelope struct {
			Data struct {
				Ticket string `json:"ticket"`
				URL    string `json:"url"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(body, &envelope))
		assert.Equal(t, "ticket-token", envelope.Data.Ticket)
		assert.Equal(t, "ws://localhost:8081/ws", envelope.Data.URL)
	})
}
