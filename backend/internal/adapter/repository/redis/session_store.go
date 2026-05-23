package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domainsession "github.com/boms/backend/internal/domain/session"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	goredis "github.com/redis/go-redis/v9"
)

const sessionKeyPrefix = "session:"

// SessionStore implements port.SessionStore using Redis keys session:{userID}:{sessionID}.
type SessionStore struct {
	rdb *goredis.Client
	ttl time.Duration
}

// NewSessionStore returns a Redis-backed session store.
func NewSessionStore(client *Client, ttl time.Duration) *SessionStore {
	return &SessionStore{rdb: client.RDB(), ttl: ttl}
}

func sessionKey(userID, sessionID string) string {
	return sessionKeyPrefix + userID + ":" + sessionID
}

func scanPattern(userID string) string {
	return sessionKeyPrefix + userID + ":*"
}

// Create implements port.SessionStore.
func (s *SessionStore) Create(ctx context.Context, userID, sessionID string, meta domainsession.SessionMeta) error {
	if userID == "" || sessionID == "" {
		return apperrors.ErrValidation.WithDetail("reason", "user_id and session_id are required")
	}
	if meta.RefreshJTI == "" {
		return apperrors.ErrValidation.WithDetail("field", "refresh_jti")
	}
	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = time.Now().UTC()
	}
	payload, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	key := sessionKey(userID, sessionID)
	if err := s.rdb.Set(ctx, key, payload, s.ttl).Err(); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// Get implements port.SessionStore.
func (s *SessionStore) Get(ctx context.Context, userID, sessionID string) (domainsession.SessionMeta, error) {
	if userID == "" || sessionID == "" {
		return domainsession.SessionMeta{}, apperrors.ErrValidation.WithDetail("reason", "user_id and session_id are required")
	}
	key := sessionKey(userID, sessionID)
	raw, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == goredis.Nil {
			return domainsession.SessionMeta{}, apperrors.ErrNotFound
		}
		return domainsession.SessionMeta{}, fmt.Errorf("get session: %w", err)
	}
	var meta domainsession.SessionMeta
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return domainsession.SessionMeta{}, fmt.Errorf("unmarshal session: %w", err)
	}
	return meta, nil
}

// Delete implements port.SessionStore.
func (s *SessionStore) Delete(ctx context.Context, userID, sessionID string) error {
	if userID == "" || sessionID == "" {
		return apperrors.ErrValidation.WithDetail("reason", "user_id and session_id are required")
	}
	key := sessionKey(userID, sessionID)
	if err := s.rdb.Unlink(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteAllForUser implements port.SessionStore using SCAN (never KEYS) and UNLINK.
func (s *SessionStore) DeleteAllForUser(ctx context.Context, userID string) error {
	if userID == "" {
		return apperrors.ErrValidation.WithDetail("field", "user_id")
	}
	pattern := scanPattern(userID)
	var cursor uint64
	for {
		keys, next, err := s.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("scan sessions: %w", err)
		}
		if len(keys) > 0 {
			if err := s.rdb.Unlink(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("unlink sessions: %w", err)
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}

var _ port.SessionStore = (*SessionStore)(nil)
