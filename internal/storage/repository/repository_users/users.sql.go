// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: users.sql

package repository_users

import (
	"context"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
)

const getByEmail = `-- name: GetByEmail :one
SELECT id, email, email_verified_at, password, remember_token, location, language, created_at, updated_at, deleted_at, is_admin, banned
FROM users
WHERE email = $1
LIMIT 1
`

func (q *Queries) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	row := q.db.QueryRow(ctx, getByEmail, email)
	var i models.User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.EmailVerifiedAt,
		&i.Password,
		&i.RememberToken,
		&i.Location,
		&i.Language,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.IsAdmin,
		&i.Banned,
	)
	return &i, err
}

const getByID = `-- name: GetByID :one
SELECT id, email, email_verified_at, password, remember_token, location, language, created_at, updated_at, deleted_at, is_admin, banned
FROM users
WHERE id = $1
LIMIT 1
`

func (q *Queries) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	row := q.db.QueryRow(ctx, getByID, id)
	var i models.User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.EmailVerifiedAt,
		&i.Password,
		&i.RememberToken,
		&i.Location,
		&i.Language,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.IsAdmin,
		&i.Banned,
	)
	return &i, err
}
