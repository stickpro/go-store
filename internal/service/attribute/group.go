package attribute

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/search"
	"github.com/stickpro/go-store/internal/storage/base"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_attribute_groups"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

type IAttributeGroupService interface {
	GetAttributeGroups(ctx context.Context, d dto.GetDTO) (*base.FindResponseWithFullPagination[*models.AttributeGroup], error)
	GetAttributeGroupByID(ctx context.Context, id uuid.UUID) (*models.AttributeGroup, error)
	CreateAttributeGroup(ctx context.Context, d dto.CreateAttributeGroupDTO) (*models.AttributeGroup, error)
	UpdateAttributeGroup(ctx context.Context, d dto.UpdateAttributeGroupDTO, id uuid.UUID) (*models.AttributeGroup, error)
	DeleteAttributeGroup(ctx context.Context, id uuid.UUID) error
	SearchAttributeGroup(ctx context.Context, q string, d dto.GetDTO) (*base.FindResponseWithFullPagination[*models.AttributeGroup], error)
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
		Slug:        d.Slug,
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
		Slug:        d.Slug,
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
		err := s.storage.AttributeGroups(repository.WithTx(tx)).Delete(ctx, id)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Error("failed to delete attribute group", "error", parsedErr)
			return parsedErr
		}
		err = s.storage.Attributes(repository.WithTx(tx)).DeleteByAttributeGroupID(ctx, id)
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

func (s *Service) SearchAttributeGroup(ctx context.Context, q string, d dto.GetDTO) (*base.FindResponseWithFullPagination[*models.AttributeGroup], error) {
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

	searchResult, err := s.searchService.Search(constant.AttributeGroupsIndex, q, limit, offset)
	if err != nil {
		return nil, err
	}
	attrGroup, err := search.UnmarshalHits[*models.AttributeGroup](searchResult.Hits)
	if err != nil {
		return nil, err
	}
	total := uint64(searchResult.TotalHits)
	lastPage := uint64(1)
	if pageSize > 0 {
		lastPage = (total + pageSize - 1) / pageSize
		if lastPage == 0 {
			lastPage = 1
		}
	}

	return &base.FindResponseWithFullPagination[*models.AttributeGroup]{
		Items: attrGroup,
		Pagination: base.FullPagingData{
			Total:    total,
			PageSize: pageSize,
			Page:     page,
			LastPage: lastPage,
		},
	}, nil
}
