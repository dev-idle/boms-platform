package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainrealtime "github.com/boms/backend/internal/domain/realtime"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

var _ port.RealtimeTicketStore = (*stubTicketStore)(nil)

type stubTicketStore struct {
	issued domainrealtime.Ticket
	ttl    time.Duration
	err    error
}

func (s *stubTicketStore) Issue(_ context.Context, t domainrealtime.Ticket, ttl time.Duration) (string, error) {
	s.issued, s.ttl = t, ttl
	return "token", s.err
}

func (s *stubTicketStore) Consume(context.Context, string) (domainrealtime.Ticket, error) {
	return domainrealtime.Ticket{}, errors.New("not used by the usecase")
}

func TestRealtimeUsecase_IssueTicket(t *testing.T) {
	t.Parallel()
	ticket := domainrealtime.Ticket{UserID: uuid.New(), Role: domainuser.RoleBaker, SessionID: uuid.New()}

	t.Run("issues_a_ticket_for_the_session", func(t *testing.T) {
		t.Parallel()
		store := &stubTicketStore{}
		uc := usecase.NewRealtimeUsecase(store, "wss://app.example.com/ws", 30*time.Second)

		out, err := uc.IssueTicket(context.Background(), ticket)

		require.NoError(t, err)
		assert.Equal(t, ticket, store.issued)
		assert.Equal(t, 30*time.Second, store.ttl)
		assert.Equal(t, "token", out.Ticket)
		assert.Equal(t, "wss://app.example.com/ws", out.URL)
	})

	t.Run("fails_when_the_store_fails", func(t *testing.T) {
		t.Parallel()
		uc := usecase.NewRealtimeUsecase(&stubTicketStore{err: errors.New("redis down")}, "wss://app.example.com/ws", 30*time.Second)

		_, err := uc.IssueTicket(context.Background(), ticket)

		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, apperrors.ErrInternal.Code, appErr.Code)
	})
}
