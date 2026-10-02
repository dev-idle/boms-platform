package dto

import "time"

// SendPromotionRequest is a promotion as a manager wrote it; the usecase
// checks its subject and message.
type SendPromotionRequest struct {
	Subject string `json:"subject" validate:"required"`
	Body    string `json:"body" validate:"required"`
}

// PromotionResponse is a promotion as managers list it. RecipientCount is
// null while the worker is still queuing its emails.
type PromotionResponse struct {
	ID             string    `json:"id"`
	Subject        string    `json:"subject"`
	Status         string    `json:"status"`
	RecipientCount *int32    `json:"recipient_count"`
	CreatedAt      time.Time `json:"created_at"`
	SenderName     *string   `json:"sender_name"`
}

// PromotionAudienceResponse is how many customers a promotion sent now would
// go to.
type PromotionAudienceResponse struct {
	Recipients int64 `json:"recipients"`
}

// UnsubscribeRequest is the token of a promotion email's unsubscribe link.
type UnsubscribeRequest struct {
	Token string `json:"token" validate:"required,max=128"`
}
