package v1

import (
	"github.com/gofiber/fiber/v3"

	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/usecase"
)

type StoreHandler struct {
	usecase *usecase.StoreUsecase
}

func NewStoreHandler(uc *usecase.StoreUsecase) *StoreHandler {
	return &StoreHandler{usecase: uc}
}

// PickupSlots handles GET /api/v1/store/pickup-slots?date=YYYY-MM-DD for anyone.
func (h *StoreHandler) PickupSlots(c fiber.Ctx) error {
	out, err := h.usecase.PickupSlots(c.Context(), c.Query("date"))
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// PickupRules handles GET /api/v1/store/pickup-rules for anyone, signed in or not.
func (h *StoreHandler) PickupRules(c fiber.Ctx) error {
	out, err := h.usecase.PickupRules(c.Context())
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
