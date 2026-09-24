package postgres

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

func optionalSearch(search *string) sql.NullString {
	if search == nil || *search == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *search, Valid: true}
}

func optionalString(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}

// stringOrNil reads a nullable text column back into the domain's optional string.
func stringOrNil(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	value := v.String
	return &value
}

func optionalUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}

func optionalTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
