package port

import (
	"context"
	"time"

	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
)

// ActiveOrders is how many orders the bakery still has to make or hand over,
// by status.
type ActiveOrders struct {
	Pending      int64
	Confirmed    int64
	InProduction int64
	Ready        int64
}

// StationLoad is how many tickets one station holds of the orders being made
// or waiting at the counter, by status.
type StationLoad struct {
	Station    domaincategory.Station
	Queued     int64
	InProgress int64
	Ready      int64
}

// DayPickups is how many orders the bakery took are due on a day: collected,
// still to collect, and not collected.
type DayPickups struct {
	Due       int64
	Collected int64
	ToCollect int64
	Missed    int64
}

// LateOrder is an order still to collect whose pickup slot has ended.
type LateOrder struct {
	Code     string
	Status   domainorder.Status
	PickupAt time.Time
}

// OperationsRepository reads where the bakery's work stands now.
type OperationsRepository interface {
	ActiveOrders(ctx context.Context) (ActiveOrders, error)
	// StationLoads leaves out a station with no ticket.
	StationLoads(ctx context.Context) ([]StationLoad, error)
	// DayPickups covers the pickups in [from, to).
	DayPickups(ctx context.Context, from, to time.Time) (DayPickups, error)
	// LateOrders returns up to limit orders late at now, the longest waiting
	// first, and how many are late in all.
	LateOrders(ctx context.Context, now time.Time, limit int32) ([]LateOrder, int64, error)
}
