package realtime_test

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	domainrealtime "github.com/boms/backend/internal/domain/realtime"
)

func TestWellFormedToken(t *testing.T) {
	t.Parallel()

	raw := make([]byte, domainrealtime.TicketBytes)
	_, _ = rand.Read(raw)
	issued := base64.RawURLEncoding.EncodeToString(raw)
	assert.True(t, domainrealtime.WellFormedToken(issued))

	for name, token := range map[string]string{
		"empty":     "",
		"too_short": issued[:len(issued)-1],
		"too_long":  issued + "A",
		"padded":    issued[:len(issued)-1] + "=",
		"not_url":   strings.Repeat("+", len(issued)),
	} {
		assert.False(t, domainrealtime.WellFormedToken(token), name)
	}
}
