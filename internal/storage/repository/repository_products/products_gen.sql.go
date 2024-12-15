// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: products_gen.sql

package repository_products

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/models"
)

const create = `-- name: Create :one
INSERT INTO products (name, model, slug, description, meta_title, meta_h1, meta_description, meta_keyword, sku, upc, ean, jan, isbn, mpn, location, quantity, stock_status, image, manufacturer_id, price, weight, length, width, height, subtract, minimum, sort_order, is_enable, viewed, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, now())
	RETURNING id, name, model, slug, description, meta_title, meta_h1, meta_description, meta_keyword, sku, upc, ean, jan, isbn, mpn, location, quantity, stock_status, image, manufacturer_id, price, weight, length, width, height, subtract, minimum, sort_order, is_enable, viewed, created_at, updated_at
`

type CreateParams struct {
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
}

func (q *Queries) Create(ctx context.Context, arg CreateParams) (*models.Product, error) {
	row := q.db.QueryRow(ctx, create,
		arg.Name,
		arg.Model,
		arg.Slug,
		arg.Description,
		arg.MetaTitle,
		arg.MetaH1,
		arg.MetaDescription,
		arg.MetaKeyword,
		arg.Sku,
		arg.Upc,
		arg.Ean,
		arg.Jan,
		arg.Isbn,
		arg.Mpn,
		arg.Location,
		arg.Quantity,
		arg.StockStatus,
		arg.Image,
		arg.ManufacturerID,
		arg.Price,
		arg.Weight,
		arg.Length,
		arg.Width,
		arg.Height,
		arg.Subtract,
		arg.Minimum,
		arg.SortOrder,
		arg.IsEnable,
		arg.Viewed,
	)
	var i models.Product
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Model,
		&i.Slug,
		&i.Description,
		&i.MetaTitle,
		&i.MetaH1,
		&i.MetaDescription,
		&i.MetaKeyword,
		&i.Sku,
		&i.Upc,
		&i.Ean,
		&i.Jan,
		&i.Isbn,
		&i.Mpn,
		&i.Location,
		&i.Quantity,
		&i.StockStatus,
		&i.Image,
		&i.ManufacturerID,
		&i.Price,
		&i.Weight,
		&i.Length,
		&i.Width,
		&i.Height,
		&i.Subtract,
		&i.Minimum,
		&i.SortOrder,
		&i.IsEnable,
		&i.Viewed,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}

const update = `-- name: Update :one
UPDATE products
	SET name=$1, model=$2, slug=$3, description=$4, meta_title=$5, meta_h1=$6, 
		meta_description=$7, meta_keyword=$8, sku=$9, upc=$10, ean=$11, jan=$12, 
		isbn=$13, mpn=$14, location=$15, quantity=$16, stock_status=$17, image=$18, 
		manufacturer_id=$19, price=$20, weight=$21, length=$22, width=$23, height=$24, 
		subtract=$25, minimum=$26, sort_order=$27, is_enable=$28, viewed=$29, updated_at=now()
	WHERE id=$30
	RETURNING id, name, model, slug, description, meta_title, meta_h1, meta_description, meta_keyword, sku, upc, ean, jan, isbn, mpn, location, quantity, stock_status, image, manufacturer_id, price, weight, length, width, height, subtract, minimum, sort_order, is_enable, viewed, created_at, updated_at
`

type UpdateParams struct {
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
	ID              uuid.UUID            `db:"id" json:"id"`
}

func (q *Queries) Update(ctx context.Context, arg UpdateParams) (*models.Product, error) {
	row := q.db.QueryRow(ctx, update,
		arg.Name,
		arg.Model,
		arg.Slug,
		arg.Description,
		arg.MetaTitle,
		arg.MetaH1,
		arg.MetaDescription,
		arg.MetaKeyword,
		arg.Sku,
		arg.Upc,
		arg.Ean,
		arg.Jan,
		arg.Isbn,
		arg.Mpn,
		arg.Location,
		arg.Quantity,
		arg.StockStatus,
		arg.Image,
		arg.ManufacturerID,
		arg.Price,
		arg.Weight,
		arg.Length,
		arg.Width,
		arg.Height,
		arg.Subtract,
		arg.Minimum,
		arg.SortOrder,
		arg.IsEnable,
		arg.Viewed,
		arg.ID,
	)
	var i models.Product
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Model,
		&i.Slug,
		&i.Description,
		&i.MetaTitle,
		&i.MetaH1,
		&i.MetaDescription,
		&i.MetaKeyword,
		&i.Sku,
		&i.Upc,
		&i.Ean,
		&i.Jan,
		&i.Isbn,
		&i.Mpn,
		&i.Location,
		&i.Quantity,
		&i.StockStatus,
		&i.Image,
		&i.ManufacturerID,
		&i.Price,
		&i.Weight,
		&i.Length,
		&i.Width,
		&i.Height,
		&i.Subtract,
		&i.Minimum,
		&i.SortOrder,
		&i.IsEnable,
		&i.Viewed,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return &i, err
}
