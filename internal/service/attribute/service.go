package attribute

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/internal/storage/repository/repository_attribute_groups"
	"github.com/stickpro/go-store/internal/storage/repository/repository_attributes"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/logger"
)

type IAttributeService interface {
	GetAttributeGroups(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*models.AttributeGroup], error)
	GetAttributeGroupByID(ctx context.Context, id uuid.UUID) (*models.AttributeGroup, error)
	CreateAttributeGroup(ctx context.Context, d dto.CreateAttributeGroupDTO) (*models.AttributeGroup, error)
	UpdateAttributeGroup(ctx context.Context, d dto.UpdateAttributeGroupDTO, id uuid.UUID) (*models.AttributeGroup, error)
	DeleteAttributeGroup(ctx context.Context, id uuid.UUID) error

	CreateAttribute(ctx context.Context, d dto.CreateAttributeDTO) (*models.Attribute, error)
	GetAttributes(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*models.Attribute], error)
	GetAttributeByID(ctx context.Context, id uuid.UUID) (*models.Attribute, error)
	UpdateAttribute(ctx context.Context, d dto.UpdateAttributeDTO, id uuid.UUID) (*models.Attribute, error)
	DeleteAttribute(ctx context.Context, id uuid.UUID) error
}

type Service struct {
	cfg     *config.Config
	logger  logger.Logger
	storage storage.IStorage
}

func New(cfg *config.Config, logger logger.Logger, storage storage.IStorage) *Service {
	return &Service{
		cfg:     cfg,
		logger:  logger,
		storage: storage,
	}
}

func (s *Service) GetAttributeGroups(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*models.AttributeGroup], error) {
	commonParams := *base.NewCommonFindParams()
	if d.PageSize != nil {
		commonParams.PageSize = d.PageSize
	}
	if d.Page != nil {
		commonParams.Page = d.Page
	}

	attributeGroup, err := s.storage.AttributeGroups().GetWithPaginate(ctx, commonParams)

	if err != nil {
		return nil, err
	}
	return attributeGroup, nil
}

func (s *Service) GetAttributeGroupByID(ctx context.Context, id uuid.UUID) (*models.AttributeGroup, error) {
	attributeGroup, err := s.storage.AttributeGroups().GetByID(ctx, id)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to get attribute group by ID", "error", parsedErr)
		return nil, parsedErr
	}
	return attributeGroup, nil
}

func (s *Service) CreateAttributeGroup(ctx context.Context, d dto.CreateAttributeGroupDTO) (*models.AttributeGroup, error) {
	params := repository_attribute_groups.CreateParams{
		Name:        d.Name,
		Description: pgtypeutils.EncodeText(d.Description),
	}

	attributeGroup, err := s.storage.AttributeGroups().Create(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to create category", "error", parsedErr)
		return nil, parsedErr
	}
	return attributeGroup, nil
}

func (s *Service) UpdateAttributeGroup(ctx context.Context, d dto.UpdateAttributeGroupDTO, id uuid.UUID) (*models.AttributeGroup, error) {
	params := repository_attribute_groups.UpdateParams{
		Name:        d.Name,
		Description: pgtypeutils.EncodeText(d.Description),
		ID:          id,
	}
	attributeGroup, err := s.storage.AttributeGroups().Update(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to update attribute group", "error", parsedErr)
		return nil, parsedErr
	}
	return attributeGroup, nil
}

func (s *Service) DeleteAttributeGroup(ctx context.Context, id uuid.UUID) error {
	err := pgx.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		err := s.storage.AttributeGroups().Delete(ctx, id)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to delete attribute group", "error", parsedErr)
			return parsedErr
		}
		err = s.storage.Attributes().DeleteByAttributeGroupID(ctx, id)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to delete attributes by attribute group ID", "error", parsedErr)
			return parsedErr
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) CreateAttribute(ctx context.Context, d dto.CreateAttributeDTO) (*models.Attribute, error) {
	params := repository_attributes.CreateParams{
		AttributeGroupID: d.AttributeGroupID,
		Name:             d.Name,
		Type:             d.Type,
		IsFilterable:     pgtypeutils.EncodeBool(&d.IsFilterable),
		IsVisible:        pgtypeutils.EncodeBool(&d.IsVisible),
		SortOrder:        pgtypeutils.EncodeInt4(d.SortOrder),
	}

	attribute, err := s.storage.Attributes().Create(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to create attribute", "error", parsedErr)
		return nil, parsedErr
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
		Type:             d.Type,
		IsFilterable:     pgtypeutils.EncodeBool(&d.IsFilterable),
		IsVisible:        pgtypeutils.EncodeBool(&d.IsVisible),
		SortOrder:        pgtypeutils.EncodeInt4(d.SortOrder),
	}

	attribute, err := s.storage.Attributes().Update(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to update attribute", "error", parsedErr)
		return nil, parsedErr
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
