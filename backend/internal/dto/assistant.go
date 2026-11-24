package dto

// AskAssistantRequest is a question to the assistant. SessionID continues a
// conversation the same visitor started; without it, a new one starts.
type AskAssistantRequest struct {
	SessionID *string `json:"session_id" validate:"omitempty,uuid"`
	Question  string  `json:"question" validate:"required"`
}

// AssistantSessionEvent is the first event of an answer: the session it
// belongs to, for the visitor's next question.
type AssistantSessionEvent struct {
	SessionID string `json:"session_id"`
}

// AssistantDeltaEvent is text of the answer as the model writes it.
type AssistantDeltaEvent struct {
	Text string `json:"text"`
}

// AssistantAnswerResponse is the last event of an answer: the answer as it is
// kept, which replaces the text streamed before it. OfferStaff suggests
// talking to staff instead. HandOff says the question went to the counter.
type AssistantAnswerResponse struct {
	Reply      string                    `json:"reply"`
	OfferStaff bool                      `json:"offer_staff"`
	HandOff    *AssistantHandOffResponse `json:"hand_off"`
}

// AssistantHandOffResponse is where the counter will answer: on the order it
// names, or in the customer's general conversation when the order is null.
type AssistantHandOffResponse struct {
	OrderID   *string `json:"order_id"`
	OrderCode *string `json:"order_code"`
}

// AssistantErrorEvent ends an answer that could not be given.
type AssistantErrorEvent struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
