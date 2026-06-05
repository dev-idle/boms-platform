package usecase

import (
	domaincombo "github.com/boms/backend/internal/domain/combo"
	"github.com/boms/backend/internal/dto"
)

func toComboItemResponses(items []domaincombo.Item) []dto.ComboItemResponse {
	out := make([]dto.ComboItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.ComboItemResponse{
			ProductID:   item.ProductID.String(),
			ProductName: item.ProductName,
			ProductSlug: item.ProductSlug,
			Quantity:    item.Quantity,
			PriceCents:  item.PriceCents,
		})
	}
	return out
}
