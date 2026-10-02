package usecase

import (
	"time"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainpayment "github.com/boms/backend/internal/domain/payment"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/dto"
)

func mapOrderPaymentToDTO(p *domainpayment.Payment) *dto.OrderPaymentResponse {
	if p == nil {
		return nil
	}
	return &dto.OrderPaymentResponse{
		Provider:          string(p.Provider),
		Status:            string(p.Status),
		CapturedAt:        p.CapturedAt,
		RefundRequestedAt: p.RefundRequestedAt,
		RefundedAt:        p.RefundedAt,
	}
}

func mapFulfillmentToDTO(items domainorder.Fulfillment) dto.FulfillmentResponse {
	resp := dto.FulfillmentResponse{HasKitchenItems: items.Kitchen, LeadMinutes: int(items.Lead / time.Minute)}
	if items.SoldOutOn != nil {
		day := items.SoldOutOn.Format(domainstore.DayLayout)
		resp.SoldOutOn = &day
	}
	return resp
}

// reasonToDTO is a cancellation's reason, null for any other move.
func reasonToDTO(reason string) *string {
	if reason == "" {
		return nil
	}
	return &reason
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
			Customization:  mapCustomizationToDTO(item.Customization),
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

func mapCustomizationToDTO(c *domainorder.Customization) *dto.CustomizationResponse {
	if c == nil {
		return nil
	}
	options := make([]dto.ChosenOptionResponse, 0, len(c.Options))
	for _, option := range c.Options {
		options = append(options, dto.ChosenOptionResponse{
			Group:           string(option.Group),
			Label:           option.Label,
			PriceDeltaCents: option.PriceDeltaCents,
		})
	}
	return &dto.CustomizationResponse{Options: options, Message: c.Message, ReferenceImageURL: c.ReferenceImageURL}
}

func mapOrderTimelineToDTO(events []domainorder.StatusEvent) []dto.OrderTimelineEntryResponse {
	out := make([]dto.OrderTimelineEntryResponse, 0, len(events))
	for _, event := range events {
		out = append(out, dto.OrderTimelineEntryResponse{Status: string(event.To), Reason: reasonToDTO(event.Reason), At: event.At})
	}
	return out
}

func mapStaffOrderTimelineToDTO(events []domainorder.StatusEvent) []dto.StaffOrderTimelineEntryResponse {
	out := make([]dto.StaffOrderTimelineEntryResponse, 0, len(events))
	for _, event := range events {
		entry := dto.StaffOrderTimelineEntryResponse{Status: string(event.To), Reason: reasonToDTO(event.Reason), At: event.At}
		if event.ActorRole != "" {
			role := string(event.ActorRole)
			entry.ActorRole = &role
		}
		out = append(out, entry)
	}
	return out
}

// toOrderIncidentResponses maps an order's incidents, an empty list for none.
func toOrderIncidentResponses(incidents []domainorder.Incident) []dto.OrderIncidentResponse {
	out := make([]dto.OrderIncidentResponse, 0, len(incidents))
	for _, incident := range incidents {
		out = append(out, toOrderIncidentResponse(incident))
	}
	return out
}

func toOrderIncidentResponse(incident domainorder.Incident) dto.OrderIncidentResponse {
	return dto.OrderIncidentResponse{
		ID:        incident.ID.String(),
		Type:      string(incident.Type),
		Source:    string(incident.Type.Source()),
		Note:      incident.Note,
		CreatedAt: incident.CreatedAt,
	}
}
