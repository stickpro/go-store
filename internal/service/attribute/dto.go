package attribute

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/attribute_request"
)

type GetDTO struct {
	Page     *uint32 `json:"page" query:"page"`
	PageSize *uint32 `json:"page_size" query:"page_size"`
}

type CreateGroupDTO struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type UpdateGroupDTO struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func RequestToCreateGroupDTO(req *attribute_request.CreateAttributeGroupRequest) CreateGroupDTO {
	return CreateGroupDTO{
		Name:        req.Name,
		Description: req.Description,
	}
}

func RequestToUpdateGroupDTO(req *attribute_request.UpdateAttributeGroupRequest) UpdateGroupDTO {
	return UpdateGroupDTO{
		Name:        req.Name,
		Description: req.Description,
	}
}

type CreateDTO struct {
	Name             string        `json:"name"`
	AttributeGroupID uuid.NullUUID `json:"attribute_group_id"`
	Type             string        `json:"type"`
	IsFilterable     *bool         `json:"is_filterable"`
	IsVisible        *bool         `json:"is_visible"`
	SortOrder        *int32        `json:"sort_order"`
}

type UpdateDTO struct {
	Name             string        `json:"name"`
	AttributeGroupID uuid.NullUUID `json:"attribute_group_id"`
	Type             string        `json:"type"`
	IsFilterable     *bool         `json:"is_filterable"`
	IsVisible        *bool         `json:"is_visible"`
	SortOrder        *int32        `json:"sort_order"`
}
