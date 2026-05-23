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
type SessionMeta struct {
	RefreshJTI string    `json:"refresh_jti"`
	CreatedAt  time.Time `json:"created_at"`
	UserAgent  string    `json:"user_agent"`
	IP         string    `json:"ip"`
}
