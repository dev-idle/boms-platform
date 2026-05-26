// Package session defines server-side session entities.
package session

import "time"

// Session identifies a user's active login session.
type Session struct {
	UserID    string
	SessionID string
	Meta      SessionMeta
}

// SessionMeta is stored in Redis at session:{userID}:{sessionID}.
//
// MustChangePassword mirrors users.must_change_password at session creation time, allowing
// the RequirePasswordChanged middleware to gate writes from Redis instead of hitting Postgres
// on every mutation. The flag is refreshed on every session rotation (refresh) and reset
// when ChangePassword revokes all sessions.
type SessionMeta struct {
	RefreshJTI         string    `json:"refresh_jti"`
	CreatedAt          time.Time `json:"created_at"`
	UserAgent          string    `json:"user_agent"`
	IP                 string    `json:"ip"`
	MustChangePassword bool      `json:"must_change_password,omitempty"`
}
