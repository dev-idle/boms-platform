package dto

import "time"

// MessageResponse is one message of a conversation. From is "customer" or "staff";
// AuthorName, the staff writer's name, is shown to the counter only.
type MessageResponse struct {
	ID         string    `json:"id"`
	From       string    `json:"from"`
	AuthorName *string   `json:"author_name"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
}

// ThreadResponse is a page of a conversation's messages as its customer reads them,
// oldest first; HasMore says older ones exist. Unread is how many the
// customer has not read.
type ThreadResponse struct {
	Unread   int32             `json:"unread"`
	Messages []MessageResponse `json:"messages"`
	HasMore  bool              `json:"has_more"`
}

// StaffThreadResponse is a page of a conversation's messages as the counter
// reads them. Conversation is null before an order's first message. Customer
// is who a general conversation is with; null for an order's, whose order
// names them.
type StaffThreadResponse struct {
	Conversation *StaffConversationResponse   `json:"conversation"`
	Customer     *StaffThreadCustomerResponse `json:"customer"`
	Messages     []MessageResponse            `json:"messages"`
	HasMore      bool                         `json:"has_more"`
}

// StaffThreadCustomerResponse is who a general conversation is with.
type StaffThreadCustomerResponse struct {
	DisplayName *string `json:"display_name"`
	Email       string  `json:"email"`
}

// StaffConversationResponse is where a conversation stands at the counter.
type StaffConversationResponse struct {
	Status            string  `json:"status"`
	Unread            int32   `json:"unread"`
	AssignedStaffName *string `json:"assigned_staff_name"`
}

// StaffInboxConversationResponse is one conversation in the counter's inbox:
// about an order, or — order_id and order_code null — the customer's general
// conversation.
type StaffInboxConversationResponse struct {
	CustomerID    string    `json:"customer_id"`
	OrderID       *string   `json:"order_id"`
	OrderCode     *string   `json:"order_code"`
	Status        string    `json:"status"`
	Unread        int32     `json:"unread"`
	LastMessageAt time.Time `json:"last_message_at"`
	CustomerName  *string   `json:"customer_name"`
	CustomerEmail string    `json:"customer_email"`
	Preview       string    `json:"preview"`
}

// StaffConversationCountsResponse is how many conversations are open, and how
// many hold messages nobody at the counter has read.
type StaffConversationCountsResponse struct {
	Open   int64 `json:"open"`
	Unread int64 `json:"unread"`
}

// PostMessageRequest is a message as written; the usecase checks its text.
type PostMessageRequest struct {
	Body string `json:"body" validate:"required"`
}

// PatchConversationRequest resolves a conversation, or opens it again.
type PatchConversationRequest struct {
	Status string `json:"status" validate:"required,oneof=open closed"`
}
