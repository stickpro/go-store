package attribute

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/internal/storage/repository/repository_attribute_values"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

type IAttributeValueService interface {
	GetAttributeValues(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*models.AttributeValue], error)
	CreateAttributeValue(ctx context.Context, d dto.CreateAttributeValueDTO) (*models.AttributeValue, error)
	GetAttributeValuesByAttributeID(ctx context.Context, attributeID uuid.UUID) ([]*dto.AttributeValueDTO, error)
	UpdateAttributeValue(ctx context.Context, d dto.UpdateAttributeValueDTO, id uuid.UUID) (*models.AttributeValue, error)
	DeleteAttributeValue(ctx context.Context, id uuid.UUID) error
}

func (s *Service) GetAttributeValues(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*models.AttributeValue], error) {
	commonParams := *base.NewCommonFindParams()
	if d.PageSize != nil {
		commonParams.PageSize = d.PageSize
	}
	if d.Page != nil {
		commonParams.Page = d.Page
	}
	attributeValues, err := s.storage.AttributeValues().GetWithPaginate(ctx, commonParams)

	if err != nil {
		return nil, err
	}
	return attributeValues, nil
}

func (s *Service) CreateAttributeValue(ctx context.Context, d dto.CreateAttributeValueDTO) (*models.AttributeValue, error) {
	var valueNumeric decimal.NullDecimal
	if d.ValueNumeric != nil {
		valueNumeric = decimal.NullDecimal{Decimal: decimal.NewFromFloat(*d.ValueNumeric), Valid: true}
	}

	value, err := s.storage.AttributeValues().Create(ctx, repository_attribute_values.CreateParams{
		AttributeID:     d.AttributeID,
		Value:           d.Value,
		ValueNormalized: pgtypeutils.EncodeText(d.ValueNormalized),
		ValueNumeric:    valueNumeric,
		DisplayOrder:    pgtypeutils.EncodeInt4(&d.DisplayOrder),
		IsActive:        pgtypeutils.EncodeBool(&[]bool{true}[0]),
	})
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to create attribute value", "error", parsedErr)
		return nil, parsedErr
	}

	return value, nil
}

func (s *Service) GetAttributeValuesByAttributeID(ctx context.Context, attributeID uuid.UUID) ([]*dto.AttributeValueDTO, error) {
	rows, err := s.storage.AttributeValues().GetWithUsageCount(ctx, attributeID)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get attribute values", "error", parsedErr)
		return nil, parsedErr
	}

	values := make([]*dto.AttributeValueDTO, 0, len(rows))
	for _, row := range rows {
		values = append(values, &dto.AttributeValueDTO{
			ID:              row.ID,
			Value:           row.Value,
			ValueNormalized: pgtypeutils.DecodeText(row.ValueNormalized),
			ValueNumeric:    row.ValueNumeric,
			DisplayOrder:    row.DisplayOrder.Int32,
			UsageCount:      row.UsageCount,
		})
	}

	return values, nil
}

func (s *Service) UpdateAttributeValue(ctx context.Context, d dto.UpdateAttributeValueDTO, id uuid.UUID) (*models.AttributeValue, error) {
	var valueNumeric decimal.NullDecimal
	if d.ValueNumeric != nil {
		valueNumeric = decimal.NullDecimal{Decimal: decimal.NewFromFloat(*d.ValueNumeric), Valid: true}
	}

	value, err := s.storage.AttributeValues().Update(ctx, repository_attribute_values.UpdateParams{
		ID:              id,
		Value:           d.Value,
		ValueNormalized: pgtypeutils.EncodeText(d.ValueNormalized),
		ValueNumeric:    valueNumeric,
		DisplayOrder:    pgtypeutils.EncodeInt4(&d.DisplayOrder),
		IsActive:        pgtypeutils.EncodeBool(&d.IsActive),
	})
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to update attribute value", "error", parsedErr)
		return nil, parsedErr
	}

	return value, nil
}

func (s *Service) DeleteAttributeValue(ctx context.Context, id uuid.UUID) error {
	err := s.storage.AttributeValues().Delete(ctx, id)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to delete attribute value", "error", parsedErr)
		return parsedErr
	}

	return nil
}
