package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainconversation "github.com/boms/backend/internal/domain/conversation"
	domainuser "github.com/boms/backend/internal/domain/user"
)

// ConversationRepository keeps the conversations about orders, one per order.
type ConversationRepository interface {
	// OpenForCustomerMessage starts, or opens again, the order's conversation
	// for a message from its customer, and returns the conversation's id.
	OpenForCustomerMessage(ctx context.Context, orderID uuid.UUID) (uuid.UUID, error)
	// OpenForStaffMessage does the same for a message from staffID, who takes
	// the conversation over.
	OpenForStaffMessage(ctx context.Context, orderID, staffID uuid.UUID) (uuid.UUID, error)
	CreateMessage(
		ctx context.Context,
		conversationID, authorID uuid.UUID,
		authorRole domainuser.Role,
		body string,
	) (*domainconversation.Message, error)
	// GetByOrderID returns apperrors.ErrNotFound before the order's first message.
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domainconversation.Conversation, error)
	// ListMessages pages the order's messages newest first, from the latest or
	// from before the message beforeID.
	ListMessages(ctx context.Context, orderID uuid.UUID, beforeID *uuid.UUID, limit int32) ([]domainconversation.Message, error)
	// MarkReadByCustomer and MarkReadByStaff report whether there was anything
	// left to read.
	MarkReadByCustomer(ctx context.Context, orderID uuid.UUID) (bool, error)
	MarkReadByStaff(ctx context.Context, orderID uuid.UUID) (bool, error)
	// SetStatus reports whether the conversation was in another status.
	SetStatus(ctx context.Context, orderID uuid.UUID, status domainconversation.Status) (bool, error)
	StaffList(ctx context.Context, params StaffListConversationsParams) ([]StaffConversation, error)
	StaffListCount(ctx context.Context, status *domainconversation.Status) (int64, error)
	StaffCounts(ctx context.Context) (StaffConversationCounts, error)
	// UnreadByCustomer returns, for each of the orders that has any, how many
	// messages its customer has not read.
	UnreadByCustomer(ctx context.Context, orderIDs []uuid.UUID) (map[uuid.UUID]int32, error)
	// ListMessagesByOrderIDs returns every message on the orders, oldest first.
	ListMessagesByOrderIDs(ctx context.Context, orderIDs []uuid.UUID) (map[uuid.UUID][]domainconversation.Message, error)
	// EraseForCustomer erases every message on the customer's orders and
	// closes their conversations.
	EraseForCustomer(ctx context.Context, userID uuid.UUID) error
}

// StaffListConversationsParams pages the counter's inbox. Status nil lists
// every conversation.
type StaffListConversationsParams struct {
	Status *domainconversation.Status
	Limit  int32
	Offset int32
}

// StaffConversation is one inbox row: the conversation, its order and
// customer, and the start of its last message.
type StaffConversation struct {
	OrderID       uuid.UUID
	OrderCode     string
	Status        domainconversation.Status
	UnreadByStaff int32
	LastMessageAt time.Time
	CustomerName  *string
	CustomerEmail string
	LastBody      string
}

// StaffConversationCounts is how many conversations are open, and how many
// hold messages the counter has not read.
type StaffConversationCounts struct {
	Open   int64
	Unread int64
}
