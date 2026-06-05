package postgres

import (
	"database/sql"

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

func optionalUUID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: *id, Valid: true}
}
