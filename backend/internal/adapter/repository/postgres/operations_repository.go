package postgres

import (
	"context"
	"time"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
)

// OperationsRepository reads where the bakery's work stands now.
type OperationsRepository struct {
	queries *sqlcgen.Queries
}

func NewOperationsRepository(pool *Pool) *OperationsRepository {
	return &OperationsRepository{queries: pool.Queries()}
}

var _ port.OperationsRepository = (*OperationsRepository)(nil)

func (r *OperationsRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

// ActiveOrders implements port.OperationsRepository.
func (r *OperationsRepository) ActiveOrders(ctx context.Context) (port.ActiveOrders, error) {
	row, err := r.q(ctx).OperationsActiveOrders(ctx)
	if err != nil {
		return port.ActiveOrders{}, mapRepoError(err, "count active orders")
	}
	return port.ActiveOrders{
		Pending:      row.Pending,
		Confirmed:    row.Confirmed,
		InProduction: row.InProduction,
		Ready:        row.Ready,
	}, nil
}

// StationLoads implements port.OperationsRepository.
func (r *OperationsRepository) StationLoads(ctx context.Context) ([]port.StationLoad, error) {
	rows, err := r.q(ctx).OperationsStationLoads(ctx)
	if err != nil {
		return nil, mapRepoError(err, "count station tickets")
	}
	out := make([]port.StationLoad, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.StationLoad{
			Station:    domaincategory.Station(row.Station),
			Queued:     row.Queued,
			InProgress: row.InProgress,
			Ready:      row.Ready,
		})
	}
	return out, nil
}

// DayPickups implements port.OperationsRepository.
func (r *OperationsRepository) DayPickups(ctx context.Context, from, to time.Time) (port.DayPickups, error) {
	row, err := r.q(ctx).OperationsDayPickups(ctx, sqlcgen.OperationsDayPickupsParams{FromAt: from, ToAt: to})
	if err != nil {
		return port.DayPickups{}, mapRepoError(err, "count day pickups")
	}
	return port.DayPickups{Due: row.Due, Collected: row.Collected, ToCollect: row.ToCollect, Missed: row.Missed}, nil
}

// LateOrders implements port.OperationsRepository.
func (r *OperationsRepository) LateOrders(ctx context.Context, now time.Time, limit int32) ([]port.LateOrder, int64, error) {
	rows, err := r.q(ctx).OperationsLateOrders(ctx, sqlcgen.OperationsLateOrdersParams{Now: now, MaxRows: limit})
	if err != nil {
		return nil, 0, mapRepoError(err, "list late orders")
	}
	out := make([]port.LateOrder, 0, len(rows))
	var total int64
	for _, row := range rows {
		out = append(out, port.LateOrder{Code: row.Code, Status: domainorder.Status(row.Status), PickupAt: row.PickupAt})
		total = row.Total
	}
	return out, total, nil
}
