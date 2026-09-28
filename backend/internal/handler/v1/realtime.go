package v1

import (
	"github.com/gofiber/fiber/v3"

	domainrealtime "github.com/boms/backend/internal/domain/realtime"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/usecase"
)

type RealtimeHandler struct {
	usecase *usecase.RealtimeUsecase
}

func NewRealtimeHandler(uc *usecase.RealtimeUsecase) *RealtimeHandler {
	return &RealtimeHandler{usecase: uc}
}

// IssueTicket handles POST /api/v1/realtime/tickets for any signed-in session.
func (h *RealtimeHandler) IssueTicket(c fiber.Ctx) error {
	userID, hasUser := middleware.GetUserID(c)
	role, hasRole := middleware.GetRole(c)
	sessionID, hasSession := middleware.GetSessionID(c)
	if !hasUser || !hasRole || !hasSession {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	resp, err := h.usecase.IssueTicket(c.Context(), domainrealtime.Ticket{
		UserID:    userID,
		Role:      role,
		SessionID: sessionID,
	})
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	// A ticket is a short-lived credential: never cached anywhere on the way.
	noStore(c)
	return response.Created(c, resp)
}
