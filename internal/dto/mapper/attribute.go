package mapper

import (
	"github.com/goccy/go-json"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/storage/repository/repository_attributes"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

func MapProductAttributesToDTO(rows []*repository_attributes.GetAttributesProductRow) ([]*dto.AttributeGroupDTO, error) {
	result := make([]*dto.AttributeGroupDTO, 0, len(rows))

	for _, r := range rows {
		var attrs []dto.AttributeDTO
		if err := json.Unmarshal(r.Attributes, &attrs); err != nil {
			return nil, err
		}

		group := dto.AttributeGroupDTO{
			GroupID:          r.GroupID,
			GroupName:        r.GroupName,
			GroupDescription: pgtypeutils.DecodeText(r.GroupDescription),
			Attributes:       attrs,
		}

		result = append(result, &group)
	}

	return result, nil
}
