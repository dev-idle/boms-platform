package v1

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"

	domainassistant "github.com/boms/backend/internal/domain/assistant"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
)

// AssistantHandler streams the assistant's answers.
type AssistantHandler struct {
	usecase *usecase.AssistantUsecase
	timeout time.Duration
	logger  *zap.Logger
}

// NewAssistantHandler bounds every answer by timeout, from the moment it
// starts streaming.
func NewAssistantHandler(uc *usecase.AssistantUsecase, timeout time.Duration, logger *zap.Logger) *AssistantHandler {
	return &AssistantHandler{usecase: uc, timeout: timeout, logger: logger}
}

// answerFailure is the error event for an answer that could not be given.
func answerFailure(err error) *apperrors.AppError {
	switch {
	case errors.Is(err, domainassistant.ErrBusy):
		return apperrors.ErrAssistantBusy
	case errors.Is(err, domainassistant.ErrUnavailable):
		return apperrors.ErrAssistantUnavailable
	default:
		return apperrors.ErrInternal
	}
}

// Ask answers a question as server-sent events: "session" first, "delta" as
// the model writes, then "answer" — or "error" when no answer can be given.
// A question that cannot be asked is refused before the stream starts, with
// the usual error envelope.
func (h *AssistantHandler) Ask(c fiber.Ctx) error {
	requestID := response.EnsureRequestID(c)
	var req dto.AskAssistantRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	var userID *uuid.UUID
	if id, ok := middleware.GetUserID(c); ok {
		userID = &id
	}
	question, err := h.usecase.Ask(c.Context(), userID, middleware.ClientIP(c), req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	c.Set(fiber.HeaderContentType, "text/event-stream")
	// no-transform keeps a proxy from compressing, and so buffering, the stream.
	c.Set(fiber.HeaderCacheControl, "no-store, no-transform")
	// The handler returns before the stream is written, so the stream cannot
	// use the request's context: it gets its own, ended by the timeout or by a
	// write the visitor is no longer there to read. Past the timeout the
	// visitor is still there, so the closing event is still written.
	return c.SendStreamWriter(func(w *bufio.Writer) {
		ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
		defer cancel()
		gone := false
		send := func(event string, data any) {
			if gone {
				return
			}
			payload, err := json.Marshal(data)
			if err == nil {
				_, err = w.WriteString("event: " + event + "\ndata: " + string(payload) + "\n\n")
			}
			if err == nil {
				err = w.Flush()
			}
			if err != nil {
				gone = true
				cancel()
			}
		}
		send("session", dto.AssistantSessionEvent{SessionID: question.SessionID()})
		answer, err := question.Answer(ctx, func(text string) {
			send("delta", dto.AssistantDeltaEvent{Text: text})
		})
		if err != nil && gone {
			return
		}
		if err != nil {
			appErr := answerFailure(err)
			if appErr == apperrors.ErrAssistantBusy {
				// The load the cap is there for, not a failure.
				h.logger.Warn("assistant_busy", zap.String("request_id", requestID))
			} else {
				h.logger.Error("assistant answer failed", zap.String("request_id", requestID), zap.Error(err))
			}
			code, message, _ := appErr.ToErrorBody()
			send("error", dto.AssistantErrorEvent{Code: code, Message: message})
			return
		}
		send("answer", answer)
	})
}
