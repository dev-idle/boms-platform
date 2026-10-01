package v1

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/usecase"
)

// DataExportHandler serves GET /me/export: a person's own data, for them to
// download.
type DataExportHandler struct {
	usecase *usecase.DataExportUsecase
}

func NewDataExportHandler(uc *usecase.DataExportUsecase) *DataExportHandler {
	return &DataExportHandler{usecase: uc}
}

func (h *DataExportHandler) Export(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	out, err := h.usecase.Export(c.Context(), userID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	// Personal data: no shared or browser cache may keep a copy.
	c.Set(fiber.HeaderCacheControl, "no-store")
	return response.OK(c, dto.DataExportResponse{
		ExportedAt:      time.Now().UTC(),
		Account:         toMeResponse(out.User, out.Profile),
		TermsAcceptance: out.Terms,
		Sessions:        out.Sessions,
		AccountActivity: out.Activity,
		Cart:            out.Cart,
		SavedProducts:   out.Saved,
		Orders:          out.Orders,
	})
}
