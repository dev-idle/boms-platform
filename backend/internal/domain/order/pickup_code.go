package order

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
)

// PickupCodes makes the four digits a customer gives at the counter to collect
// an order. A code is the order's id signed with a secret key, so no code is
// stored anywhere — a copy of the database holds none — and none can be
// worked out from an order without the key. The customer sees it; staff never
// do, and ask for it at handoff.
type PickupCodes struct {
	key []byte
}

func NewPickupCodes(key string) PickupCodes {
	return PickupCodes{key: []byte(key)}
}

// Of is the order's pickup code.
func (p PickupCodes) Of(orderID uuid.UUID) string {
	mac := hmac.New(sha256.New, p.key)
	mac.Write(orderID[:])
	return fmt.Sprintf("%04d", binary.BigEndian.Uint32(mac.Sum(nil))%10000)
}

// Matches reports, in constant time, whether code is the order's pickup code.
func (p PickupCodes) Matches(orderID uuid.UUID, code string) bool {
	return hmac.Equal([]byte(p.Of(orderID)), []byte(code))
}

// HasPickupCode reports whether an order in this status is collected with its
// pickup code: from when the bakery accepts it until it is handed over.
func (s Status) HasPickupCode() bool {
	switch s {
	case StatusConfirmed, StatusInProduction, StatusReady:
		return true
	default:
		return false
	}
}
