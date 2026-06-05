package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	domaincatalog "github.com/boms/backend/internal/domain/catalog"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ManagerDiscountCodeUsecase struct {
	codes port.DiscountCodeRepository
	audit *auditlogger.Service
	log   *zap.Logger
}

func NewManagerDiscountCodeUsecase(
	codes port.DiscountCodeRepository,
	audit *auditlogger.Service,
	log *zap.Logger,
) *ManagerDiscountCodeUsecase {
	return &ManagerDiscountCodeUsecase{codes: codes, audit: audit, log: log}
}

func (u *ManagerDiscountCodeUsecase) Create(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	req dto.CreateDiscountCodeRequest,
) (*dto.DiscountCodeResponse, error) {
	params, err := parseDiscountCodeRequest(req.Code, req.DiscountType, req.Value, req.MinOrderCents, req.MaxUses, req.StartsAt, req.EndsAt, req.IsActive, nil)
	if err != nil {
		return nil, err
	}

	created, err := u.codes.Create(ctx, params)
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, domaindiscount.ErrCodeExists
		}
		return nil, err
	}

	resp := toDiscountCodeResponse(created)
	u.logAudit(ctx, domaincatalog.AuditActionManagerCreatedDiscountCode, actorID, actorRole, &created.ID, "discount_code", nil, toDiscountCodeAudit(resp))
	return resp, nil
}

func (u *ManagerDiscountCodeUsecase) Get(ctx context.Context, id uuid.UUID) (*dto.DiscountCodeResponse, error) {
	code, err := u.codes.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domaindiscount.ErrNotFound
		}
		return nil, err
	}
	return toDiscountCodeResponse(code), nil
}

func (u *ManagerDiscountCodeUsecase) List(
	ctx context.Context,
	page, pageSize int32,
	search string,
) ([]dto.DiscountCodeResponse, int64, int32, int32, error) {
	page, pageSize = normalizeCatalogListPage(page, pageSize)
	var searchPtr *string
	trimmed := strings.TrimSpace(search)
	if trimmed != "" {
		searchPtr = &trimmed
	}
	codes, err := u.codes.ManagerList(ctx, port.ManagerListDiscountCodesParams{
		Search: searchPtr,
		Limit:  pageSize,
		Offset: utils.PageOffset(page, pageSize),
	})
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	total, err := u.codes.ManagerListCount(ctx, searchPtr)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]dto.DiscountCodeResponse, 0, len(codes))
	for i := range codes {
		out = append(out, *toDiscountCodeResponse(&codes[i]))
	}
	return out, total, page, pageSize, nil
}

func (u *ManagerDiscountCodeUsecase) Update(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	id uuid.UUID,
	req dto.UpdateDiscountCodeRequest,
) (*dto.DiscountCodeResponse, error) {
	beforeEntity, err := u.codes.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domaindiscount.ErrNotFound
		}
		return nil, err
	}
	before := toDiscountCodeResponse(beforeEntity)

	params, err := parseDiscountCodeRequest(req.Code, req.DiscountType, req.Value, req.MinOrderCents, req.MaxUses, req.StartsAt, req.EndsAt, req.IsActive, &beforeEntity.UsedCount)
	if err != nil {
		return nil, err
	}
	updateParams := port.UpdateDiscountCodeParams{
		ID:            id,
		Code:          params.Code,
		DiscountType:  params.DiscountType,
		Value:         params.Value,
		MinOrderCents: params.MinOrderCents,
		MaxUses:       params.MaxUses,
		StartsAt:      params.StartsAt,
		EndsAt:        params.EndsAt,
		IsActive:      params.IsActive,
	}

	updated, err := u.codes.Update(ctx, updateParams)
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, domaindiscount.ErrCodeExists
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domaindiscount.ErrNotFound
		}
		return nil, err
	}

	after := toDiscountCodeResponse(updated)
	u.logAudit(ctx, domaincatalog.AuditActionManagerUpdatedDiscountCode, actorID, actorRole, &id, "discount_code", toDiscountCodeAudit(before), toDiscountCodeAudit(after))
	return after, nil
}

