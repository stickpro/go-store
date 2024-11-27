// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: users_gen.sql

package repository_users

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stickpro/go-store/internal/models"
)

const create = `-- name: Create :one
INSERT INTO users (email, email_verified_at, password, remember_token, location, language, created_at, deleted_at, banned)
	VALUES ($1, $2, $3, $4, $5, $6, now(), $7, $8)
	RETURNING id, email, email_verified_at, password, remember_token, location, language, created_at, updated_at, deleted_at, banned
`

type CreateParams struct {
	Email           string           `db:"email" json:"email" validate:"required,email"`
	EmailVerifiedAt pgtype.Timestamp `db:"email_verified_at" json:"email_verified_at"`
	Password        string           `db:"password" json:"password" validate:"required,min=8,max=32"`
	RememberToken   pgtype.Text      `db:"remember_token" json:"remember_token"`
	Location        string           `db:"location" json:"location" validate:"required,timezone"`
	Language        string           `db:"language" json:"language"`
	DeletedAt       pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	Banned          pgtype.Bool      `db:"banned" json:"banned"`
}

func (q *Queries) Create(ctx context.Context, arg CreateParams) (*models.User, error) {
	row := q.db.QueryRow(ctx, create,
		arg.Email,
		arg.EmailVerifiedAt,
		arg.Password,
		arg.RememberToken,
		arg.Location,
		arg.Language,
		arg.DeletedAt,
		arg.Banned,
	)
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
		&i.Banned,
	)
	return &i, err
}

const delete = `-- name: Delete :exec
DELETE FROM users WHERE id=$1
`

func (q *Queries) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, delete, id)
	return err
}

const getAll = `-- name: GetAll :many
SELECT id, email, email_verified_at, password, remember_token, location, language, created_at, updated_at, deleted_at, banned FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2
`

type GetAllParams struct {
	Limit  int32 `db:"limit" json:"limit"`
	Offset int32 `db:"offset" json:"offset"`
}

func (q *Queries) GetAll(ctx context.Context, arg GetAllParams) ([]*models.User, error) {
	rows, err := q.db.Query(ctx, getAll, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []*models.User{}
	for rows.Next() {
		var i models.User
		if err := rows.Scan(
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
			&i.Banned,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const update = `-- name: Update :one
UPDATE users
	SET location=$1, language=$2, updated_at=$3, banned=$4
	WHERE id=$5
	RETURNING id, email, email_verified_at, password, remember_token, location, language, created_at, updated_at, deleted_at, banned
`

type UpdateParams struct {
	Location  string           `db:"location" json:"location" validate:"required,timezone"`
	Language  string           `db:"language" json:"language"`
	UpdatedAt pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	Banned    pgtype.Bool      `db:"banned" json:"banned"`
	ID        uuid.UUID        `db:"id" json:"id"`
}

func (q *Queries) Update(ctx context.Context, arg UpdateParams) (*models.User, error) {
	row := q.db.QueryRow(ctx, update,
		arg.Location,
		arg.Language,
		arg.UpdatedAt,
		arg.Banned,
		arg.ID,
	)
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
		&i.Banned,
	)
	return &i, err
}
