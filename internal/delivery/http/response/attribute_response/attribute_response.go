package attribute_response

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/models"
)

type AttributeResponse struct {
	ID               uuid.UUID     `json:"id"`
	Name             string        `json:"name"`
	Value            string        `json:"value"`
	AttributeGroupID uuid.NullUUID `json:"attribute_group_id"`
	Type             string        `json:"type"`
	IsFilterable     bool          `json:"is_filterable"`
	IsVisible        bool          `json:"is_visible"`
	SortOrder        int32         `json:"sort_order,omitempty"`
	CreatedAt        string        `json:"created_at"`
	UpdatedAt        string        `json:"updated_at"`
}

func NewFromAttributeModel(attr *models.Attribute) AttributeResponse {
	return AttributeResponse{
		ID:               attr.ID,
		Name:             attr.Name,
		Value:            attr.Value,
		AttributeGroupID: attr.AttributeGroupID,
		Type:             attr.Type,
		IsFilterable:     attr.IsFilterable.Bool,
		IsVisible:        attr.IsVisible.Bool,
		SortOrder:        attr.SortOrder.Int32,
		CreatedAt:        attr.CreatedAt.Time.String(),
		UpdatedAt:        attr.UpdatedAt.Time.String(),
	}
}