func (u *ManagerDiscountCodeUsecase) Delete(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	id uuid.UUID,
) error {
	beforeEntity, err := u.codes.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return domaindiscount.ErrNotFound
		}
		return err
	}
	before := toDiscountCodeResponse(beforeEntity)

	if err := u.codes.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return domaindiscount.ErrNotFound
		}
		return err
	}
	u.logAudit(ctx, domaincatalog.AuditActionManagerDeletedDiscountCode, actorID, actorRole, &id, "discount_code", toDiscountCodeAudit(before), nil)
	return nil
}

func parseDiscountCodeRequest(
	codeRaw, discountTypeRaw string,
	value int64,
	minOrderCents *int64,
	maxUses *int32,
	startsAt, endsAt time.Time,
	isActive bool,
	usedCount *int32,
) (port.CreateDiscountCodeParams, error) {
	code, err := domaindiscount.NormalizeCode(codeRaw)
	if err != nil {
		return port.CreateDiscountCodeParams{}, apperrors.ErrValidation.WithDetail("code", "must be 3-64 uppercase letters, numbers, hyphens, or underscores")
	}
	discountType := domaindiscount.Type(discountTypeRaw)
	if !discountType.Valid() {
		return port.CreateDiscountCodeParams{}, apperrors.ErrValidation.WithDetail("discount_type", "must be percent or fixed_cents")
	}
	if !endsAt.After(startsAt) {
		return port.CreateDiscountCodeParams{}, apperrors.ErrValidation.WithDetail("ends_at", "must be after starts_at")
	}
	if err := validateDiscountValue(discountType, value); err != nil {
		return port.CreateDiscountCodeParams{}, err
	}
	if maxUses != nil && usedCount != nil && *maxUses < *usedCount {
		return port.CreateDiscountCodeParams{}, apperrors.ErrValidation.WithDetail("max_uses", "cannot be less than used count")
	}

	return port.CreateDiscountCodeParams{
		Code:          code,
		DiscountType:  discountType,
		Value:         value,
		MinOrderCents: minOrderCents,
		MaxUses:       maxUses,
		StartsAt:      startsAt,
		EndsAt:        endsAt,
		IsActive:      isActive,
	}, nil
}

func validateDiscountValue(discountType domaindiscount.Type, value int64) error {
	switch discountType {
	case domaindiscount.TypePercent:
		if value < 1 || value > 100 {
			return apperrors.ErrValidation.WithDetail("value", "percent must be between 1 and 100")
		}
	case domaindiscount.TypeFixedCents:
		if value < 1 {
			return apperrors.ErrValidation.WithDetail("value", "fixed discount must be at least 1 cent")
		}
	}
	return nil
}

func toDiscountCodeResponse(code *domaindiscount.Code) *dto.DiscountCodeResponse {
	return &dto.DiscountCodeResponse{
		ID:            code.ID.String(),
		Code:          code.Code,
		DiscountType:  string(code.DiscountType),
		Value:         code.Value,
		MinOrderCents: code.MinOrderCents,
		MaxUses:       code.MaxUses,
		UsedCount:     code.UsedCount,
		StartsAt:      code.StartsAt,
		EndsAt:        code.EndsAt,
		IsActive:      code.IsActive,
		CreatedAt:     code.CreatedAt,
		UpdatedAt:     code.UpdatedAt,
	}
}

func toDiscountCodeAudit(resp *dto.DiscountCodeResponse) map[string]any {
	return map[string]any{
		"code":          resp.Code,
		"discount_type": resp.DiscountType,
		"value":         resp.Value,
		"is_active":     resp.IsActive,
		"starts_at":     resp.StartsAt,
		"ends_at":       resp.EndsAt,
	}
}

func (u *ManagerDiscountCodeUsecase) logAudit(
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
