package category

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/category_request"
)

type CreateDTO struct {
	Name            string        `json:"name"`
	ParentID        uuid.NullUUID `json:"parent_id"`
	Slug            string        `json:"slug"`
	Description     *string       `json:"description"`
	ImagePath       *string       `json:"image_path"`
	MetaTitle       *string       `json:"meta_title"`
	MetaH1          *string       `json:"meta_h1"`
	MetaDescription *string       `json:"meta_description"`
	MetaKeyword     *string       `json:"meta_keyword"`
	IsEnable        bool          `json:"is_enable"`
}

type UpdateDTO struct {
	ID              uuid.UUID     `json:"id"`
	Name            string        `json:"name"`
	ParentID        uuid.NullUUID `json:"parent_id"`
	Slug            string        `json:"slug"`
	Description     *string       `json:"description"`
	ImagePath       *string       `json:"image_path"`
	MetaTitle       *string       `json:"meta_title"`
	MetaH1          *string       `json:"meta_h1"`
	MetaDescription *string       `json:"meta_description"`
	MetaKeyword     *string       `json:"meta_keyword"`
	IsEnable        bool          `json:"is_enable"`
}

func RequestToCreateDTO(req *category_request.CreateCategoryRequest) CreateDTO {
	var parentID uuid.NullUUID
	if req.ParentID != nil {
		parentID = uuid.NullUUID{UUID: *req.ParentID, Valid: true}
	} else {
		parentID = uuid.NullUUID{Valid: false}
	}

	return CreateDTO{
		Name:            req.Name,
		ParentID:        parentID,
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

func RequestToUpdateDTO(req *category_request.UpdateCategoryRequest, id uuid.UUID) UpdateDTO {
	var parentID uuid.NullUUID
	if req.ParentID != nil {
		parentID = uuid.NullUUID{UUID: *req.ParentID, Valid: true}
	}

	return UpdateDTO{
		ID:              id,
		Name:            req.Name,
		ParentID:        parentID,
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
