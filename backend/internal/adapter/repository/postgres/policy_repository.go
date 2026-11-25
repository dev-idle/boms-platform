package postgres

import (
	"context"
	"time"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
)

// PolicyRepository records the policy versions the API serves.
type PolicyRepository struct {
	queries *sqlcgen.Queries
}

func NewPolicyRepository(pool *Pool) *PolicyRepository {
	return &PolicyRepository{queries: pool.Queries()}
}

// Publish records version as served and returns when an API first served it.
func (r *PolicyRepository) Publish(ctx context.Context, version string) (time.Time, error) {
	publishedAt, err := r.queries.PublishPolicyVersion(ctx, version)
	if err != nil {
		return time.Time{}, mapRepoError(err, "publish policy version")
	}
	return publishedAt, nil
}
