package dto

import "time"

// OperationsResponse is where the bakery's work stands now: the orders still
// to make or hand over by status, each station's tickets, today's pickups,
// the orders late, and how long the orders that became ready today took.
type OperationsResponse struct {
	Orders     ActiveOrdersResponse   `json:"orders"`
	Stations   []StationLoadResponse  `json:"stations"`
	Pickups    DayPickupsResponse     `json:"pickups"`
	Late       LateOrdersResponse     `json:"late"`
	Production ProductionTimeResponse `json:"production"`
}

// ActiveOrdersResponse is how many orders stand in each status the bakery
// still works on.
type ActiveOrdersResponse struct {
	Pending      int64 `json:"pending"`
	Confirmed    int64 `json:"confirmed"`
	InProduction int64 `json:"in_production"`
	Ready        int64 `json:"ready"`
}

// StationLoadResponse is one station's tickets by status.
type StationLoadResponse struct {
	Station    string `json:"station"`
	Queued     int64  `json:"queued"`
	InProgress int64  `json:"in_progress"`
	Ready      int64  `json:"ready"`
}

// DayPickupsResponse is today's pickups: due, collected, still to collect,
// and not collected.
type DayPickupsResponse struct {
	Due       int64 `json:"due"`
	Collected int64 `json:"collected"`
	ToCollect int64 `json:"to_collect"`
	Missed    int64 `json:"missed"`
}

// LateOrdersResponse is how many orders are late and the ones waiting
// longest.
type LateOrdersResponse struct {
	Total  int64               `json:"total"`
	Orders []LateOrderResponse `json:"orders"`
}

// LateOrderResponse is an order still to collect whose pickup slot ended.
type LateOrderResponse struct {
	Code     string    `json:"code"`
	Status   string    `json:"status"`
	PickupAt time.Time `json:"pickup_at"`
}
