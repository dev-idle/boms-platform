package port

import (
	"context"

	"github.com/google/uuid"
)

// ReferencePhotoErasures are the erased customers whose reference photos are
// still to be deleted.
type ReferencePhotoErasures interface {
	// RequestErasure asks, in the erasure's transaction, for the customer's
	// photos to be deleted.
	RequestErasure(ctx context.Context, userID uuid.UUID) error
	// ListErasuresDue returns up to limit customers whose photos are still to
	// be deleted, the oldest request first.
	ListErasuresDue(ctx context.Context, limit int32) ([]uuid.UUID, error)
	// MarkErased records that the customer's photos are deleted.
	MarkErased(ctx context.Context, userID uuid.UUID) error
}

// ImageStore holds the images uploaded to the shop (adapter/cloudinary).
type ImageStore interface {
	// DeleteFolder deletes every image under folder, and the copies the CDN
	// keeps of them.
	DeleteFolder(ctx context.Context, folder string) error
}
