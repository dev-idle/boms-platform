package usecase

import (
	"strings"

	"github.com/boms/backend/internal/config"
	domainmedia "github.com/boms/backend/internal/domain/media"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// sanitizeManagerComboImageURL trims the one promotional image a combo carries and
// holds it to the same Cloudinary folder as product images: the manager uploads
// through the same signed endpoint, so an URL from anywhere else is not ours.
// A blank value clears the image.
func sanitizeManagerComboImageURL(
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
			"must be a Cloudinary image URL for this environment",
		)
	}

	return &trimmed, nil
}
