package notification

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainconversation "github.com/boms/backend/internal/domain/conversation"
	domainevent "github.com/boms/backend/internal/domain/event"
	"github.com/boms/backend/internal/port"
)

type recordingReplyQueue struct{ tasks []port.ReplyEmailTask }

func (q *recordingReplyQueue) EnqueueReplyEmail(_ context.Context, task port.ReplyEmailTask) error {
	q.tasks = append(q.tasks, task)
	return nil
}

func TestReplyEmails_Publish(t *testing.T) {
	t.Parallel()
	queue := &recordingReplyQueue{}
	conversationID := uuid.New()
	asked := domainconversation.RepliesUnreadEvent(conversationID)

	require.NoError(t, NewReplyEmails(queue).Publish(context.Background(), []domainevent.Event{
		domainconversation.MessageCreatedEvent(uuid.New(), nil),
		asked,
	}))

	assert.Equal(t, []port.ReplyEmailTask{{EventID: asked.ID, ConversationID: conversationID}}, queue.tasks,
		"only the replies left unread ask for an email")
}
