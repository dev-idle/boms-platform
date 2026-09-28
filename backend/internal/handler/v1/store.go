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

// PickupRules handles GET /api/v1/store/pickup-rules for anyone, signed in or not.
func (h *StoreHandler) PickupRules(c fiber.Ctx) error {
	out, err := h.usecase.PickupRules(c.Context())
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
