package redis

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	domainrealtime "github.com/boms/backend/internal/domain/realtime"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
)

const realtimeTicketKeyPrefix = "realtime_ticket:"

// RealtimeTicketStore keeps tickets in Redis under a hash of their token, so a
// Redis dump never holds a usable ticket.
type RealtimeTicketStore struct {
	rdb *redis.Client
}

func NewRealtimeTicketStore(client *Client) *RealtimeTicketStore {
	return &RealtimeTicketStore{rdb: client.RDB()}
}

type ticketRecord struct {
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	SessionID uuid.UUID `json:"session_id"`
}

func realtimeTicketKey(token string) string {
	sum := sha256.Sum256([]byte(token))
	return realtimeTicketKeyPrefix + hex.EncodeToString(sum[:])
}

func (s *RealtimeTicketStore) Issue(ctx context.Context, t domainrealtime.Ticket, ttl time.Duration) (string, error) {
	raw := make([]byte, domainrealtime.TicketBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate realtime ticket: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	record, err := json.Marshal(ticketRecord{UserID: t.UserID, Role: string(t.Role), SessionID: t.SessionID})
	if err != nil {
		return "", fmt.Errorf("encode realtime ticket: %w", err)
	}
	if err := s.rdb.Set(ctx, realtimeTicketKey(token), record, ttl).Err(); err != nil {
		return "", fmt.Errorf("store realtime ticket: %w", err)
	}
	return token, nil
}

func (s *RealtimeTicketStore) Consume(ctx context.Context, token string) (domainrealtime.Ticket, error) {
	// Anything not shaped like an issued token is refused without a round trip.
	if !domainrealtime.WellFormedToken(token) {
		return domainrealtime.Ticket{}, domainrealtime.ErrTicketInvalid
	}
	raw, err := s.rdb.GetDel(ctx, realtimeTicketKey(token)).Bytes()
	if errors.Is(err, redis.Nil) {
		return domainrealtime.Ticket{}, domainrealtime.ErrTicketInvalid
	}
	if err != nil {
		return domainrealtime.Ticket{}, fmt.Errorf("redeem realtime ticket: %w", err)
	}
	var record ticketRecord
	if err := json.Unmarshal(raw, &record); err != nil {
		return domainrealtime.Ticket{}, fmt.Errorf("decode realtime ticket: %w", err)
	}
	return domainrealtime.Ticket{
		UserID:    record.UserID,
		Role:      domainuser.Role(record.Role),
		SessionID: record.SessionID,
	}, nil
}

var _ port.RealtimeTicketStore = (*RealtimeTicketStore)(nil)
