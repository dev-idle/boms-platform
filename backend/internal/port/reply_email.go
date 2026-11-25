package port

import (
	"context"

	"github.com/google/uuid"
)

// ReplyEmailTask asks for the email about the counter's replies a customer
// left unread on a conversation.
type ReplyEmailTask struct {
	EventID        uuid.UUID
	ConversationID uuid.UUID
}

// ReplyEmailQueue holds reply emails until a worker sends them, one task per
// event.
type ReplyEmailQueue interface {
	EnqueueReplyEmail(ctx context.Context, task ReplyEmailTask) error
}

// ReplyEmail is what the email says: the replies waiting, on an order or in
// the customer's general conversation when OrderID is nil.
type ReplyEmail struct {
	To        string
	ToName    string
	OrderID   *uuid.UUID
	OrderCode string
	Count     int32
}

// ReplyEmailComposer writes the reply email.
type ReplyEmailComposer interface {
	ComposeReplyEmail(msg ReplyEmail) (Email, error)
}
