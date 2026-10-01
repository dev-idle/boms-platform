package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	domaincatalog "github.com/boms/backend/internal/domain/catalog"
	domainproduct "github.com/boms/backend/internal/domain/product"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
)

// StaffProductUsecase is the counter marking products sold out for the day.
type StaffProductUsecase struct {
	products port.ProductRepository
	tx       port.TxManager
	events   port.EventOutbox
	audit    *auditlogger.Service
	log      *zap.Logger
}

func NewStaffProductUsecase(
	products port.ProductRepository,
	tx port.TxManager,
	events port.EventOutbox,
	audit *auditlogger.Service,
	log *zap.Logger,
) *StaffProductUsecase {
	return &StaffProductUsecase{products: products, tx: tx, events: events, audit: audit, log: log}
}

// List returns a page of the products the counter sells, by category and
// name; soldOutToday narrows it to those out today.
func (u *StaffProductUsecase) List(
	ctx context.Context,
	page, pageSize int32,
	soldOutToday bool,
) ([]dto.StaffProductResponse, int64, int32, int32, error) {
	page, pageSize = normalizeCatalogListPage(page, pageSize)
	today := domainstore.DayOf(time.Now())
	var soldOutOn *time.Time
	if soldOutToday {
		soldOutOn = &today
	}
	rows, total, err := listWithTotal(ctx,
		func(ctx context.Context) ([]port.StaffProduct, error) {
			return u.products.StaffList(ctx, port.StaffListProductsParams{
				SoldOutOn: soldOutOn,
				Limit:     pageSize,
				Offset:    utils.PageOffset(page, pageSize),
			})
		},
		func(ctx context.Context) (int64, error) {
			return u.products.StaffListCount(ctx, soldOutOn)
		},
	)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]dto.StaffProductResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toStaffProductResponse(row, today))
	}
	return out, total, page, pageSize, nil
}

// SetSoldOut marks a product sold out for today, so no order collected today
// may hold it, or back. Open carts and pickers hear of it at once.
func (u *StaffProductUsecase) SetSoldOut(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	productID uuid.UUID,
	soldOut bool,
) (*dto.StaffProductResponse, error) {
	today := domainstore.DayOf(time.Now())
	var day *time.Time
	if soldOut {
		day = &today
	}
	var updated *port.StaffProduct
	err := u.tx.WithTx(ctx, func(txCtx context.Context) (err error) {
		if updated, err = u.products.SetSoldOut(txCtx, productID, day); err != nil {
			return err
		}
		return u.events.Add(txCtx, domainproduct.SoldOutChangedEvent(productID))
	})
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, domainproduct.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	recordAudit(u.log, u.audit, ctx, domaincatalog.AuditActionStaffMarkedSoldOut, actorID, actorRole, &productID, "product",
		nil, map[string]bool{"sold_out_today": soldOut})
	resp := toStaffProductResponse(*updated, today)
	return &resp, nil
}

func toStaffProductResponse(p port.StaffProduct, today time.Time) dto.StaffProductResponse {
	return dto.StaffProductResponse{
		ID:           p.ID.String(),
		Name:         p.Name,
		CategoryName: p.CategoryName,
		Station:      string(p.Station),
		SoldOutToday: p.SoldOutOn != nil && p.SoldOutOn.Equal(today),
	}
}
