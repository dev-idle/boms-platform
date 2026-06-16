package usecase

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	domaincatalog "github.com/boms/backend/internal/domain/catalog"
	domaincombo "github.com/boms/backend/internal/domain/combo"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ManagerComboUsecase struct {
	combos port.ComboRepository
	tx     port.TxManager
	audit  *auditlogger.Service
	log    *zap.Logger
}

func NewManagerComboUsecase(
	combos port.ComboRepository,
	tx port.TxManager,
	audit *auditlogger.Service,
	log *zap.Logger,
) *ManagerComboUsecase {
	return &ManagerComboUsecase{combos: combos, tx: tx, audit: audit, log: log}
}

func (u *ManagerComboUsecase) Create(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	req dto.CreateComboRequest,
) (*dto.ComboResponse, error) {
	parsed, err := parseComboRequest(req.Name, req.Slug, req.PriceCents, req.StartsAt, req.EndsAt, req.IsActive, req.Items, true)
	if err != nil {
		return nil, err
	}

	var createdID uuid.UUID
	if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := u.ensureComboProductsAvailable(txCtx, parsed.productIDs); err != nil {
			return err
		}
		created, createErr := u.combos.Create(txCtx, parsed.params)
		if createErr != nil {
			if errors.Is(createErr, apperrors.ErrConflict) {
				return domaincombo.ErrSlugExists
			}
			return createErr
		}
		createdID = created.ID
		return u.combos.ReplaceItems(txCtx, created.ID, parsed.itemParams)
	}); err != nil {
		return nil, err
	}

	resp, err := u.Get(ctx, createdID)
	if err != nil {
		return nil, err
	}
	u.logAudit(ctx, domaincatalog.AuditActionManagerCreatedCombo, actorID, actorRole, &createdID, "combo", nil, toComboAudit(resp))
	return resp, nil
}

func (u *ManagerComboUsecase) Get(ctx context.Context, id uuid.UUID) (*dto.ComboResponse, error) {
	combo, err := u.combos.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domaincombo.ErrNotFound
		}
		return nil, err
	}
	items, err := u.combos.ListItemsByComboID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toComboResponse(combo, items), nil
}

func (u *ManagerComboUsecase) List(
	ctx context.Context,
	page, pageSize int32,
	search string,
) ([]dto.ComboResponse, int64, int32, int32, error) {
	page, pageSize = normalizeCatalogListPage(page, pageSize)
	var searchPtr *string
	trimmed := strings.TrimSpace(search)
	if trimmed != "" {
		searchPtr = &trimmed
	}
	combos, err := u.combos.ManagerList(ctx, port.ManagerListCombosParams{
		Search: searchPtr,
		Limit:  pageSize,
		Offset: utils.PageOffset(page, pageSize),
	})
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	total, err := u.combos.ManagerListCount(ctx, searchPtr)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	comboIDs := make([]uuid.UUID, 0, len(combos))
	for _, combo := range combos {
		comboIDs = append(comboIDs, combo.ID)
	}
	itemsByCombo, err := u.combos.ListItemsByComboIDs(ctx, comboIDs)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]dto.ComboResponse, 0, len(combos))
	for _, combo := range combos {
		out = append(out, *toComboResponse(&combo, itemsByCombo[combo.ID]))
	}
	return out, total, page, pageSize, nil
}

func (u *ManagerComboUsecase) Update(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	id uuid.UUID,
	req dto.UpdateComboRequest,
) (*dto.ComboResponse, error) {
	before, err := u.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	parsed, parseErr := parseComboRequest(req.Name, req.Slug, req.PriceCents, req.StartsAt, req.EndsAt, req.IsActive, req.Items, false)
	if parseErr != nil {
		return nil, parseErr
	}
	updateParams := port.UpdateComboParams{
		ID:         id,
		Name:       parsed.params.Name,
		Slug:       parsed.params.Slug,
		PriceCents: parsed.params.PriceCents,
		StartsAt:   parsed.params.StartsAt,
		EndsAt:     parsed.params.EndsAt,
		IsActive:   parsed.params.IsActive,
	}

	if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		if err := u.ensureComboProductsAvailable(txCtx, parsed.productIDs); err != nil {
			return err
		}
		if _, updateErr := u.combos.Update(txCtx, updateParams); updateErr != nil {
			if errors.Is(updateErr, apperrors.ErrConflict) {
				return domaincombo.ErrSlugExists
			}
			if errors.Is(updateErr, apperrors.ErrNotFound) {
				return domaincombo.ErrNotFound
			}
			return updateErr
		}
		return u.combos.ReplaceItems(txCtx, id, parsed.itemParams)
	}); err != nil {
		return nil, err
	}

	after, err := u.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	u.logAudit(ctx, domaincatalog.AuditActionManagerUpdatedCombo, actorID, actorRole, &id, "combo", toComboAudit(before), toComboAudit(after))
	return after, nil
}

