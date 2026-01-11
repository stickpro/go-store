package attribute

import (
	"context"

	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/tools"
	utils "github.com/stickpro/go-store/pkg/util"
)

func (s *Service) CreateAttributeIndex(ctx context.Context, reindex bool) error {
	shouldCreate := reindex

	if !reindex {
		exists, err := s.searchService.CheckIndex(constant.AttributesIndex)
		if err != nil {
			s.logger.Error("Failed to check index", "error", err)
			return err
		}
		shouldCreate = !exists
	}

	if !shouldCreate {
		return nil
	}

	page := uint64(1)
	pageSize := uint64(100)

	for {
		d := dto.GetDTO{
			Page:     tools.Pointer(page),
			PageSize: tools.Pointer(pageSize),
		}

		res, err := s.GetAttributes(ctx, d)
		if err != nil {
			s.logger.Error("Failed to get attributes page", "page", page, "error", err)
			return err
		}

		if len(res.Items) == 0 {
			break
		}
		data, err := utils.StructToMap(res.Items)
		if err != nil {
			s.logger.Error("Failed to index attributes batch ", "page ", page, "error ", err)
			return err
		}

		err = s.searchService.CreateIndex(constant.AttributesIndex, data)
		if err != nil {
			s.logger.Error("Failed to index attributes batch ", "page ", page, "error ", err)
			return err
		}

		s.logger.Debug("Indexed attributes batch ", "page", page, "count ", len(res.Items))

		if uint64(page) >= res.Pagination.LastPage {
			break
		}

		page++
	}

	s.logger.Info("Finished indexing attributes")
	return nil
}

func (s *Service) CreateAttributeGroupIndex(ctx context.Context, reindex bool) error {
	shouldCreate := reindex

	if !reindex {
		exists, err := s.searchService.CheckIndex(constant.AttributeGroupsIndex)
		if err != nil {
			s.logger.Error("Failed to check index", "error", err)
			return err
		}
		shouldCreate = !exists
	}

	if !shouldCreate {
		return nil
	}

	page := uint64(1)
	pageSize := uint64(100)

	for {
		d := dto.GetDTO{
			Page:     tools.Pointer(page),
			PageSize: tools.Pointer(pageSize),
		}

		res, err := s.GetAttributeGroups(ctx, d)
		if err != nil {
			s.logger.Error("Failed to get attribute groups page", "page", page, "error", err)
			return err
		}

		if len(res.Items) == 0 {
			break
		}
		data, err := utils.StructToMap(res.Items)
		if err != nil {
			s.logger.Error("Failed to index attribute groups batch ", "page ", page, "error ", err)
			return err
		}

		err = s.searchService.CreateIndex(constant.AttributeGroupsIndex, data)
		if err != nil {
			s.logger.Error("Failed to index attribute groups batch ", "page ", page, "error ", err)
			return err
		}

		s.logger.Debug("Indexed attribute groups batch ", "page", page, "count ", len(res.Items))

		if page >= res.Pagination.LastPage {
			break
		}

		page++
	}

	s.logger.Info("Finished indexing attribute groups")
	return nil
}

func (s *Service) IndexAttribute(attribute *models.Attribute) error {
	data, err := utils.StructToMap([]*models.Attribute{attribute})
	if err != nil {
		return err
	}

	return s.searchService.UpsertDocument(constant.AttributesIndex, data)
}

func (s *Service) IndexAttributeGroup(attribute *models.AttributeGroup) error {
	data, err := utils.StructToMap([]*models.AttributeGroup{attribute})
	if err != nil {
		return err
	}

	return s.searchService.UpsertDocument(constant.AttributeGroupsIndex, data)
}
