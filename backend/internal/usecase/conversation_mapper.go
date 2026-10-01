package usecase

import (
	domainconversation "github.com/boms/backend/internal/domain/conversation"
	"github.com/boms/backend/internal/dto"
)

// toMessageResponses maps messages; withNames shows who at the counter wrote
// each staff message, which only the counter sees.
func toMessageResponses(messages []domainconversation.Message, withNames bool) []dto.MessageResponse {
	out := make([]dto.MessageResponse, 0, len(messages))
	for _, message := range messages {
		out = append(out, toMessageResponse(message, withNames))
	}
	return out
}

func toMessageResponse(message domainconversation.Message, withName bool) dto.MessageResponse {
	resp := dto.MessageResponse{
		ID:        message.ID.String(),
		From:      string(message.AuthorRole),
		Body:      message.Body,
		CreatedAt: message.CreatedAt,
	}
	if withName {
		resp.AuthorName = message.AuthorName
	}
	return resp
}
