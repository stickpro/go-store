package product

import (
	"context"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/tools"
	utils "github.com/stickpro/go-store/pkg/util"
)

func (s *Service) CreateProductIndex(ctx context.Context, reindex bool) error {
	shouldCreate := reindex

	if !reindex {
		exists, err := s.searchService.CheckIndex(constant.ProductsIndex)
		if err != nil {
			s.logger.Error("Failed to check index", "error", err)
			return err
		}
		shouldCreate = !exists
	}

	if !shouldCreate {
		return nil
	}

	page := uint32(1)
	pageSize := uint32(100)

	for {
		dto := GetDTO{
			Page:     tools.Pointer(page),
			PageSize: tools.Pointer(pageSize),
		}

		res, err := s.GetProductWithPagination(ctx, dto)
		if err != nil {
			s.logger.Error("Failed to get product page", "page", page, "error", err)
			return err
		}

		if len(res.Items) == 0 {
			break
		}

		data, err := utils.StructToMap(res.Items)
		if err != nil {
			s.logger.Error("Failed to convert product page to map", "error", err)
			return err
		}

		err = s.searchService.CreateIndex(constant.ProductsIndex, data)
		if err != nil {
			s.logger.Error("Failed to index product batch ", "page ", page, "error ", err)
			return err
		}

		s.logger.Debug("Indexed product batch ", "page", page, "count ", len(res.Items))

		if uint64(page) >= res.Pagination.LastPage {
			break
		}

		page++
	}

	s.logger.Info("Product index created successfully")

	return nil
}
