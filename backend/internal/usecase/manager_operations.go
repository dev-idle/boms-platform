package usecase

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
)

// operationsLateRows is how many late orders the dashboard lists.
const operationsLateRows int32 = 10

// ManagerOperationsUsecase is a manager watching the bakery's work as it goes.
type ManagerOperationsUsecase struct {
	operations port.OperationsRepository
	sales      port.SalesReportRepository
}

func NewManagerOperationsUsecase(operations port.OperationsRepository, sales port.SalesReportRepository) *ManagerOperationsUsecase {
	return &ManagerOperationsUsecase{operations: operations, sales: sales}
}

// Dashboard reads where the work stands at now: the orders still to make or
// hand over, each station's tickets, the bakery day's pickups, the orders
// late, and how long the orders that became ready today took. Its parts are
// read at once.
func (u *ManagerOperationsUsecase) Dashboard(ctx context.Context, now time.Time) (*dto.OperationsResponse, error) {
	local := now.In(domainstore.Location)
	todayStart := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, domainstore.Location)
	tomorrowStart := todayStart.AddDate(0, 0, 1)
	var (
		active     port.ActiveOrders
		stations   []port.StationLoad
		pickups    port.DayPickups
		late       []port.LateOrder
		lateTotal  int64
		production port.ProductionTime
	)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		active, err = u.operations.ActiveOrders(gctx)
		return err
	})
	g.Go(func() (err error) {
		stations, err = u.operations.StationLoads(gctx)
		return err
	})
	g.Go(func() (err error) {
		pickups, err = u.operations.DayPickups(gctx, todayStart, tomorrowStart)
		return err
	})
	g.Go(func() (err error) {
		late, lateTotal, err = u.operations.LateOrders(gctx, now, operationsLateRows)
		return err
	})
	g.Go(func() (err error) {
		production, err = u.sales.ProductionTime(gctx, todayStart, tomorrowStart)
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	out := &dto.OperationsResponse{
		Orders: dto.ActiveOrdersResponse{
			Pending:      active.Pending,
			Confirmed:    active.Confirmed,
			InProduction: active.InProduction,
			Ready:        active.Ready,
		},
		Stations: make([]dto.StationLoadResponse, 0, len(stations)),
		Pickups: dto.DayPickupsResponse{
			Due:       pickups.Due,
			Collected: pickups.Collected,
			ToCollect: pickups.ToCollect,
			Missed:    pickups.Missed,
		},
		Late:       dto.LateOrdersResponse{Total: lateTotal, Orders: make([]dto.LateOrderResponse, 0, len(late))},
		Production: mapProductionTimeToDTO(production),
	}
	for _, station := range stations {
		out.Stations = append(out.Stations, dto.StationLoadResponse{
			Station:    string(station.Station),
			Queued:     station.Queued,
			InProgress: station.InProgress,
			Ready:      station.Ready,
		})
	}
	for _, order := range late {
		out.Late.Orders = append(out.Late.Orders, dto.LateOrderResponse{
			Code: order.Code, Status: string(order.Status), PickupAt: order.PickupAt,
		})
	}
	return out, nil
}
