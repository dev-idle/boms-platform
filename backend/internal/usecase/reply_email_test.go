package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

// waitingReplies answers with replies, err, or not found when both are nil;
// any other call panics.
type waitingReplies struct {
	port.ConversationRepository
	replies *port.UnreadReplies
	err     error
}

func (w waitingReplies) UnreadReplies(context.Context, uuid.UUID) (*port.UnreadReplies, error) {
	if w.err != nil {
		return nil, w.err
	}
	if w.replies == nil {
		return nil, apperrors.ErrNotFound
	}
	return w.replies, nil
}

// replyComposer keeps what it was asked to write, or refuses with err.
type replyComposer struct {
	got []port.ReplyEmail
	err error
}

func (c *replyComposer) ComposeReplyEmail(msg port.ReplyEmail) (port.Email, error) {
	c.got = append(c.got, msg)
	return port.Email{To: msg.To, Subject: "Our staff replied"}, c.err
}

// sentEmails keeps what it sent, or fails with err.
type sentEmails struct {
	sent []port.Email
	err  error
}

func (s *sentEmails) Send(_ context.Context, email port.Email) error {
	if s.err != nil {
		return s.err
	}
	s.sent = append(s.sent, email)
	return nil
}

func TestReplyEmailUsecase_Send(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	task := port.ReplyEmailTask{EventID: uuid.New(), ConversationID: uuid.New()}
	send := func(replies *port.UnreadReplies) (*replyComposer, *sentEmails, error) {
		composer, mailer := &replyComposer{}, &sentEmails{}
		err := usecase.NewReplyEmailUsecase(waitingReplies{replies: replies}, composer, mailer, zap.NewNop()).Send(ctx, task)
		return composer, mailer, err
	}
	waiting := &port.UnreadReplies{Count: 1, Email: "mai@example.com"}

	t.Run("tells_the_customer_about_the_replies_on_their_order", func(t *testing.T) {
		t.Parallel()
		orderID, code, name := uuid.New(), "CH-261125-004", "Mai"
		composer, mailer, err := send(&port.UnreadReplies{OrderID: &orderID, OrderCode: &code, Count: 2, Email: "mai@example.com", Name: &name})

		require.NoError(t, err)
		require.Len(t, mailer.sent, 1)
		assert.Equal(t, []port.ReplyEmail{{To: "mai@example.com", ToName: "Mai", OrderID: &orderID, OrderCode: code, Count: 2}}, composer.got)
	})

	t.Run("replies_read_meanwhile_send_nothing", func(t *testing.T) {
		t.Parallel()
		_, mailer, err := send(&port.UnreadReplies{Count: 0, Email: "mai@example.com"})

		require.NoError(t, err)
		assert.Empty(t, mailer.sent)
	})

	t.Run("a_closed_account_is_not_emailed", func(t *testing.T) {
		t.Parallel()
		_, mailer, err := send(nil)

		require.NoError(t, err, "nothing to retry")
		assert.Empty(t, mailer.sent)
	})

	t.Run("a_passing_failure_is_left_to_retry", func(t *testing.T) {
		t.Parallel()
		down := errors.New("connection refused")
		for name, uc := range map[string]*usecase.ReplyEmailUsecase{
			"database": usecase.NewReplyEmailUsecase(waitingReplies{err: down}, &replyComposer{}, &sentEmails{}, zap.NewNop()),
			"mailer":   usecase.NewReplyEmailUsecase(waitingReplies{replies: waiting}, &replyComposer{}, &sentEmails{err: down}, zap.NewNop()),
		} {
			err := uc.Send(ctx, task)

			require.ErrorIs(t, err, down, name)
			assert.NotErrorIs(t, err, port.ErrEmailUndeliverable, "%s: the queue tries it again", name)
		}
	})

	t.Run("an_email_no_retry_can_write_is_final", func(t *testing.T) {
		t.Parallel()
		mailer := &sentEmails{}
		composer := &replyComposer{err: errors.New("template broke")}

		err := usecase.NewReplyEmailUsecase(waitingReplies{replies: waiting}, composer, mailer, zap.NewNop()).Send(ctx, task)

		require.ErrorIs(t, err, port.ErrEmailUndeliverable)
		assert.Empty(t, mailer.sent)
	})
}
