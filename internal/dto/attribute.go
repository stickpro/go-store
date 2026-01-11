package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/delivery/http/request/attribute_request"
)

type CreateAttributeGroupDTO struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
}

type UpdateAttributeGroupDTO struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
}

func RequestToCreateAttributeGroupDTO(req *attribute_request.CreateAttributeGroupRequest) CreateAttributeGroupDTO {
	return CreateAttributeGroupDTO{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	}
}

func RequestToUpdateAttributeGroupDTO(req *attribute_request.UpdateAttributeGroupRequest) UpdateAttributeGroupDTO {
	return UpdateAttributeGroupDTO{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	}
}

// Attribute Definition DTOs (NEW STRUCTURE)
type CreateAttributeDTO struct {
	Name             string        `json:"name"`
	Slug             string        `json:"slug"`
	AttributeGroupID uuid.NullUUID `json:"attribute_group_id"`
	Type             string        `json:"type"` // select, number, boolean, text
	Unit             *string       `json:"unit,omitempty"`
	IsFilterable     bool          `json:"is_filterable"`
	IsVisible        bool          `json:"is_visible"`
	IsRequired       bool          `json:"is_required"`
	SortOrder        *int32        `json:"sort_order"`
}

type UpdateAttributeDTO struct {
	Name             string        `json:"name"`
	Slug             string        `json:"slug"`
	AttributeGroupID uuid.NullUUID `json:"attribute_group_id"`
	Type             string        `json:"type"`
	Unit             *string       `json:"unit,omitempty"`
	IsFilterable     bool          `json:"is_filterable"`
	IsVisible        bool          `json:"is_visible"`
	IsRequired       bool          `json:"is_required"`
	SortOrder        *int32        `json:"sort_order"`
}

// Attribute Value DTOs
type CreateAttributeValueDTO struct {
	AttributeID     uuid.UUID `json:"attribute_id"`
	Value           string    `json:"value"`
	ValueNormalized *string   `json:"value_normalized,omitempty"`
	ValueNumeric    *float64  `json:"value_numeric,omitempty"`
	DisplayOrder    int32     `json:"display_order"`
}

type UpdateAttributeValueDTO struct {
	Value           string   `json:"value"`
	ValueNormalized *string  `json:"value_normalized,omitempty"`
	ValueNumeric    *float64 `json:"value_numeric,omitempty"`
	DisplayOrder    int32    `json:"display_order"`
	IsActive        bool     `json:"is_active"`
}

type AttributeValueDTO struct {
	ID              uuid.UUID           `json:"id"`
	Value           string              `json:"value"`
	ValueNormalized *string             `json:"value_normalized,omitempty"`
	ValueNumeric    decimal.NullDecimal `json:"value_numeric,omitempty"`
	DisplayOrder    int32               `json:"display_order"`
	UsageCount      int64               `json:"usage_count,omitempty"`
}

// Product Attribute DTOs
type AssignProductAttributesDTO struct {
	ProductID         uuid.UUID   `json:"product_id"`
	AttributeValueIDs []uuid.UUID `json:"attribute_value_ids"`
}

type ProductAttributeDTO struct {
	AttributeID   uuid.UUID `json:"attribute_id"`
	AttributeName string    `json:"attribute_name"`
	AttributeSlug string    `json:"attribute_slug"`
	AttributeType string    `json:"attribute_type"`
	AttributeUnit *string   `json:"attribute_unit,omitempty"`
	GroupID       uuid.UUID `json:"group_id"`
	GroupName     string    `json:"group_name"`
	GroupSlug     string    `json:"group_slug"`
	ValueID       uuid.UUID `json:"value_id"`
	Value         string    `json:"value"`
	ValueNumeric  *float64  `json:"value_numeric,omitempty"`
}

// Filter DTOs (for frontend)
type AttributeGroupWithValuesDTO struct {
	GroupID          uuid.UUID                `json:"group_id"`
	GroupName        string                   `json:"group_name"`
	GroupSlug        string                   `json:"group_slug"`
	GroupDescription *string                  `json:"group_description,omitempty"`
	Attributes       []AttributeWithValuesDTO `json:"attributes"`
}

type AttributeWithValuesDTO struct {
	ID           uuid.UUID           `json:"id"`
	Name         string              `json:"name"`
	Slug         string              `json:"slug"`
	Type         string              `json:"type"`
	Unit         *string             `json:"unit,omitempty"`
	IsFilterable bool                `json:"is_filterable"`
	Values       []AttributeValueDTO `json:"values"`
}

// Facet Distribution DTO (для фильтров из Meilisearch)
type FacetDistributionDTO struct {
	AttributeSlug string           `json:"attribute_slug"`
	AttributeName string           `json:"attribute_name"`
	Values        map[string]int64 `json:"values"` // value -> count
}

func RequestToCreateAttributeDTO(req *attribute_request.CreateAttributeRequest) CreateAttributeDTO {
	var aGroupID uuid.NullUUID
	if req.AttributeGroupID != nil {
		aGroupID = uuid.NullUUID{UUID: *req.AttributeGroupID, Valid: true}
	}
	return CreateAttributeDTO{
		Name:             req.Name,
		Slug:             req.Slug,
		AttributeGroupID: aGroupID,
		Type:             req.Type,
		Unit:             req.Unit,
		IsFilterable:     req.IsFilterable,
		IsVisible:        req.IsVisible,
		IsRequired:       req.IsRequired,
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
		Slug:             req.Slug,
		AttributeGroupID: aGroupID,
		Type:             req.Type,
		Unit:             req.Unit,
		IsFilterable:     req.IsFilterable,
		IsVisible:        req.IsVisible,
		IsRequired:       req.IsRequired,
		SortOrder:        req.SortOrder,
	}
}
