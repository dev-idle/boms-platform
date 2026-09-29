package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

type TicketRepository struct {
	queries *sqlcgen.Queries
}

func NewTicketRepository(pool *Pool) *TicketRepository {
	return &TicketRepository{queries: pool.Queries()}
}

func (r *TicketRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *TicketRepository) ListLines(ctx context.Context, orderID uuid.UUID) ([]domainorder.TicketLine, error) {
	rows, err := r.q(ctx).ListOrderTicketLines(ctx, orderID)
	if err != nil {
		return nil, mapRepoError(err, "list order ticket lines")
	}
	out := make([]domainorder.TicketLine, 0, len(rows))
	for _, row := range rows {
		out = append(out, domainorder.TicketLine{
			OrderItemID: row.OrderItemID,
			ProductID:   row.ProductID,
			Name:        row.Name,
			Station:     domaincategory.Station(row.Station),
			Quantity:    row.Quantity,
		})
	}
	return out, nil
}

// ticketRecord and ticketItemRecord are elements of the JSON arrays the create
// queries expand with jsonb_to_recordset; field names match their column lists.
type ticketRecord struct {
	Station sqlcgen.Station `json:"station"`
}

type ticketItemRecord struct {
	TicketID    uuid.UUID `json:"ticket_id"`
	OrderItemID uuid.UUID `json:"order_item_id"`
	ProductID   uuid.UUID `json:"product_id"`
	Name        string    `json:"name"`
	Quantity    int32     `json:"quantity"`
}

// CreateForOrder inserts the tickets in one statement and their items in another.
func (r *TicketRepository) CreateForOrder(
	ctx context.Context,
	orderID uuid.UUID,
	drafts []domainorder.TicketDraft,
) ([]domainorder.Ticket, error) {
	if len(drafts) == 0 {
		return nil, apperrors.Errorf("create order tickets: an order makes at least one ticket")
	}
	tickets := make([]ticketRecord, 0, len(drafts))
	for _, draft := range drafts {
		tickets = append(tickets, ticketRecord{Station: sqlcgen.Station(draft.Station)})
	}
	ticketsJSON, err := json.Marshal(tickets)
	if err != nil {
		return nil, apperrors.Errorf("encode order tickets: %w", err)
	}
	rows, err := r.q(ctx).CreateOrderTickets(ctx, sqlcgen.CreateOrderTicketsParams{OrderID: orderID, Tickets: ticketsJSON})
	if err != nil {
		return nil, mapRepoError(err, "create order tickets")
	}
	if len(rows) != len(drafts) {
		return nil, apperrors.Errorf("create order tickets: created %d of %d", len(rows), len(drafts))
	}

	byStation := make(map[domaincategory.Station]domainorder.Ticket, len(rows))
	for _, row := range rows {
		byStation[domaincategory.Station(row.Station)] = mapTicket(row)
	}
	var items []ticketItemRecord
	out := make([]domainorder.Ticket, 0, len(drafts))
	for _, draft := range drafts {
		ticket := byStation[draft.Station]
		for _, item := range draft.Items {
			items = append(items, ticketItemRecord{
				TicketID:    ticket.ID,
				OrderItemID: item.OrderItemID,
				ProductID:   item.ProductID,
				Name:        item.Name,
				Quantity:    item.Quantity,
			})
		}
		ticket.Items = draft.Items
		out = append(out, ticket)
	}
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return nil, apperrors.Errorf("encode order ticket items: %w", err)
	}
	inserted, err := r.q(ctx).CreateOrderTicketItems(ctx, itemsJSON)
	if err != nil {
		return nil, mapRepoError(err, "create order ticket items")
	}
	if inserted != int64(len(items)) {
		return nil, apperrors.Errorf("create order ticket items: inserted %d of %d", inserted, len(items))
	}
	return out, nil
}

func (r *TicketRepository) Get(ctx context.Context, id uuid.UUID) (domainorder.Ticket, error) {
	row, err := r.q(ctx).GetOrderTicket(ctx, id)
	if err != nil {
		return domainorder.Ticket{}, mapRepoError(err, "get order ticket")
	}
	return mapTicket(row), nil
}

func (r *TicketRepository) ListByOrder(ctx context.Context, orderID uuid.UUID) ([]domainorder.Ticket, error) {
	rows, err := r.q(ctx).ListOrderTickets(ctx, orderID)
	if err != nil {
		return nil, mapRepoError(err, "list order tickets")
	}
	itemRows, err := r.q(ctx).ListOrderTicketItems(ctx, orderID)
	if err != nil {
		return nil, mapRepoError(err, "list order ticket items")
	}
	items := make(map[uuid.UUID][]domainorder.TicketItem, len(rows))
	for _, row := range itemRows {
		items[row.TicketID] = append(items[row.TicketID], domainorder.TicketItem{
			OrderItemID: row.OrderItemID,
			ProductID:   row.ProductID,
			Name:        row.Name,
			Quantity:    row.Quantity,
		})
	}
	out := make([]domainorder.Ticket, 0, len(rows))
	for _, row := range rows {
		ticket := mapTicket(row)
		ticket.Items = items[ticket.ID]
		out = append(out, ticket)
	}
	return out, nil
}

