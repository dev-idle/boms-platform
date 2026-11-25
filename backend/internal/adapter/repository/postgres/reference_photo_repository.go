package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	"github.com/boms/backend/internal/port"
)

// ReferencePhotoRepository keeps the reference photo erasures still to do.
type ReferencePhotoRepository struct {
	queries *sqlcgen.Queries
}

func NewReferencePhotoRepository(pool *Pool) *ReferencePhotoRepository {
	return &ReferencePhotoRepository{queries: pool.Queries()}
}

func (r *ReferencePhotoRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

// RequestErasure implements port.ReferencePhotoErasures.
func (r *ReferencePhotoRepository) RequestErasure(ctx context.Context, userID uuid.UUID) error {
	if err := r.q(ctx).RequestReferencePhotoErasure(ctx, userID); err != nil {
		return mapRepoError(err, "request reference photo erasure")
	}
	return nil
}

// ListErasuresDue implements port.ReferencePhotoErasures.
func (r *ReferencePhotoRepository) ListErasuresDue(ctx context.Context, limit int32) ([]uuid.UUID, error) {
	ids, err := r.q(ctx).ListReferencePhotoErasuresDue(ctx, limit)
	if err != nil {
		return nil, mapRepoError(err, "list reference photo erasures")
	}
	return ids, nil
}

// MarkErased implements port.ReferencePhotoErasures.
func (r *ReferencePhotoRepository) MarkErased(ctx context.Context, userID uuid.UUID) error {
	if err := r.q(ctx).MarkReferencePhotosErased(ctx, userID); err != nil {
		return mapRepoError(err, "mark reference photos erased")
	}
	return nil
}

var _ port.ReferencePhotoErasures = (*ReferencePhotoRepository)(nil)
