package usecase

import (
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
)

// mapProductionTimeToDTO leaves the average out when no order became ready.
func mapProductionTimeToDTO(production port.ProductionTime) dto.ProductionTimeResponse {
	out := dto.ProductionTimeResponse{Orders: production.Orders}
	if production.Orders > 0 {
		out.AverageMinutes = &production.AverageMinutes
	}
	return out
}
