package conversation

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestReplyEmailFor(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	event := RepliesUnreadEvent(id)

	got, ok := ReplyEmailFor(event)
	assert.True(t, ok)
	assert.Equal(t, id, got)
	assert.Empty(t, event.Audience.UserIDs, "it reaches the email queue, never an open page")
	assert.Empty(t, event.Audience.Roles)
	assert.False(t, event.Audience.Public)

	_, ok = ReplyEmailFor(MessageCreatedEvent(uuid.New(), nil))
	assert.False(t, ok, "only the replies left unread ask for it")
}
