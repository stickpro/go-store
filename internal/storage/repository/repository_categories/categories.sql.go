// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: categories.sql

package repository_categories

import (
	"context"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
)

const getByID = `-- name: GetByID :one
SELECT id, parent_id, name, slug, description, meta_title, meta_h1, meta_description, meta_keyword, is_enable, created_at, updated_at FROM categories WHERE id = $1 LIMIT 1
`

func (q *Queries) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	row := q.db.QueryRow(ctx, getByID, id)
	var i models.Category
	err := row.Scan(
		&i.ID,
		&i.ParentID,
		&i.Name,
		&i.Slug,
		&i.Description,
		&i.MetaTitle,
		&i.MetaH1,
		&i.MetaDescription,
		&i.MetaKeyword,
		&i.IsEnable,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const getBySlug = `-- name: GetBySlug :one
SELECT id, parent_id, name, slug, description, meta_title, meta_h1, meta_description, meta_keyword, is_enable, created_at, updated_at FROM categories WHERE slug = $1 LIMIT 1
`

func (q *Queries) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	row := q.db.QueryRow(ctx, getBySlug, slug)
	var i models.Category
	err := row.Scan(
		&i.ID,
		&i.ParentID,
		&i.Name,
		&i.Slug,
		&i.Description,
		&i.MetaTitle,
		&i.MetaH1,
		&i.MetaDescription,
		&i.MetaKeyword,
		&i.IsEnable,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}
