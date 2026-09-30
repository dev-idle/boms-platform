package usecase

import (
	domainorder "github.com/boms/backend/internal/domain/order"
	domainpayment "github.com/boms/backend/internal/domain/payment"
	"github.com/boms/backend/internal/dto"
)

func mapOrderPaymentToDTO(p *domainpayment.Payment) *dto.OrderPaymentResponse {
	if p == nil {
		return nil
	}
	return &dto.OrderPaymentResponse{Provider: string(p.Provider), Status: string(p.Status), CapturedAt: p.CapturedAt}
}

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

func mapOrderTimelineToDTO(events []domainorder.StatusEvent) []dto.OrderTimelineEntryResponse {
	out := make([]dto.OrderTimelineEntryResponse, 0, len(events))
	for _, event := range events {
		out = append(out, dto.OrderTimelineEntryResponse{Status: string(event.To), At: event.At})
	}
	return out
}

func mapStaffOrderTimelineToDTO(events []domainorder.StatusEvent) []dto.StaffOrderTimelineEntryResponse {
	out := make([]dto.StaffOrderTimelineEntryResponse, 0, len(events))
	for _, event := range events {
		entry := dto.StaffOrderTimelineEntryResponse{Status: string(event.To), At: event.At}
		if event.ActorRole != "" {
			role := string(event.ActorRole)
			entry.ActorRole = &role
		}
		out = append(out, entry)
	}
	return out
}
