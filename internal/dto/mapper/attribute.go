package mapper

import (
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/storage/repository/repository_product_attribute_values"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

// MapProductAttributesToGroupedDTO maps product attribute rows to grouped structure
func MapProductAttributesToGroupedDTO(rows []*repository_product_attribute_values.GetByProductIDRow) []*dto.AttributeGroupWithValuesDTO {
	if len(rows) == 0 {
		return []*dto.AttributeGroupWithValuesDTO{}
	}

	// Maps to track groups and attributes
	groupMap := make(map[uuid.UUID]*dto.AttributeGroupWithValuesDTO)
	attrMap := make(map[uuid.UUID]*dto.AttributeWithValuesDTO)
	groupOrder := []uuid.UUID{}
	attrOrder := make(map[uuid.UUID][]uuid.UUID)

	for _, row := range rows {
		// Get or create group
		group, groupExists := groupMap[row.GroupID]
		if !groupExists {
			group = &dto.AttributeGroupWithValuesDTO{
				GroupID:   row.GroupID,
				GroupName: row.GroupName,
				GroupSlug: row.GroupSlug,
			}
			groupMap[row.GroupID] = group
			groupOrder = append(groupOrder, row.GroupID)
			attrOrder[row.GroupID] = []uuid.UUID{}
		}

		// Get or create attribute
		attr, attrExists := attrMap[row.AttributeID]
		if !attrExists {
			attr = &dto.AttributeWithValuesDTO{
				ID:           row.AttributeID,
				Name:         row.AttributeName,
				Slug:         row.AttributeSlug,
				Type:         row.AttributeType,
				Unit:         pgtypeutils.DecodeText(row.AttributeUnit),
				IsFilterable: row.IsFilterable.Bool,
				Values:       []dto.AttributeValueDTO{},
			}
			attrMap[row.AttributeID] = attr
			attrOrder[row.GroupID] = append(attrOrder[row.GroupID], row.AttributeID)
		}

		value := dto.AttributeValueDTO{
			ID:              row.AttributeValueID,
			Value:           row.AttributeValue,
			ValueNormalized: pgtypeutils.DecodeText(row.ValueNormalized),
			ValueNumeric:    row.ValueNumeric,
		}
		if row.ValueDisplayOrder.Valid {
			value.DisplayOrder = row.ValueDisplayOrder.Int32
		}

		attr.Values = append(attr.Values, value)
	}

	// Build final result maintaining order
	result := make([]*dto.AttributeGroupWithValuesDTO, 0, len(groupOrder))
	for _, groupID := range groupOrder {
		group := groupMap[groupID]

		// Add attributes in order
		for _, attrID := range attrOrder[groupID] {
			attr := attrMap[attrID]
			group.Attributes = append(group.Attributes, *attr)
		}

		result = append(result, group)
	}

	return result
}
