package usecase

import (
	"context"
	"time"

	domainrealtime "github.com/boms/backend/internal/domain/realtime"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// RealtimeUsecase hands a signed-in session the ticket that opens its push
// connection. The browser cannot set headers on a WebSocket and the listener
// never trusts cookies, so this ticket — obtained through the BFF — is the only
// way in.
type RealtimeUsecase struct {
	tickets   port.RealtimeTicketStore
	publicURL string
	ttl       time.Duration
}

func NewRealtimeUsecase(tickets port.RealtimeTicketStore, publicURL string, ttl time.Duration) *RealtimeUsecase {
	return &RealtimeUsecase{tickets: tickets, publicURL: publicURL, ttl: ttl}
}

// IssueTicket returns a single-use ticket for the caller's session and where to
// redeem it.
func (u *RealtimeUsecase) IssueTicket(ctx context.Context, ticket domainrealtime.Ticket) (*dto.RealtimeTicketResponse, error) {
	token, err := u.tickets.Issue(ctx, ticket, u.ttl)
	if err != nil {
		return nil, apperrors.Errorf("issue realtime ticket: %w", err)
	}
	return &dto.RealtimeTicketResponse{
		Ticket: token,
		URL:    u.publicURL,
	}, nil
}
