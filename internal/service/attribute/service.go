package attribute

import (
	"context"
	"github.com/stickpro/go-store/internal/config"
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
	GetAttributeGroup(ctx context.Context, dto GetDTO) (*base.FindResponseWithFullPagination[*models.AttributeGroup], error)
	CreateAttributeGroup(ctx context.Context, dto CreateGroupDTO) (*models.AttributeGroup, error)
	UpdateAttributeGroup(ctx context.Context, dto UpdateGroupDTO) (*models.AttributeGroup, error)
	CreateAttribute(ctx context.Context, dto CreateDTO) (*models.Attribute, error)
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

func (s *Service) GetAttributeGroup(ctx context.Context, dto GetDTO) (*base.FindResponseWithFullPagination[*models.AttributeGroup], error) {
	commonParams := base.NewCommonFindParams()
	if dto.PageSize != nil {
		commonParams.PageSize = dto.PageSize
	}
	if dto.Page != nil {
		commonParams.Page = dto.Page
	}

	attributeGroup, err := s.storage.AttributeGroups().GetWithPaginate(ctx, *commonParams)

	if err != nil {
		return nil, err
	}
	return attributeGroup, nil
}

func (s *Service) CreateAttributeGroup(ctx context.Context, dto CreateGroupDTO) (*models.AttributeGroup, error) {
	params := repository_attribute_groups.CreateParams{
		Name:        dto.Name,
		Description: pgtypeutils.EncodeText(dto.Description),
	}

	attributeGroup, err := s.storage.AttributeGroups().Create(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to create category", "error", parsedErr)
		return nil, parsedErr
	}
	return attributeGroup, nil
}

func (s *Service) UpdateAttributeGroup(ctx context.Context, dto UpdateGroupDTO) (*models.AttributeGroup, error) {
	params := repository_attribute_groups.UpdateParams{
		Name:        dto.Name,
		Description: pgtypeutils.EncodeText(dto.Description),
	}
	attributeGroup, err := s.storage.AttributeGroups().Update(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to update attribute group", "error", parsedErr)
		return nil, parsedErr
	}
	return attributeGroup, nil
}

func (s *Service) CreateAttribute(ctx context.Context, dto CreateDTO) (*models.Attribute, error) {
	params := repository_attributes.CreateParams{
		AttributeGroupID: dto.AttributeGroupID,
		Name:             dto.Name,
		Type:             dto.Type,
		IsFilterable:     pgtypeutils.EncodeBool(dto.IsFilterable),
		IsVisible:        pgtypeutils.EncodeBool(dto.IsVisible),
		SortOrder:        pgtypeutils.EncodeInt4(dto.SortOrder),
	}

	attribute, err := s.storage.Attributes().Create(ctx, params)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.logger.Error("failed to create attribute", "error", parsedErr)
		return nil, parsedErr
	}
	return attribute, nil
}
