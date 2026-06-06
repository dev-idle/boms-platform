package usecase

import (
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/dto"
)

func mapOrderItemsToDTO(items []domainorder.Item) []dto.OrderItemResponse {
	out := make([]dto.OrderItemResponse, 0, len(items))
	for _, item := range items {
		row := dto.OrderItemResponse{
			ID:             item.ID.String(),
			LineType:       string(item.LineType),
			Name:           item.Name,
			Slug:           item.Slug,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			LineTotalCents: item.LineTotalCents,
		}
		if item.ProductID != nil {
			id := item.ProductID.String()
			row.ProductID = &id
		}
		if item.ComboID != nil {
			id := item.ComboID.String()
			row.ComboID = &id
		}
		out = append(out, row)
	}
	return out
}
