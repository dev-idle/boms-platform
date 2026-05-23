// Package postgres: pgx pool + sqlx + sqlc (sql/{schema,query}, generated sqlcgen/).
// Schema source of truth: db/schema.hcl + migrations/ (Atlas). Keep sql/schema in sync; regenerate: make sqlc.
package postgres
