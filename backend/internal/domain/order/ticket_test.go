package order

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	domaincategory "github.com/boms/backend/internal/domain/category"
)

var (
	kitchen = domaincategory.StationKitchen
	counter = domaincategory.StationCounter
)

func TestDecompose(t *testing.T) {
	t.Parallel()
	cakeLine, pastryLine, comboLine := uuid.New(), uuid.New(), uuid.New()
	cake, croissant, tart := uuid.New(), uuid.New(), uuid.New()
	line := func(orderItem, product uuid.UUID, name string, station domaincategory.Station, quantity int32) TicketLine {
		return TicketLine{OrderItemID: orderItem, ProductID: product, Name: name, Station: station, Quantity: quantity}
	}
	item := func(orderItem, product uuid.UUID, name string, quantity int32) TicketItem {
		return TicketItem{OrderItemID: orderItem, ProductID: product, Name: name, Quantity: quantity}
	}

	cases := map[string]struct {
		lines []TicketLine
		want  []TicketDraft
	}{
		"kitchen_only": {
			lines: []TicketLine{line(cakeLine, cake, "Cake", kitchen, 1)},
			want:  []TicketDraft{{Station: kitchen, Items: []TicketItem{item(cakeLine, cake, "Cake", 1)}}},
		},
		"counter_only": {
			lines: []TicketLine{line(pastryLine, croissant, "Croissant", counter, 3)},
			want:  []TicketDraft{{Station: counter, Items: []TicketItem{item(pastryLine, croissant, "Croissant", 3)}}},
		},
		"mixed_order_goes_to_both_stations_kitchen_first": {
			lines: []TicketLine{line(pastryLine, croissant, "Croissant", counter, 3), line(cakeLine, cake, "Cake", kitchen, 1)},
			want: []TicketDraft{
				{Station: kitchen, Items: []TicketItem{item(cakeLine, cake, "Cake", 1)}},
				{Station: counter, Items: []TicketItem{item(pastryLine, croissant, "Croissant", 3)}},
			},
		},
		"combo_spanning_stations_lands_on_both": {
			lines: []TicketLine{
				line(comboLine, cake, "Cake", kitchen, 2),
				line(comboLine, croissant, "Croissant", counter, 4),
				line(pastryLine, croissant, "Croissant", counter, 1),
			},
			want: []TicketDraft{
				{Station: kitchen, Items: []TicketItem{item(comboLine, cake, "Cake", 2)}},
				{Station: counter, Items: []TicketItem{
					item(comboLine, croissant, "Croissant", 4),
					item(pastryLine, croissant, "Croissant", 1),
				}},
			},
		},
		"combo_within_one_station_stays_on_one_ticket": {
			lines: []TicketLine{line(comboLine, croissant, "Croissant", counter, 2), line(comboLine, tart, "Tart", counter, 2)},
			want: []TicketDraft{{Station: counter, Items: []TicketItem{
				item(comboLine, croissant, "Croissant", 2),
				item(comboLine, tart, "Tart", 2),
			}}},
		},
		"no_lines_no_tickets": {lines: nil, want: []TicketDraft{}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, Decompose(tc.lines))
		})
	}
}

func TestCanAdvanceTicket(t *testing.T) {
	t.Parallel()
	statuses := []TicketStatus{TicketQueued, TicketInProgress, TicketReady, TicketCancelled}
	allowed := map[[2]TicketStatus]bool{
		{TicketQueued, TicketInProgress}: true,
		{TicketInProgress, TicketReady}:  true,
	}
	for _, from := range statuses {
		for _, to := range statuses {
			assert.Equal(t, allowed[[2]TicketStatus{from, to}], CanAdvanceTicket(from, to), "%s -> %s", from, to)
		}
	}
}

// The case table backing synchronized completion: an order is ready exactly
// when every ticket it still has is ready, and in production from the first
// start until then.
func TestDeriveStatus(t *testing.T) {
	t.Parallel()
	cases := []struct {
		current Status
		tickets []TicketStatus
		want    Status
	}{
		{StatusConfirmed, []TicketStatus{TicketQueued}, StatusConfirmed},
		{StatusConfirmed, []TicketStatus{TicketQueued, TicketQueued}, StatusConfirmed},
		{StatusConfirmed, []TicketStatus{TicketInProgress}, StatusInProduction},
		{StatusConfirmed, []TicketStatus{TicketInProgress, TicketQueued}, StatusInProduction},
		{StatusInProduction, []TicketStatus{TicketReady, TicketQueued}, StatusInProduction},
		{StatusInProduction, []TicketStatus{TicketReady, TicketInProgress}, StatusInProduction},
		{StatusInProduction, []TicketStatus{TicketReady}, StatusReady},
		{StatusInProduction, []TicketStatus{TicketReady, TicketReady}, StatusReady},
		{StatusInProduction, []TicketStatus{TicketReady, TicketCancelled}, StatusReady},
		{StatusInProduction, []TicketStatus{TicketCancelled, TicketCancelled}, StatusInProduction},
		{StatusPending, []TicketStatus{TicketInProgress}, StatusPending},
		{StatusReady, []TicketStatus{TicketReady, TicketReady}, StatusReady},
		{StatusCancelled, []TicketStatus{TicketCancelled}, StatusCancelled},
		{StatusFulfilled, []TicketStatus{TicketReady}, StatusFulfilled},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, DeriveStatus(tc.current, tc.tickets), "%s with %v", tc.current, tc.tickets)
	}
}