func (r *TicketRepository) UpdateStatus(ctx context.Context, params port.UpdateTicketStatusParams) (domainorder.Ticket, error) {
	row, err := r.q(ctx).UpdateOrderTicketStatus(ctx, sqlcgen.UpdateOrderTicketStatusParams{
		ID:         params.TicketID,
		Station:    sqlcgen.Station(params.Station),
		FromStatus: sqlcgen.TicketStatus(params.From),
		ToStatus:   sqlcgen.TicketStatus(params.To),
	})
	if err != nil {
		return domainorder.Ticket{}, mapRepoError(err, "update order ticket status")
	}
	return mapTicket(row), nil
}

func (r *TicketRepository) Move(ctx context.Context, id uuid.UUID, from, to domaincategory.Station) (domainorder.Ticket, error) {
	row, err := r.q(ctx).MoveOrderTicket(ctx, sqlcgen.MoveOrderTicketParams{
		ID:          id,
		FromStation: sqlcgen.Station(from),
		ToStation:   sqlcgen.Station(to),
	})
	if err != nil {
		return domainorder.Ticket{}, mapRepoError(err, "move order ticket")
	}
	return mapTicket(row), nil
}

func (r *TicketRepository) CancelForOrder(ctx context.Context, orderID uuid.UUID) ([]domainorder.Ticket, error) {
	rows, err := r.q(ctx).CancelOrderTickets(ctx, orderID)
	if err != nil {
		return nil, mapRepoError(err, "cancel order tickets")
	}
	out := make([]domainorder.Ticket, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapTicket(row))
	}
	return out, nil
}

// stationTicketItem is one element of the items array ListStationTickets builds.
type stationTicketItem struct {
	Name     string `json:"name"`
	Quantity int32  `json:"quantity"`
}

func (r *TicketRepository) ListByStation(ctx context.Context, params port.StationTicketsParams) ([]port.StationTicket, error) {
	rows, err := r.q(ctx).ListStationTickets(ctx, sqlcgen.ListStationTicketsParams{
		Station: sqlcgen.Station(params.Station),
		Status:  optionalTicketStatus(params.Status),
		Limit:   params.Limit,
		Offset:  params.Offset,
	})
	if err != nil {
		return nil, mapRepoError(err, "list station tickets")
	}
	out := make([]port.StationTicket, 0, len(rows))
	for _, row := range rows {
		var listed []stationTicketItem
		if err := json.Unmarshal(row.Items, &listed); err != nil {
			return nil, apperrors.Errorf("decode station ticket items: %w", err)
		}
		items := make([]domainorder.TicketItem, 0, len(listed))
		for _, item := range listed {
			items = append(items, domainorder.TicketItem{Name: item.Name, Quantity: item.Quantity})
		}
		out = append(out, port.StationTicket{
			Ticket: domainorder.Ticket{
				ID:        row.ID,
				OrderID:   row.OrderID,
				Station:   domaincategory.Station(row.Station),
				Status:    domainorder.TicketStatus(row.Status),
				Items:     items,
				CreatedAt: row.CreatedAt,
				UpdatedAt: row.UpdatedAt,
			},
			OrderCode:           row.OrderCode,
			OrderStatus:         mapOrderStatusFromSQL(row.OrderStatus),
			PickupAt:            row.PickupAt,
			CustomerDisplayName: row.CustomerDisplayName,
		})
	}
	return out, nil
}

func (r *TicketRepository) CountByStation(
	ctx context.Context,
	station domaincategory.Station,
	status *domainorder.TicketStatus,
) (int64, error) {
	count, err := r.q(ctx).CountStationTickets(ctx, sqlcgen.CountStationTicketsParams{
		Station: sqlcgen.Station(station),
		Status:  optionalTicketStatus(status),
	})
	if err != nil {
		return 0, mapRepoError(err, "count station tickets")
	}
	return count, nil
}

func optionalTicketStatus(status *domainorder.TicketStatus) *sqlcgen.TicketStatus {
	if status == nil {
		return nil
	}
	mapped := sqlcgen.TicketStatus(*status)
	return &mapped
}

func mapTicket(row sqlcgen.OrderTicket) domainorder.Ticket {
	return domainorder.Ticket{
		ID:        row.ID,
		OrderID:   row.OrderID,
		Station:   domaincategory.Station(row.Station),
		Status:    domainorder.TicketStatus(row.Status),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

var _ port.TicketRepository = (*TicketRepository)(nil)
