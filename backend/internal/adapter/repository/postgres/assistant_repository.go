package postgres

import (
	"context"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainassistant "github.com/boms/backend/internal/domain/assistant"
	"github.com/boms/backend/internal/port"
)

// AssistantRepository keeps the assistant's sessions and their turns.
type AssistantRepository struct {
	queries *sqlcgen.Queries
}

func NewAssistantRepository(pool *Pool) *AssistantRepository {
	return &AssistantRepository{queries: pool.Queries()}
}

func (r *AssistantRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

// RecentTurns implements port.AssistantRepository.
func (r *AssistantRepository) RecentTurns(ctx context.Context, sessionID uuid.UUID, userID *uuid.UUID, limit int32) ([]domainassistant.Turn, error) {
	rows, err := r.q(ctx).ListRecentAssistantTurns(ctx, sqlcgen.ListRecentAssistantTurnsParams{
		SessionID: sessionID,
		UserID:    userID,
		Limit:     limit,
	})
	if err != nil {
		return nil, mapRepoError(err, "list recent assistant turns")
	}
	out := make([]domainassistant.Turn, 0, len(rows))
	for _, row := range rows {
		out = append(out, domainassistant.Turn{Role: domainassistant.Role(row.Role), Body: row.Body, At: row.CreatedAt})
	}
	slices.Reverse(out)
	return out, nil
}

// SaveTurn implements port.AssistantRepository. It runs inside the caller's
// transaction, so the session and both turns are written together.
func (r *AssistantRepository) SaveTurn(ctx context.Context, sessionID uuid.UUID, userID *uuid.UUID, question, answer string) error {
	if _, err := r.q(ctx).TouchAssistantSession(ctx, sqlcgen.TouchAssistantSessionParams{ID: sessionID, UserID: userID}); err != nil {
		return mapRepoError(err, "touch assistant session")
	}
	err := r.q(ctx).CreateAssistantTurn(ctx, sqlcgen.CreateAssistantTurnParams{
		SessionID: sessionID,
		Question:  question,
		Answer:    answer,
	})
	if err != nil {
		return mapRepoError(err, "create assistant turn")
	}
	return nil
}

// DeleteIdle implements port.AssistantRepository.
func (r *AssistantRepository) DeleteIdle(ctx context.Context, retention time.Duration) (int64, error) {
	deleted, err := r.q(ctx).DeleteIdleAssistantSessions(ctx, retention.Seconds())
	if err != nil {
		return 0, mapRepoError(err, "delete idle assistant sessions")
	}
	return deleted, nil
}

// ListForUser implements port.AssistantRepository.
func (r *AssistantRepository) ListForUser(ctx context.Context, userID uuid.UUID) ([]domainassistant.Session, error) {
	rows, err := r.q(ctx).ListAssistantTurnsForUser(ctx, &userID)
	if err != nil {
		return nil, mapRepoError(err, "list assistant turns for user")
	}
	out := make([]domainassistant.Session, 0)
	for _, row := range rows {
		if len(out) == 0 || out[len(out)-1].ID != row.SessionID {
			out = append(out, domainassistant.Session{ID: row.SessionID, StartedAt: row.StartedAt})
		}
		session := &out[len(out)-1]
		session.Turns = append(session.Turns, domainassistant.Turn{
			Role: domainassistant.Role(row.Role),
			Body: row.Body,
			At:   row.CreatedAt,
		})
	}
	return out, nil
}

// DeleteForUser implements port.AssistantRepository.
func (r *AssistantRepository) DeleteForUser(ctx context.Context, userID uuid.UUID) error {
	if err := r.q(ctx).DeleteAssistantSessionsForUser(ctx, &userID); err != nil {
		return mapRepoError(err, "delete assistant sessions for user")
	}
	return nil
}

var _ port.AssistantRepository = (*AssistantRepository)(nil)
