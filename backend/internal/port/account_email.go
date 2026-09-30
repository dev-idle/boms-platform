package port

import (
	"context"

	"github.com/google/uuid"

	domainaccount "github.com/boms/backend/internal/domain/account"
)

// AccountEmailTask asks for one account email — a confirmation or a reset
// link. It carries ids only: the link is made when the email is sent, so no
// queue or log ever holds one.
type AccountEmailTask struct {
	EventID uuid.UUID
	UserID  uuid.UUID
	Purpose domainaccount.Purpose
}

// AccountEmailQueue holds account emails until a worker sends them, one task
// per event.
type AccountEmailQueue interface {
	EnqueueAccountEmail(ctx context.Context, task AccountEmailTask) error
}

// AccountEmail is what an account email says: its link carries Token.
type AccountEmail struct {
	Purpose domainaccount.Purpose
	To      string
	Token   string
}

// AccountEmailComposer writes an account email.
type AccountEmailComposer interface {
	ComposeAccountEmail(msg AccountEmail) (Email, error)
}
