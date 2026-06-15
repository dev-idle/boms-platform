package usecase

import (
	"strings"

	"github.com/boms/backend/internal/config"
	domainmedia "github.com/boms/backend/internal/domain/media"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

func sanitizeManagerProductImageURL(
	cfg config.CloudinaryConfig,
	raw *string,
) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}
	if cfg.Enabled() &&
		!domainmedia.IsCloudinaryDeliveryURLInFolder(cfg.CloudName, cfg.ResolvedUploadFolder(), trimmed) {
		return nil, apperrors.ErrValidation.WithDetail(
			"image_url",
			"must be a Cloudinary product image URL for this environment",
		)
	}
	return &trimmed, nil
}
