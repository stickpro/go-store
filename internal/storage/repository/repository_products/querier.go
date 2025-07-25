// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package repository_products

import (
	"context"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
)

type Querier interface {
	Create(ctx context.Context, arg CreateParams) (*models.Product, error)
	CreateProductMedia(ctx context.Context, arg CreateProductMediaParams) error
	DeleteProductMedia(ctx context.Context, productID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetBySlug(ctx context.Context, slug string) (*models.Product, error)
	GetMediaByProductID(ctx context.Context, productID uuid.UUID) ([]*models.Medium, error)
	Update(ctx context.Context, arg UpdateParams) (*models.Product, error)
}

var _ Querier = (*Queries)(nil)
