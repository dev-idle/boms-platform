package promotion

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainuser "github.com/boms/backend/internal/domain/user"
)

func TestNewSubject(t *testing.T) {
	t.Parallel()
	subject, err := NewSubject("  Matcha week: 10% off  ")
	require.NoError(t, err)
	assert.Equal(t, "Matcha week: 10% off", subject)

	longest := strings.Repeat("é", MaxSubjectLength)
	subject, err = NewSubject(longest)
	require.NoError(t, err)
	assert.Equal(t, longest, subject, "the limit counts characters, not bytes")

	for _, bad := range []string{"", "   ", longest + "a", "Two\nlines", "Hidden" + string(rune(0x202e))} {
		_, err := NewSubject(bad)
		assert.ErrorIs(t, err, ErrInvalidSubject, "%q", bad)
	}
}

func TestNewBody(t *testing.T) {
	t.Parallel()
	body, err := NewBody("  New this week:\r\nmatcha éclairs 🍵  ")
	require.NoError(t, err)
	assert.Equal(t, "New this week:\nmatcha éclairs 🍵", body, "trimmed, one kind of line break")

	longest := strings.Repeat("é", MaxBodyLength)
	_, err = NewBody(longest)
	require.NoError(t, err)

	for _, bad := range []string{"", " \n ", longest + "a", "tab\there"} {
		_, err := NewBody(bad)
		assert.ErrorIs(t, err, ErrInvalidBody, "%q", bad)
	}
}

func TestUnsubscribeTokens(t *testing.T) {
	t.Parallel()
	tokens := NewUnsubscribeTokens(strings.Repeat("u", 32))
	customer := uuid.New()

	token := tokens.Of(customer)
	assert.Regexp(t, `^[A-Za-z0-9_-]{64}$`, token, "URL-safe like every emailed link's token")
	got, err := tokens.UserOf(token)
	require.NoError(t, err)
	assert.Equal(t, customer, got)

	other := tokens.Of(uuid.New())
	forged := []string{
		"",
		customer.String(),
		token[:63],
		token + "A",
		other[:22] + token[22:],
		"not base64!",
	}
	for _, bad := range forged {
		_, err := tokens.UserOf(bad)
		assert.ErrorIs(t, err, ErrInvalidUnsubscribeToken, "%q", bad)
	}

	_, err = NewUnsubscribeTokens(strings.Repeat("v", 32)).UserOf(token)
	require.ErrorIs(t, err, ErrInvalidUnsubscribeToken, "a token made with another key")
}

func TestEvents(t *testing.T) {
	t.Parallel()
	promotionID := uuid.New()

	created := CreatedEvent(promotionID)
	assert.Equal(t, TopicCreated, created.Topic)
	assert.Equal(t, []domainuser.Role{domainuser.RoleManager}, created.Audience.Roles)
	assert.Empty(t, created.Audience.UserIDs)
	got, ok := ToSend(created)
	require.True(t, ok, "a promotion created is one to email")
	assert.Equal(t, promotionID, got)

	sent := SentEvent(promotionID)
	assert.Equal(t, TopicSent, sent.Topic)
	_, ok = ToSend(sent)
	assert.False(t, ok, "a promotion sent is emailed already")

	_, ok = ToSend(domainevent.New(TopicCreated, domainevent.Audience{}, map[string]string{"promotion_id": "nope"}))
	assert.False(t, ok)
}
