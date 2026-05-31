package postgres

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	apperrors "github.com/boms/backend/internal/shared/errors"
)

// mapRepoError maps sqlc/pgx failures to application errors for usecases and handlers.
func mapRepoError(err error, op string) error {
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
		return apperrors.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperrors.ErrConflict
	}
	return apperrors.Errorf("%s: %w", op, err)
}
