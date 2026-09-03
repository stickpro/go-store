package product_response

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/dto"
)

// AttributeValueResponse is a single attribute value in a product's attribute listing.
type AttributeValueResponse struct {
	ID              uuid.UUID           `json:"id"`
	Value           string              `json:"value"`
	ValueNormalized *string             `json:"value_normalized,omitempty"`
	ValueNumeric    decimal.NullDecimal `json:"value_numeric,omitempty"`
	DisplayOrder    int32               `json:"display_order"`
} //	@name	AttributeValueResponse

// AttributeResponse is one attribute with its values.
type AttributeResponse struct {
	ID           uuid.UUID                `json:"id"`
	Name         string                   `json:"name"`
	Slug         string                   `json:"slug"`
	Type         string                   `json:"type"`
	Unit         *string                  `json:"unit,omitempty"`
	IsFilterable bool                     `json:"is_filterable"`
	Values       []AttributeValueResponse `json:"values"`
} //	@name	AttributeResponse

// AttributeGroupResponse is a group of attributes with their values.
type AttributeGroupResponse struct {
	GroupID          uuid.UUID           `json:"group_id"`
	GroupName        string              `json:"group_name"`
	GroupSlug        string              `json:"group_slug"`
	GroupDescription *string             `json:"group_description,omitempty"`
	Attributes       []AttributeResponse `json:"attributes"`
} //	@name	AttributeGroupResponse

func NewAttributeGroups(groups []*dto.AttributeGroupWithValuesDTO) []AttributeGroupResponse {
	out := make([]AttributeGroupResponse, 0, len(groups))
	for _, g := range groups {
		group := AttributeGroupResponse{
			GroupID:          g.GroupID,
			GroupName:        g.GroupName,
			GroupSlug:        g.GroupSlug,
			GroupDescription: g.GroupDescription,
			Attributes:       make([]AttributeResponse, 0, len(g.Attributes)),
		}
		for _, a := range g.Attributes {
			attr := AttributeResponse{
				ID:           a.ID,
				Name:         a.Name,
				Slug:         a.Slug,
				Type:         a.Type,
				Unit:         a.Unit,
				IsFilterable: a.IsFilterable,
				Values:       make([]AttributeValueResponse, 0, len(a.Values)),
			}
			for _, v := range a.Values {
				attr.Values = append(attr.Values, AttributeValueResponse{
					ID:              v.ID,
					Value:           v.Value,
					ValueNormalized: v.ValueNormalized,
					ValueNumeric:    v.ValueNumeric,
					DisplayOrder:    v.DisplayOrder,
				})
			}
			group.Attributes = append(group.Attributes, attr)
		}
		out = append(out, group)
	}
	return out
}