func (u *ManagerComboUsecase) Delete(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	id uuid.UUID,
) error {
	before, err := u.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := u.combos.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return domaincombo.ErrNotFound
		}
		return err
	}
	u.logAudit(ctx, domaincatalog.AuditActionManagerDeletedCombo, actorID, actorRole, &id, "combo", toComboAudit(before), nil)
	return nil
}

type parsedComboRequest struct {
	params     port.CreateComboParams
	itemParams []port.ComboItemParams
	productIDs []uuid.UUID
}

func parseComboRequest(
	nameRaw, slugRaw string,
	priceCents int64,
	startsAt, endsAt time.Time,
	isActive bool,
	items []dto.ComboItemInput,
	deriveSlugFromName bool,
) (parsedComboRequest, error) {
	name := strings.TrimSpace(nameRaw)
	if name == "" {
		return parsedComboRequest{}, apperrors.ErrValidation.WithDetail("name", "required")
	}
	slug, err := resolveManagerCatalogSlug(name, slugRaw, deriveSlugFromName)
	if err != nil {
		return parsedComboRequest{}, err
	}
	if !endsAt.After(startsAt) {
		return parsedComboRequest{}, apperrors.ErrValidation.WithDetail("ends_at", "must be after starts_at")
	}

	itemParams, productIDs, err := parseComboItems(items)
	if err != nil {
		return parsedComboRequest{}, err
	}

	return parsedComboRequest{
		params: port.CreateComboParams{
			Name:       name,
			Slug:       slug,
			PriceCents: priceCents,
			StartsAt:   startsAt,
			EndsAt:     endsAt,
			IsActive:   isActive,
		},
		itemParams: itemParams,
		productIDs: productIDs,
	}, nil
}

func (u *ManagerComboUsecase) ensureComboProductsAvailable(ctx context.Context, productIDs []uuid.UUID) error {
	available, err := u.combos.CountAvailableProducts(ctx, productIDs)
	if err != nil {
		return err
	}
	if available != int64(len(productIDs)) {
		return apperrors.ErrValidation.WithDetail("items", "one or more products are unavailable")
	}
	return nil
}

func parseComboItems(items []dto.ComboItemInput) ([]port.ComboItemParams, []uuid.UUID, error) {
	if len(items) == 0 {
		return nil, nil, apperrors.ErrValidation.WithDetail("items", "at least one item is required")
	}
	seen := make(map[uuid.UUID]struct{}, len(items))
	itemParams := make([]port.ComboItemParams, 0, len(items))
	productIDs := make([]uuid.UUID, 0, len(items))
	for i, item := range items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, nil, apperrors.ErrValidation.WithDetail("items", "invalid product_id at index "+strconv.Itoa(i))
		}
		if _, dup := seen[productID]; dup {
			return nil, nil, apperrors.ErrValidation.WithDetail("items", "duplicate product in combo")
		}
		seen[productID] = struct{}{}
		itemParams = append(itemParams, port.ComboItemParams{
			ProductID: productID,
			Quantity:  item.Quantity,
		})
		productIDs = append(productIDs, productID)
	}
	if err := validateComboBundleSize(items); err != nil {
		return nil, nil, err
	}
	return itemParams, productIDs, nil
}

const comboBundleMinDetail = "Add at least two products, or one product with quantity of at least 2"

func validateComboBundleSize(items []dto.ComboItemInput) error {
	if len(items) >= 2 {
		return nil
	}
	if len(items) == 1 && items[0].Quantity >= 2 {
		return nil
	}
	return apperrors.ErrValidation.WithDetail("items", comboBundleMinDetail)
}

func toComboResponse(combo *domaincombo.Combo, items []domaincombo.Item) *dto.ComboResponse {
	return &dto.ComboResponse{
		ID:         combo.ID.String(),
		Name:       combo.Name,
		Slug:       combo.Slug,
		PriceCents: combo.PriceCents,
		StartsAt:   combo.StartsAt,
		EndsAt:     combo.EndsAt,
		IsActive:   combo.IsActive,
		Items:      toComboItemResponses(items),
		CreatedAt:  combo.CreatedAt,
		UpdatedAt:  combo.UpdatedAt,
	}
}

func toComboAudit(resp *dto.ComboResponse) map[string]any {
	return map[string]any{
		"name":        resp.Name,
		"slug":        resp.Slug,
		"price_cents": resp.PriceCents,
		"starts_at":   resp.StartsAt,
		"ends_at":     resp.EndsAt,
		"is_active":   resp.IsActive,
		"item_count":  len(resp.Items),
	}
}

func (u *ManagerComboUsecase) logAudit(
	ctx context.Context,
	action domainuser.AuditAction,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	targetID *uuid.UUID,
	targetType string,
	before, after any,
) {
	recordAudit(u.log, u.audit, ctx, action, actorID, actorRole, targetID, targetType, before, after)
}
