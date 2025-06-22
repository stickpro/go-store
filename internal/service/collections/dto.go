package collections

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/collection_request"
)

type CreateDTO struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Slug        string  `json:"slug"`
}

type GetDTO struct {
	Page     *uint32 `json:"page" query:"page"`
	PageSize *uint32 `json:"page_size" query:"page_size"`
}

type UpdateDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Slug        string    `json:"slug"`
}

func RequestToCreateDTO(req *collection_request.CreateCollectionRequest) CreateDTO {
	return CreateDTO{
		Name:        req.Name,
		Description: req.Description,
		Slug:        req.Slug,
	}
}

func RequestToUpdateDTO(req *collection_request.UpdateCollectionRequest, ID uuid.UUID) UpdateDTO {
	return UpdateDTO{
		ID:          ID,
		Name:        req.Name,
		Description: req.Description,
		Slug:        req.Slug,
	}
}
