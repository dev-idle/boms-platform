package usecase

import (
	"strings"

	"github.com/boms/backend/internal/config"
	domainmedia "github.com/boms/backend/internal/domain/media"
	domainproduct "github.com/boms/backend/internal/domain/product"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

func sanitizeManagerProductImageURLs(
	cfg config.CloudinaryConfig,
	raw []string,
) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if len(raw) > domainproduct.MaxImagesPerProduct {
		return nil, apperrors.ErrValidation.WithDetail(
			"image_urls",
			"maximum 5 images allowed",
		)
	}

	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			return nil, apperrors.ErrValidation.WithDetail(
				"image_urls",
				"duplicate image URLs are not allowed",
			)
		}
		if cfg.Enabled() &&
			!domainmedia.IsCloudinaryDeliveryURLInFolder(cfg.CloudName, cfg.ResolvedUploadFolder(), trimmed) {
			return nil, apperrors.ErrValidation.WithDetail(
				"image_urls",
				"must be Cloudinary product image URLs for this environment",
			)
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	if len(out) > domainproduct.MaxImagesPerProduct {
		return nil, apperrors.ErrValidation.WithDetail(
			"image_urls",
			"maximum 5 images allowed",
		)
	}
	return out, nil
}
