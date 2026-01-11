package manufacturer

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/manufacturer_request"
)

type CreateDTO struct {
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	Description     *string `json:"description"`
	ImagePath       *string `json:"image_path"`
	MetaTitle       *string `json:"meta_title"`
	MetaH1          *string `json:"meta_h1"`
	MetaDescription *string `json:"meta_description"`
	MetaKeyword     *string `json:"meta_keyword"`
	IsEnable        bool    `json:"is_enable"`
}
type UpdateDTO struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     *string   `json:"description"`
	ImagePath       *string   `json:"image_path"`
	MetaTitle       *string   `json:"meta_title"`
	MetaH1          *string   `json:"meta_h1"`
	MetaDescription *string   `json:"meta_description"`
	MetaKeyword     *string   `json:"meta_keyword"`
	IsEnable        bool      `json:"is_enable"`
}

type GetDTO struct {
	Page     *uint64 `json:"page" query:"page"`
	PageSize *uint64 `json:"page_size" query:"page_size"`
}

func RequestToCreateDTO(req *manufacturer_request.CreateManufacturerRequest) CreateDTO {
	return CreateDTO{
		Name:            req.Name,
		Slug:            req.Slug,
		Description:     req.Description,
		ImagePath:       req.ImagePath,
		MetaTitle:       req.MetaTitle,
		MetaH1:          req.MetaH1,
		MetaDescription: req.MetaDescription,
		MetaKeyword:     req.MetaKeyword,
		IsEnable:        req.IsEnabled,
	}
}

func RequestToUpdateDTO(req *manufacturer_request.UpdateManufacturerRequest, id uuid.UUID) UpdateDTO {
	return UpdateDTO{
		ID:              id,
		Name:            req.Name,
		Slug:            req.Slug,
		Description:     req.Description,
		ImagePath:       req.ImagePath,
		MetaTitle:       req.MetaTitle,
		MetaH1:          req.MetaH1,
		MetaDescription: req.MetaDescription,
		MetaKeyword:     req.MetaKeyword,
		IsEnable:        req.IsEnabled,
	}
}
