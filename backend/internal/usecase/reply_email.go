package usecase

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// ReplyEmailUsecase emails a customer that the counter replied and they have
// not read it. It runs in the worker and reads the conversation as it sends,
// so replies read since the email was asked for send nothing.
type ReplyEmailUsecase struct {
	conversations port.ConversationRepository
	composer      port.ReplyEmailComposer
	mailer        port.Mailer
	log           *zap.Logger
}

func NewReplyEmailUsecase(
	conversations port.ConversationRepository,
	composer port.ReplyEmailComposer,
	mailer port.Mailer,
	log *zap.Logger,
) *ReplyEmailUsecase {
	return &ReplyEmailUsecase{conversations: conversations, composer: composer, mailer: mailer, log: log}
}

// Send emails the customer of task's conversation. It skips an account closed
// or with an address not confirmed, and replies read meanwhile. Errors
// wrapping port.ErrEmailUndeliverable are final; any other may pass on a retry.
func (u *ReplyEmailUsecase) Send(ctx context.Context, task port.ReplyEmailTask) error {
	replies, err := u.conversations.UnreadReplies(ctx, task.ConversationID)
	if errors.Is(err, apperrors.ErrNotFound) {
		u.skip(task, "no_open_confirmed_account")
		return nil
	}
	if err != nil {
		return err
	}
	if replies.Count == 0 {
		u.skip(task, "read_already")
		return nil
	}
	msg := port.ReplyEmail{To: replies.Email, OrderID: replies.OrderID, Count: replies.Count}
	if replies.Name != nil {
		msg.ToName = *replies.Name
	}
	if replies.OrderCode != nil {
		msg.OrderCode = *replies.OrderCode
	}
	email, err := u.composer.ComposeReplyEmail(msg)
	if err != nil {
		return fmt.Errorf("%w: compose reply email: %w", port.ErrEmailUndeliverable, err)
	}
	if err := u.mailer.Send(ctx, email); err != nil {
		return err
	}
	u.log.Info("reply_email_sent", zap.String("conversation_id", task.ConversationID.String()))
	return nil
}

func (u *ReplyEmailUsecase) skip(task port.ReplyEmailTask, reason string) {
	u.log.Info("reply_email_skipped",
		zap.String("conversation_id", task.ConversationID.String()),
		zap.String("reason", reason),
	)
}
