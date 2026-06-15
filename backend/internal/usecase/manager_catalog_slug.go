package usecase

import (
	"strings"

	domaincatalog "github.com/boms/backend/internal/domain/catalog"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

const catalogSlugValidationDetail = "must be lowercase letters, numbers, and hyphens"

func normalizeManagerCatalogSlug(slug string) (string, error) {
	normalized, err := domaincatalog.NormalizeSlug(slug)
	if err != nil {
		return "", apperrors.ErrValidation.WithDetail("slug", catalogSlugValidationDetail)
	}
	return normalized, nil
}

func resolveManagerCatalogSlug(name, slug string, isCreate bool) (string, error) {
	if trimmed := strings.TrimSpace(slug); trimmed != "" {
		return normalizeManagerCatalogSlug(trimmed)
	}
	if !isCreate {
		return "", apperrors.ErrValidation.WithDetail("slug", catalogSlugValidationDetail)
	}
	derived, err := domaincatalog.SlugFromName(name)
	if err != nil {
		return "", apperrors.ErrValidation.WithDetail("slug", catalogSlugValidationDetail)
	}
	return derived, nil
}
