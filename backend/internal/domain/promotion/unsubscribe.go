package promotion

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"github.com/google/uuid"
)

// UnsubscribeTokens signs a customer's id into the token their promotion
// emails link to, so the link withdraws their agreement without signing in and
// nothing is stored: a copy of the database holds no token, and none can be
// made for another customer without the key.
type UnsubscribeTokens struct {
	key []byte
}

func NewUnsubscribeTokens(key string) UnsubscribeTokens {
	return UnsubscribeTokens{key: []byte(key)}
}

// Of is the customer's token: their id and its signature, as one URL-safe
// string like every other emailed link's.
func (t UnsubscribeTokens) Of(userID uuid.UUID) string {
	return base64.RawURLEncoding.EncodeToString(append(userID[:], t.sign(userID)...))
}

// UserOf is the customer token was made for. It checks the signature in
// constant time and returns ErrInvalidUnsubscribeToken for any token it did
// not make.
func (t UnsubscribeTokens) UserOf(token string) (uuid.UUID, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != len(uuid.UUID{})+sha256.Size {
		return uuid.Nil, ErrInvalidUnsubscribeToken
	}
	userID := uuid.UUID(raw[:len(uuid.UUID{})])
	if !hmac.Equal(raw[len(uuid.UUID{}):], t.sign(userID)) {
		return uuid.Nil, ErrInvalidUnsubscribeToken
	}
	return userID, nil
}

func (t UnsubscribeTokens) sign(userID uuid.UUID) []byte {
	mac := hmac.New(sha256.New, t.key)
	mac.Write(userID[:])
	return mac.Sum(nil)
}
