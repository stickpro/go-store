package attribute

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/search"
	"github.com/stickpro/go-store/internal/service/search/searchtypes"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/internal/storage/repository/repository_attributes"
	"github.com/stickpro/go-store/internal/storage/repository/repository_product_attribute_values"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/logger"
)

type IAttributeService interface {
	IAttributeGroupService

	CreateAttribute(ctx context.Context, d dto.CreateAttributeDTO) (*models.Attribute, error)
	GetAttributes(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*models.Attribute], error)
	GetAttributeByID(ctx context.Context, id uuid.UUID) (*models.Attribute, error)
	UpdateAttribute(ctx context.Context, d dto.UpdateAttributeDTO, id uuid.UUID) (*models.Attribute, error)
	DeleteAttribute(ctx context.Context, id uuid.UUID) error
	SearchAttributes(ctx context.Context, q string, d dto.GetDTO) (*base.FindResponseWithFullPagination[*models.Attribute], error)

	// Attribute Values
	IAttributeValueService
	// Filters
	GetAvailableFiltersForCategory(ctx context.Context, categoryID *uuid.UUID) ([]*dto.AttributeGroupWithValuesDTO, error)

	// Indexing
	CreateAttributeIndex(ctx context.Context, reindex bool) error
	CreateAttributeGroupIndex(ctx context.Context, reindex bool) error
}

type Service struct {
	cfg           *config.Config
	logger        logger.Logger
	storage       storage.IStorage
	searchService searchtypes.ISearchService
}

func New(cfg *config.Config, logger logger.Logger, storage storage.IStorage, ss searchtypes.ISearchService) *Service {
	return &Service{
		cfg:           cfg,
		logger:        logger,
		storage:       storage,
		searchService: ss,
	}
}

func (s *Service) CreateAttribute(ctx context.Context, d dto.CreateAttributeDTO) (*models.Attribute, error) {
	params := repository_attributes.CreateParams{
		AttributeGroupID: d.AttributeGroupID,
		Name:             d.Name,
		Slug:             d.Slug,
		Type:             d.Type,
		Unit:             pgtypeutils.EncodeText(d.Unit),
		IsFilterable:     pgtypeutils.EncodeBool(&d.IsFilterable),
		IsVisible:        pgtypeutils.EncodeBool(&d.IsVisible),
		IsRequired:       pgtypeutils.EncodeBool(&d.IsRequired),
		SortOrder:        pgtypeutils.EncodeInt4(d.SortOrder),
	}

	attribute, err := s.storage.Attributes().Create(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to create attribute", "error", parsedErr)
		return nil, parsedErr
	}

	err = s.IndexAttribute(attribute)
	if err != nil {
		s.logger.Error("failed to index attribute", err)
		return nil, err
	}
	return attribute, nil
}

func (s *Service) GetAttributeByID(ctx context.Context, id uuid.UUID) (*models.Attribute, error) {
	attribute, err := s.storage.Attributes().GetByID(ctx, id)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get attribute by ID", "error", parsedErr)
		return nil, parsedErr
	}

	return attribute, nil
}

func (s *Service) GetAttributes(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*models.Attribute], error) {
	commonParams := *base.NewCommonFindParams()
	if d.PageSize != nil {
		commonParams.PageSize = d.PageSize
	}
	if d.Page != nil {
		commonParams.Page = d.Page
	}

	attributes, err := s.storage.Attributes().GetWithPaginate(ctx, commonParams)
	if err != nil {
		return nil, err
	}

	return attributes, nil
}

func (s *Service) UpdateAttribute(ctx context.Context, d dto.UpdateAttributeDTO, id uuid.UUID) (*models.Attribute, error) {
	params := repository_attributes.UpdateParams{
		ID:               id,
		AttributeGroupID: d.AttributeGroupID,
		Name:             d.Name,
		Slug:             d.Slug,
		Type:             d.Type,
		Unit:             pgtypeutils.EncodeText(d.Unit),
		IsFilterable:     pgtypeutils.EncodeBool(&d.IsFilterable),
		IsVisible:        pgtypeutils.EncodeBool(&d.IsVisible),
		IsRequired:       pgtypeutils.EncodeBool(&d.IsRequired),
		SortOrder:        pgtypeutils.EncodeInt4(d.SortOrder),
	}

	attribute, err := s.storage.Attributes().Update(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to update attribute", "error", parsedErr)
		return nil, parsedErr
	}
	err = s.IndexAttribute(attribute)
	if err != nil {
		s.logger.Error("failed to index attribute", err)
		return nil, err
	}
	return attribute, nil
}

func (s *Service) DeleteAttribute(ctx context.Context, id uuid.UUID) error {
	err := s.storage.Attributes().Delete(ctx, id)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to delete attribute", "error", parsedErr)
		return parsedErr
	}
	return nil
}

func (s *Service) SearchAttributes(ctx context.Context, q string, d dto.GetDTO) (*base.FindResponseWithFullPagination[*models.Attribute], error) {
	page := uint64(1)
	pageSize := uint64(10)

	if d.Page != nil && *d.Page > 0 {
		page = *d.Page
	}
	if d.PageSize != nil && *d.PageSize > 0 {
		pageSize = *d.PageSize
	}

	offset := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	searchResult, err := s.searchService.Search(constant.AttributesIndex, q, limit, offset)
	if err != nil {
		return nil, err
	}
	attributes, err := search.UnmarshalHits[*models.Attribute](searchResult.Hits)
	if err != nil {
		return nil, err
	}

	total := uint64(searchResult.TotalHits)
	lastPage := uint64(1)
	if pageSize > 0 {
		lastPage = (total + uint64(pageSize) - 1) / uint64(pageSize)
		if lastPage == 0 {
			lastPage = 1
		}
	}

	return &base.FindResponseWithFullPagination[*models.Attribute]{
		Items: attributes,
		Pagination: base.FullPagingData{
			Total:    total,
			PageSize: pageSize,
			Page:     page,
			LastPage: lastPage,
		},
	}, nil
}

// GetAvailableFiltersForCategory returns available filters for a category
func (s *Service) GetAvailableFiltersForCategory(ctx context.Context, categoryID *uuid.UUID) ([]*dto.AttributeGroupWithValuesDTO, error) {
	var rows []*repository_product_attribute_values.GetFiltersForCategoryRow
	var err error

	if categoryID == nil {
		// If no category specified, return all filterable attributes with values
		// This requires a different query - for now return error
		return nil, fmt.Errorf("category_id is required")
	}

	rows, err = s.storage.ProductAttributeValues().GetFiltersForCategory(ctx, uuid.NullUUID{UUID: *categoryID, Valid: true})
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get filters for category", "error", parsedErr)
		return nil, parsedErr
	}

	// Group results by attribute group and attribute
	groupMap := make(map[uuid.UUID]*dto.AttributeGroupWithValuesDTO)
	attrMap := make(map[uuid.UUID]*dto.AttributeWithValuesDTO)
	var groupOrder []uuid.UUID
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
				IsFilterable: true,
				Values:       []dto.AttributeValueDTO{},
			}
			attrMap[row.AttributeID] = attr
			attrOrder[row.GroupID] = append(attrOrder[row.GroupID], row.AttributeID)
		}

		value := dto.AttributeValueDTO{
			ID:              row.ValueID,
			Value:           row.Value,
			ValueNormalized: pgtypeutils.DecodeText(row.ValueNormalized),
			ValueNumeric:    row.ValueNumeric,
			UsageCount:      row.ProductCount,
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

	return result, nil
}
