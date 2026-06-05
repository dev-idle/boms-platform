package postgres

import (
	"context"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domaincombo "github.com/boms/backend/internal/domain/combo"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/google/uuid"
)

type ComboRepository struct {
	queries *sqlcgen.Queries
}

func NewComboRepository(pool *Pool) *ComboRepository {
	return &ComboRepository{queries: pool.Queries()}
}

func (r *ComboRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *ComboRepository) Create(ctx context.Context, params port.CreateComboParams) (*domaincombo.Combo, error) {
	row, err := r.q(ctx).CreateCombo(ctx, sqlcgen.CreateComboParams{
		Name:       params.Name,
		Slug:       params.Slug,
		PriceCents: params.PriceCents,
		StartsAt:   params.StartsAt,
		EndsAt:     params.EndsAt,
		IsActive:   params.IsActive,
	})
	if err != nil {
		return nil, mapRepoError(err, "create combo")
	}
	return mapCombo(row), nil
}

func (r *ComboRepository) GetByID(ctx context.Context, id uuid.UUID) (*domaincombo.Combo, error) {
	row, err := r.q(ctx).GetComboByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err, "get combo")
	}
	return mapCombo(row), nil
}

func (r *ComboRepository) Update(ctx context.Context, params port.UpdateComboParams) (*domaincombo.Combo, error) {
	row, err := r.q(ctx).UpdateCombo(ctx, sqlcgen.UpdateComboParams{
		ID:         params.ID,
		Name:       params.Name,
		Slug:       params.Slug,
		PriceCents: params.PriceCents,
		StartsAt:   params.StartsAt,
		EndsAt:     params.EndsAt,
		IsActive:   params.IsActive,
	})
	if err != nil {
		return nil, mapRepoError(err, "update combo")
	}
	return mapCombo(row), nil
}

func (r *ComboRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.q(ctx).SoftDeleteCombo(ctx, id)
	if err != nil {
		return mapRepoError(err, "delete combo")
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *ComboRepository) ReplaceItems(ctx context.Context, comboID uuid.UUID, items []port.ComboItemParams) error {
	if err := r.q(ctx).DeleteComboItemsByComboID(ctx, comboID); err != nil {
		return mapRepoError(err, "clear combo items")
	}
	for _, item := range items {
		if err := r.q(ctx).InsertComboItem(ctx, sqlcgen.InsertComboItemParams{
			ComboID:   comboID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}); err != nil {
			return mapRepoError(err, "insert combo item")
		}
	}
	return nil
}

func (r *ComboRepository) ListItemsByComboID(ctx context.Context, comboID uuid.UUID) ([]domaincombo.Item, error) {
	rows, err := r.q(ctx).ListComboItemsByComboID(ctx, comboID)
	if err != nil {
		return nil, mapRepoError(err, "list combo items")
	}
	return mapComboItemRows(rows), nil
}

func (r *ComboRepository) ListItemsByComboIDs(
	ctx context.Context,
	comboIDs []uuid.UUID,
) (map[uuid.UUID][]domaincombo.Item, error) {
	if len(comboIDs) == 0 {
		return map[uuid.UUID][]domaincombo.Item{}, nil
	}
	rows, err := r.q(ctx).ListComboItemsByComboIDs(ctx, comboIDs)
	if err != nil {
		return nil, mapRepoError(err, "list combo items by combo ids")
	}
	return groupComboItems(mapComboItemRowsFromBatch(rows)), nil
}

func (r *ComboRepository) ListCatalogItemsByComboIDs(
	ctx context.Context,
	comboIDs []uuid.UUID,
) (map[uuid.UUID][]domaincombo.Item, error) {
	if len(comboIDs) == 0 {
		return map[uuid.UUID][]domaincombo.Item{}, nil
	}
	rows, err := r.q(ctx).ListCatalogComboItemsByComboIDs(ctx, comboIDs)
	if err != nil {
		return nil, mapRepoError(err, "list catalog combo items by combo ids")
	}
	return groupComboItems(mapCatalogComboItemRows(rows)), nil
}

func (r *ComboRepository) ManagerList(ctx context.Context, params port.ManagerListCombosParams) ([]domaincombo.Combo, error) {
	rows, err := r.q(ctx).ManagerListCombos(ctx, sqlcgen.ManagerListCombosParams{
		Limit:  params.Limit,
		Offset: params.Offset,
		Search: optionalSearch(params.Search),
	})
	if err != nil {
		return nil, mapRepoError(err, "manager list combos")
	}
	out := make([]domaincombo.Combo, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapCombo(row))
	}
	return out, nil
}

func (r *ComboRepository) ManagerListCount(ctx context.Context, search *string) (int64, error) {
	count, err := r.q(ctx).ManagerListCombosCount(ctx, optionalSearch(search))
	if err != nil {
		return 0, mapRepoError(err, "manager list combos count")
	}
	return count, nil
}

