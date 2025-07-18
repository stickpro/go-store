// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: collections.sql

package repository_collections

import (
	"context"

	"github.com/stickpro/go-store/internal/models"
)

const getBySlug = `-- name: GetBySlug :one
SELECT id, name, description, slug, created_at, updated_at FROM collections WHERE slug=$1 LIMIT 1
`

func (q *Queries) GetBySlug(ctx context.Context, slug string) (*models.Collection, error) {
	row := q.db.QueryRow(ctx, getBySlug, slug)
	var i models.Collection
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Slug,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}
