// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/constant"
)

type Category struct {
	ID              uuid.UUID        `db:"id" json:"id"`
	ParentID        uuid.NullUUID    `db:"parent_id" json:"parent_id"`
	Name            string           `db:"name" json:"name"`
	Slug            string           `db:"slug" json:"slug"`
	Description     pgtype.Text      `db:"description" json:"description"`
	ImagePath       pgtype.Text      `db:"image_path" json:"image_path"`
	MetaTitle       pgtype.Text      `db:"meta_title" json:"meta_title"`
	MetaH1          pgtype.Text      `db:"meta_h1" json:"meta_h1"`
	MetaDescription pgtype.Text      `db:"meta_description" json:"meta_description"`
	MetaKeyword     pgtype.Text      `db:"meta_keyword" json:"meta_keyword"`
	IsEnable        bool             `db:"is_enable" json:"is_enable"`
	CreatedAt       pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt       pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type Medium struct {
	ID        uuid.UUID        `db:"id" json:"id"`
	Name      string           `db:"name" json:"name"`
	Path      string           `db:"path" json:"path"`
	FileName  string           `db:"file_name" json:"file_name"`
	MimeType  string           `db:"mime_type" json:"mime_type"`
	DiskType  string           `db:"disk_type" json:"disk_type"`
	Size      int64            `db:"size" json:"size"`
	CreatedAt pgtype.Timestamp `db:"created_at" json:"created_at"`
}

type PersonalAccessToken struct {
	ID            uuid.UUID        `db:"id" json:"id"`
	TokenableType string           `db:"tokenable_type" json:"tokenable_type"`
	TokenableID   uuid.UUID        `db:"tokenable_id" json:"tokenable_id"`
	Name          string           `db:"name" json:"name"`
	Token         string           `db:"token" json:"token"`
	LastUsedAt    pgtype.Timestamp `db:"last_used_at" json:"last_used_at"`
	ExpiresAt     *time.Time       `db:"expires_at" json:"expires_at"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}

type Product struct {
	ID              uuid.UUID            `db:"id" json:"id"`
	Name            string               `db:"name" json:"name"`
	Model           string               `db:"model" json:"model"`
	Slug            string               `db:"slug" json:"slug"`
	Description     pgtype.Text          `db:"description" json:"description"`
	MetaTitle       pgtype.Text          `db:"meta_title" json:"meta_title"`
	MetaH1          pgtype.Text          `db:"meta_h1" json:"meta_h1"`
	MetaDescription pgtype.Text          `db:"meta_description" json:"meta_description"`
	MetaKeyword     pgtype.Text          `db:"meta_keyword" json:"meta_keyword"`
	Sku             pgtype.Text          `db:"sku" json:"sku"`
	Upc             pgtype.Text          `db:"upc" json:"upc"`
	Ean             pgtype.Text          `db:"ean" json:"ean"`
	Jan             pgtype.Text          `db:"jan" json:"jan"`
	Isbn            pgtype.Text          `db:"isbn" json:"isbn"`
	Mpn             pgtype.Text          `db:"mpn" json:"mpn"`
	Location        pgtype.Text          `db:"location" json:"location"`
	Quantity        int64                `db:"quantity" json:"quantity"`
	StockStatus     constant.StockStatus `db:"stock_status" json:"stock_status"`
	Image           pgtype.Text          `db:"image" json:"image"`
	ManufacturerID  uuid.NullUUID        `db:"manufacturer_id" json:"manufacturer_id"`
	Price           decimal.Decimal      `db:"price" json:"price"`
	Weight          decimal.Decimal      `db:"weight" json:"weight"`
	Length          decimal.Decimal      `db:"length" json:"length"`
	Width           decimal.Decimal      `db:"width" json:"width"`
	Height          decimal.Decimal      `db:"height" json:"height"`
	Subtract        bool                 `db:"subtract" json:"subtract"`
	Minimum         int64                `db:"minimum" json:"minimum"`
	SortOrder       int32                `db:"sort_order" json:"sort_order"`
	IsEnable        bool                 `db:"is_enable" json:"is_enable"`
	Viewed          int64                `db:"viewed" json:"viewed"`
	CreatedAt       pgtype.Timestamp     `db:"created_at" json:"created_at"`
	UpdatedAt       pgtype.Timestamp     `db:"updated_at" json:"updated_at"`
}

type User struct {
	ID              uuid.UUID        `db:"id" json:"id"`
	Email           string           `db:"email" json:"email" validate:"required,email"`
	EmailVerifiedAt pgtype.Timestamp `db:"email_verified_at" json:"email_verified_at"`
	Password        string           `db:"password" json:"password" validate:"required,min=8,max=32"`
	RememberToken   pgtype.Text      `db:"remember_token" json:"remember_token"`
	Location        string           `db:"location" json:"location" validate:"required,timezone"`
	Language        string           `db:"language" json:"language"`
	CreatedAt       pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt       pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt       pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	IsAdmin         pgtype.Bool      `db:"is_admin" json:"is_admin"`
	Banned          pgtype.Bool      `db:"banned" json:"banned"`
}
