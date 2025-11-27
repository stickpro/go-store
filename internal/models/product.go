package models

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type ShortProduct struct {
	ID       uuid.UUID       `db:"id" json:"id"`
	Name     string          `db:"name" json:"name"`
	Model    string          `db:"model" json:"model"`
	Slug     string          `db:"slug" json:"slug"`
	Image    pgtype.Text     `db:"image" json:"image"`
	Price    decimal.Decimal `db:"price" json:"price"`
	IsEnable bool            `db:"is_enable" json:"is_enable"`
} // @name ShortProduct
