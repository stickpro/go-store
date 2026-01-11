package attribute_response

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
)

type AttributeResponse struct {
	ID               uuid.UUID     `json:"id"`
	Name             string        `json:"name"`
	Slug             string        `json:"slug"`
	AttributeGroupID uuid.NullUUID `json:"attribute_group_id"`
	Type             string        `json:"type"`
	Unit             *string       `json:"unit,omitempty"`
	IsFilterable     bool          `json:"is_filterable"`
	IsVisible        bool          `json:"is_visible"`
	IsRequired       bool          `json:"is_required"`
	SortOrder        int32         `json:"sort_order,omitempty"`
	CreatedAt        string        `json:"created_at"`
	UpdatedAt        string        `json:"updated_at"`
} // @name AttributeResponse

func NewFromAttributeModel(attr *models.Attribute) AttributeResponse {
	var unit *string
	if attr.Unit.Valid {
		unit = &attr.Unit.String
	}

	return AttributeResponse{
		ID:               attr.ID,
		Name:             attr.Name,
		Slug:             attr.Slug,
		AttributeGroupID: attr.AttributeGroupID,
		Type:             attr.Type,
		Unit:             unit,
		IsFilterable:     attr.IsFilterable.Bool,
		IsVisible:        attr.IsVisible.Bool,
		IsRequired:       attr.IsRequired.Bool,
		SortOrder:        attr.SortOrder.Int32,
		CreatedAt:        attr.CreatedAt.Time.String(),
		UpdatedAt:        attr.UpdatedAt.Time.String(),
	}
}

type AttributeValueResponse struct {
	ID              uuid.UUID `json:"id"`
	AttributeID     uuid.UUID `json:"attribute_id"`
	Value           string    `json:"value"`
	ValueNormalized *string   `json:"value_normalized,omitempty"`
	ValueNumeric    *float64  `json:"value_numeric,omitempty"`
	DisplayOrder    int32     `json:"display_order"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
} // @name AttributeValueResponse

func NewFromAttributeValueModel(value *models.AttributeValue) AttributeValueResponse {
	var valueNormalized *string
	if value.ValueNormalized.Valid {
		valueNormalized = &value.ValueNormalized.String
	}

	var valueNumeric *float64
	if value.ValueNumeric.Valid {
		val, _ := value.ValueNumeric.Decimal.Float64()
		valueNumeric = &val
	}

	return AttributeValueResponse{
		ID:              value.ID,
		AttributeID:     value.AttributeID,
		Value:           value.Value,
		ValueNormalized: valueNormalized,
		ValueNumeric:    valueNumeric,
		DisplayOrder:    value.DisplayOrder.Int32,
		IsActive:        value.IsActive.Bool,
		CreatedAt:       value.CreatedAt.Time.String(),
		UpdatedAt:       value.UpdatedAt.Time.String(),
	}
}
