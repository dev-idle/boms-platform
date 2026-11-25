package usecase

import (
	"context"
	"errors"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// referencePhotoBatch is how many erased accounts one run clears of their
// reference photos; the next run takes the rest.
const referencePhotoBatch = 20

// ReferencePhotoUsecase deletes from Cloudinary the reference photos of the
// accounts erased, each customer's folder at once.
type ReferencePhotoUsecase struct {
	erasures   port.ReferencePhotoErasures
	images     port.ImageStore
	cloudinary config.CloudinaryConfig
}

func NewReferencePhotoUsecase(
	erasures port.ReferencePhotoErasures,
	images port.ImageStore,
	cloudinary config.CloudinaryConfig,
) *ReferencePhotoUsecase {
	return &ReferencePhotoUsecase{erasures: erasures, images: images, cloudinary: cloudinary}
}

// EraseDue deletes the photos of the erased accounts still waiting, a batch at
// a time, and returns how many accounts it cleared. Each stands alone: one
// Cloudinary refuses is reported with the rest and tried again on the next run.
func (u *ReferencePhotoUsecase) EraseDue(ctx context.Context) (int, error) {
	due, err := u.erasures.ListErasuresDue(ctx, referencePhotoBatch)
	if err != nil {
		return 0, err
	}
	erased := 0
	var failures []error
	for _, userID := range due {
		if err := u.images.DeleteFolder(ctx, u.cloudinary.CustomerReferenceFolder(userID)); err != nil {
			failures = append(failures, apperrors.Errorf("erase reference photos of %s: %w", userID, err))
			continue
		}
		if err := u.erasures.MarkErased(ctx, userID); err != nil {
			failures = append(failures, apperrors.Errorf("mark reference photos of %s erased: %w", userID, err))
			continue
		}
		erased++
	}
	return erased, errors.Join(failures...)
}
