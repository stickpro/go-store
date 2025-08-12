package dto

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/collection_request"
	"github.com/stickpro/go-store/internal/models"
	"time"
)

type CreateCollectionDTO struct {
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	Slug        string      `json:"slug"`
	ProductIds  []uuid.UUID `json:"product_ids,omitempty"`
}

type GetCollectionDTO struct {
	Page     *uint32 `json:"page" query:"page"`
	PageSize *uint32 `json:"page_size" query:"page_size"`
}

type UpdateCollectionDTO struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	Slug        string      `json:"slug"`
	ProductIds  []uuid.UUID `json:"product_ids,omitempty"`
}

type WithProductsCollectionDTO struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Slug        string                 `json:"slug"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   *time.Time             `json:"updated_at"`
	Products    []*models.ShortProduct `json:"products,omitempty"`
}

func RequestToCreateCollectionDTO(req *collection_request.CreateCollectionRequest) CreateCollectionDTO {
	return CreateCollectionDTO{
		Name:        req.Name,
		Description: req.Description,
		Slug:        req.Slug,
		ProductIds:  req.ProductIDs,
	}
}

func RequestToUpdateCollectionDTO(req *collection_request.UpdateCollectionRequest, ID uuid.UUID) UpdateCollectionDTO {
	return UpdateCollectionDTO{
		ID:          ID,
		Name:        req.Name,
		Description: req.Description,
		Slug:        req.Slug,
		ProductIds:  req.ProductIDs,
	}
}
