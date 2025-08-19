package dto

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/attribute_request"
)

type CreateAttributeGroupDTO struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type UpdateAttributeGroupDTO struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func RequestToCreateAttributeGroupDTO(req *attribute_request.CreateAttributeGroupRequest) CreateAttributeGroupDTO {
	return CreateAttributeGroupDTO{
		Name:        req.Name,
		Description: req.Description,
	}
}

func RequestToUpdateAttributeGroupDTO(req *attribute_request.UpdateAttributeGroupRequest) UpdateAttributeGroupDTO {
	return UpdateAttributeGroupDTO{
		Name:        req.Name,
		Description: req.Description,
	}
}

type CreateAttributeDTO struct {
	Name             string        `json:"name"`
	Value            string        `json:"value"`
	AttributeGroupID uuid.NullUUID `json:"attribute_group_id"`
	Type             string        `json:"type"`
	IsFilterable     bool          `json:"is_filterable"`
	IsVisible        bool          `json:"is_visible"`
	SortOrder        *int32        `json:"sort_order"`
}

type UpdateAttributeDTO struct {
	Name             string        `json:"name"`
	Value            string        `json:"value"`
	AttributeGroupID uuid.NullUUID `json:"attribute_group_id"`
	Type             string        `json:"type"`
	IsFilterable     bool          `json:"is_filterable"`
	IsVisible        bool          `json:"is_visible"`
	SortOrder        *int32        `json:"sort_order"`
}

func RequestToCreateAttributeDTO(req *attribute_request.CreateAttributeRequest) CreateAttributeDTO {
	var aGroupID uuid.NullUUID
	if req.AttributeGroupID != nil {
		aGroupID = uuid.NullUUID{UUID: *req.AttributeGroupID, Valid: true}
	}
	return CreateAttributeDTO{
		Name:             req.Name,
		Value:            req.Value,
		AttributeGroupID: aGroupID,
		Type:             req.Type,
		IsFilterable:     req.IsFilterable,
		IsVisible:        req.IsVisible,
		SortOrder:        req.SortOrder,
	}
}

func RequestToUpdateAttributeDTO(req *attribute_request.UpdateAttributeRequest) UpdateAttributeDTO {
	var aGroupID uuid.NullUUID
	if req.AttributeGroupID != nil {
		aGroupID = uuid.NullUUID{UUID: *req.AttributeGroupID, Valid: true}
	}
	return UpdateAttributeDTO{
		Name:             req.Name,
		Value:            req.Value,
		AttributeGroupID: aGroupID,
		Type:             req.Type,
		IsFilterable:     req.IsFilterable,
		IsVisible:        req.IsVisible,
		SortOrder:        req.SortOrder,
	}
}
