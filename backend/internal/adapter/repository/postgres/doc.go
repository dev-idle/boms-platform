// Package postgres: pgx pool + sqlc (sql/{schema,query}, generated sqlcgen/ on native pgx).
// Schema source of truth: db/schema.hcl + migrations/ (Atlas). Keep sql/schema in sync; regenerate: make sqlc.
package postgres
