package port

import (
	"context"
	"time"

	domaincombo "github.com/boms/backend/internal/domain/combo"
	"github.com/google/uuid"
)

type CreateComboParams struct {
	Name       string
	Slug       string
	PriceCents int64
	StartsAt   time.Time
	EndsAt     time.Time
	IsActive   bool
}

type UpdateComboParams struct {
	ID         uuid.UUID
	Name       string
	Slug       string
	PriceCents int64
	StartsAt   time.Time
	EndsAt     time.Time
	IsActive   bool
}

type ComboItemParams struct {
	ProductID uuid.UUID
	Quantity  int32
}

type ManagerListCombosParams struct {
	Search *string
	Limit  int32
	Offset int32
}

type CatalogListCombosParams struct {
	Limit  int32
	Offset int32
}

type CatalogCombo struct {
	ID         uuid.UUID
	Name       string
	Slug       string
	PriceCents int64
	StartsAt   time.Time
	EndsAt     time.Time
	Items      []domaincombo.Item
}

type ComboRepository interface {
	Create(ctx context.Context, params CreateComboParams) (*domaincombo.Combo, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domaincombo.Combo, error)
	Update(ctx context.Context, params UpdateComboParams) (*domaincombo.Combo, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ReplaceItems(ctx context.Context, comboID uuid.UUID, items []ComboItemParams) error
	ListItemsByComboID(ctx context.Context, comboID uuid.UUID) ([]domaincombo.Item, error)
	ListItemsByComboIDs(ctx context.Context, comboIDs []uuid.UUID) (map[uuid.UUID][]domaincombo.Item, error)
	ListCatalogItemsByComboIDs(ctx context.Context, comboIDs []uuid.UUID) (map[uuid.UUID][]domaincombo.Item, error)
	ManagerList(ctx context.Context, params ManagerListCombosParams) ([]domaincombo.Combo, error)
	ManagerListCount(ctx context.Context, search *string) (int64, error)
	CatalogList(ctx context.Context, params CatalogListCombosParams) ([]CatalogCombo, error)
	CatalogListCount(ctx context.Context) (int64, error)
	CatalogGetByID(ctx context.Context, id uuid.UUID) (*CatalogCombo, error)
	CountAvailableProducts(ctx context.Context, productIDs []uuid.UUID) (int64, error)
}
