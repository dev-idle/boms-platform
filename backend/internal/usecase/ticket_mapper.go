package usecase

import (
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
)

func mapTicketItemsToDTO(items []domainorder.TicketItem) []dto.TicketItemResponse {
	out := make([]dto.TicketItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.TicketItemResponse{Name: item.Name, Quantity: item.Quantity})
	}
	return out
}

func mapTicketSummariesToDTO(tickets []domainorder.Ticket) []dto.TicketSummaryResponse {
	out := make([]dto.TicketSummaryResponse, 0, len(tickets))
	for _, ticket := range tickets {
		out = append(out, dto.TicketSummaryResponse{Station: string(ticket.Station), Status: string(ticket.Status)})
	}
	return out
}

func mapOrderTicketsToDTO(tickets []domainorder.Ticket) []dto.OrderTicketResponse {
	out := make([]dto.OrderTicketResponse, 0, len(tickets))
	for _, ticket := range tickets {
		out = append(out, dto.OrderTicketResponse{
			ID:      ticket.ID.String(),
			Station: string(ticket.Station),
			Status:  string(ticket.Status),
			Items:   mapTicketItemsToDTO(ticket.Items),
		})
	}
	return out
}

func toStationTicketResponse(row port.StationTicket) dto.StationTicketResponse {
	return dto.StationTicketResponse{
		ID:          row.Ticket.ID.String(),
		OrderID:     row.Ticket.OrderID.String(),
		OrderCode:   row.OrderCode,
		OrderStatus: string(row.OrderStatus),
		Station:     string(row.Ticket.Station),
		Status:      string(row.Ticket.Status),
		PickupAt:    row.PickupAt,
		Customer:    dto.TicketCustomerResponse{DisplayName: row.CustomerDisplayName},
		Items:       mapTicketItemsToDTO(row.Ticket.Items),
		CreatedAt:   row.Ticket.CreatedAt,
	}
}

func toTicketChangeResponse(change ticketChange) *dto.TicketChangeResponse {
	return &dto.TicketChangeResponse{
		ID:          change.after.ID.String(),
		OrderID:     change.after.OrderID.String(),
		Station:     string(change.after.Station),
		Status:      string(change.after.Status),
		OrderStatus: string(change.order.Status),
	}
}