func (r *ComboRepository) CatalogList(ctx context.Context, params port.CatalogListCombosParams) ([]port.CatalogCombo, error) {
	rows, err := r.q(ctx).CatalogListCombos(ctx, sqlcgen.CatalogListCombosParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, mapRepoError(err, "catalog list combos")
	}
	comboIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		comboIDs = append(comboIDs, row.ID)
	}
	itemsByCombo, err := r.ListCatalogItemsByComboIDs(ctx, comboIDs)
	if err != nil {
		return nil, err
	}
	out := make([]port.CatalogCombo, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.CatalogCombo{
			ID:         row.ID,
			Name:       row.Name,
			Slug:       row.Slug,
			PriceCents: row.PriceCents,
			StartsAt:   row.StartsAt,
			EndsAt:     row.EndsAt,
			Items:      itemsByCombo[row.ID],
		})
	}
	return out, nil
}

func (r *ComboRepository) CatalogListCount(ctx context.Context) (int64, error) {
	count, err := r.q(ctx).CatalogListCombosCount(ctx)
	if err != nil {
		return 0, mapRepoError(err, "catalog list combos count")
	}
	return count, nil
}

func (r *ComboRepository) CatalogGetByID(ctx context.Context, id uuid.UUID) (*port.CatalogCombo, error) {
	row, err := r.q(ctx).CatalogGetComboByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err, "catalog get combo")
	}
	itemsByCombo, err := r.ListCatalogItemsByComboIDs(ctx, []uuid.UUID{row.ID})
	if err != nil {
		return nil, err
	}
	items := itemsByCombo[row.ID]
	if len(items) == 0 {
		return nil, apperrors.ErrNotFound
	}
	return &port.CatalogCombo{
		ID:         row.ID,
		Name:       row.Name,
		Slug:       row.Slug,
		PriceCents: row.PriceCents,
		StartsAt:   row.StartsAt,
		EndsAt:     row.EndsAt,
		Items:      items,
	}, nil
}

func (r *ComboRepository) CountAvailableProducts(ctx context.Context, productIDs []uuid.UUID) (int64, error) {
	if len(productIDs) == 0 {
		return 0, nil
	}
	count, err := r.q(ctx).CountAvailableProductsForCombo(ctx, productIDs)
	if err != nil {
		return 0, mapRepoError(err, "count available combo products")
	}
	return count, nil
}

func mapComboItem(
	id, comboID, productID uuid.UUID,
	quantity int32,
	productName, productSlug string,
	priceCents int64,
) domaincombo.Item {
	return domaincombo.Item{
		ID:          id,
		ComboID:     comboID,
		ProductID:   productID,
		Quantity:    quantity,
		ProductName: productName,
		ProductSlug: productSlug,
		PriceCents:  priceCents,
	}
}

func mapComboItemRows(rows []sqlcgen.ListComboItemsByComboIDRow) []domaincombo.Item {
	out := make([]domaincombo.Item, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapComboItem(
			row.ID, row.ComboID, row.ProductID,
			row.Quantity, row.ProductName, row.ProductSlug, row.PriceCents,
		))
	}
	return out
}

func mapComboItemRowsFromBatch(rows []sqlcgen.ListComboItemsByComboIDsRow) []domaincombo.Item {
	out := make([]domaincombo.Item, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapComboItem(
			row.ID, row.ComboID, row.ProductID,
			row.Quantity, row.ProductName, row.ProductSlug, row.PriceCents,
		))
	}
	return out
}

func mapCatalogComboItemRows(rows []sqlcgen.ListCatalogComboItemsByComboIDsRow) []domaincombo.Item {
	out := make([]domaincombo.Item, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapComboItem(
			row.ID, row.ComboID, row.ProductID,
			row.Quantity, row.ProductName, row.ProductSlug, row.PriceCents,
		))
	}
	return out
}

func groupComboItems(items []domaincombo.Item) map[uuid.UUID][]domaincombo.Item {
	out := make(map[uuid.UUID][]domaincombo.Item)
	for _, item := range items {
		out[item.ComboID] = append(out[item.ComboID], item)
	}
	return out
}

func mapCombo(row sqlcgen.Combo) *domaincombo.Combo {
	c := &domaincombo.Combo{
		ID:         row.ID,
		Name:       row.Name,
		Slug:       row.Slug,
		PriceCents: row.PriceCents,
		StartsAt:   row.StartsAt,
		EndsAt:     row.EndsAt,
		IsActive:   row.IsActive,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
	if row.DeletedAt.Valid {
		c.DeletedAt = &row.DeletedAt.Time
	}
	return c
}
