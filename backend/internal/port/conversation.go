package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainconversation "github.com/boms/backend/internal/domain/conversation"
	domainuser "github.com/boms/backend/internal/domain/user"
)

// ConversationThread names one conversation: the customer's about one of
// their orders, or their general one when OrderID is nil.
type ConversationThread struct {
	CustomerID uuid.UUID
	OrderID    *uuid.UUID
}

// ConversationRepository keeps the conversations between customers and the
// counter: one per order, and one general conversation per customer.
type ConversationRepository interface {
	// OpenForCustomerMessage starts, or opens again, the conversation for a
	// message from its customer, and returns the conversation's id.
	OpenForCustomerMessage(ctx context.Context, thread ConversationThread) (uuid.UUID, error)
	// OpenForStaffMessage does the same for a message from staffID, who takes
	// the conversation over.
	OpenForStaffMessage(ctx context.Context, thread ConversationThread, staffID uuid.UUID) (uuid.UUID, error)
	// CreateMessage writes a message; viaAssistant marks a customer's questions
	// the assistant passed on.
	CreateMessage(
		ctx context.Context,
		conversationID, authorID uuid.UUID,
		authorRole domainuser.Role,
		body string,
		viaAssistant bool,
	) (*domainconversation.Message, error)
	// Get returns apperrors.ErrNotFound before the conversation's first message.
	Get(ctx context.Context, thread ConversationThread) (*domainconversation.Conversation, error)
	// GeneralCustomer returns who the customer's general conversation is with,
	// apperrors.ErrNotFound when there is none or their account is closed.
	GeneralCustomer(ctx context.Context, customerID uuid.UUID) (*ConversationCustomer, error)
	// ListMessages pages the conversation's messages newest first, from the
	// latest or from before the message beforeID.
	ListMessages(ctx context.Context, thread ConversationThread, beforeID *uuid.UUID, limit int32) ([]domainconversation.Message, error)
	// MarkReadByCustomer and MarkReadByStaff report whether there was anything
	// left to read.
	MarkReadByCustomer(ctx context.Context, thread ConversationThread) (bool, error)
	MarkReadByStaff(ctx context.Context, thread ConversationThread) (bool, error)
	// SetStatus reports whether the conversation was in another status.
	SetStatus(ctx context.Context, thread ConversationThread, status domainconversation.Status) (bool, error)
	StaffList(ctx context.Context, params StaffListConversationsParams) ([]StaffConversation, error)
	StaffListCount(ctx context.Context, status *domainconversation.Status) (int64, error)
	StaffCounts(ctx context.Context) (StaffConversationCounts, error)
	// UnreadByCustomer returns, for each of the orders that has any, how many
	// messages its customer has not read.
	UnreadByCustomer(ctx context.Context, orderIDs []uuid.UUID) (map[uuid.UUID]int32, error)
	// ListMessagesByOrderIDs returns every message on the orders, oldest first.
	ListMessagesByOrderIDs(ctx context.Context, orderIDs []uuid.UUID) (map[uuid.UUID][]domainconversation.Message, error)
	// ListGeneralMessages returns every message of the customer's general
	// conversation, oldest first.
	ListGeneralMessages(ctx context.Context, customerID uuid.UUID) ([]domainconversation.Message, error)
	// EraseForCustomer erases every message of the customer's conversations
	// and closes them.
	EraseForCustomer(ctx context.Context, userID uuid.UUID) error
	// ListUnreadRepliesDue returns up to limit conversations whose customer
	// has left the counter's replies unread for unreadFor and has not been
	// emailed about them, the longest waiting first.
	ListUnreadRepliesDue(ctx context.Context, unreadFor time.Duration, limit int32) ([]uuid.UUID, error)
	// MarkRepliesEmailed records that the customer is emailed about them;
	// false when they read them meanwhile or are emailed already.
	MarkRepliesEmailed(ctx context.Context, conversationID uuid.UUID) (bool, error)
	// UnreadReplies returns what the email about them says;
	// apperrors.ErrNotFound when the account is closed or its address not
	// confirmed.
	UnreadReplies(ctx context.Context, conversationID uuid.UUID) (*UnreadReplies, error)
}

// UnreadReplies is the counter's replies a customer left unread on a
// conversation — about an order, or their general one — and who they are.
type UnreadReplies struct {
	OrderID   *uuid.UUID
	OrderCode *string
	Count     int32
	Email     string
	Name      *string
}

// ConversationCustomer is who a general conversation is with.
type ConversationCustomer struct {
	Name  *string
	Email string
}

// StaffListConversationsParams pages the counter's inbox. Status nil lists
// every conversation.
type StaffListConversationsParams struct {
	Status *domainconversation.Status
	Limit  int32
	Offset int32
}

// StaffConversation is one inbox row: the conversation, its customer, its
// order (none for a general conversation) and the start of its last message.
type StaffConversation struct {
	CustomerID    uuid.UUID
	OrderID       *uuid.UUID
	OrderCode     *string
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
