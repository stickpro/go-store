package category

import (
	"context"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/storage/repository"
	"github.com/stickpro/go-store/internal/storage/repository/repository_category_paths"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
)

type IBreadcrumb interface {
	GetBreadcrumbsByProductSlug(ctx context.Context, productSlug string) ([]*dto.BreadcrumbDTO, error)
	GetBreadcrumbsByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]*dto.BreadcrumbDTO, error)
	GetBreadcrumbsByCategorySlug(ctx context.Context, categorySlug string) ([]*dto.BreadcrumbDTO, error)
	GetDirectChildren(ctx context.Context, categoryID uuid.UUID) ([]*dto.CategoryChildDTO, error)
	RebuildCategoryPaths(ctx context.Context, categoryID uuid.UUID) error
}

// GetBreadcrumbsByProductSlug returns the full breadcrumb path for a product by its slug
// Example: Catalog -> Office -> Print Tech -> Scanners
func (s *Service) GetBreadcrumbsByProductSlug(ctx context.Context, productSlug string) ([]*dto.BreadcrumbDTO, error) {
	rows, err := s.storage.CategoryPaths().GetBreadcrumbsByProductSlug(ctx, productSlug)
	if err != nil {
		s.logger.Error("failed to get breadcrumbs by product slug", "slug", productSlug, "error", err)
		return nil, err
	}

	breadcrumbs := make([]*dto.BreadcrumbDTO, 0, len(rows))
	for _, row := range rows {
		breadcrumbs = append(breadcrumbs, &dto.BreadcrumbDTO{
			ID:        row.ID,
			Name:      row.Name,
			Slug:      row.Slug,
			MetaTitle: pgtypeutils.DecodeText(row.MetaTitle),
			MetaH1:    pgtypeutils.DecodeText(row.MetaH1),
			Depth:     row.Depth,
		})
	}

	return breadcrumbs, nil
}

// GetBreadcrumbsByCategoryID returns the full breadcrumb path for a category by its ID
// Example: Catalog -> Office -> Print Tech -> Scanners
func (s *Service) GetBreadcrumbsByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]*dto.BreadcrumbDTO, error) {
	rows, err := s.storage.CategoryPaths().GetBreadcrumbsByCategoryID(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to get breadcrumbs by category ID", "id", categoryID, "error", err)
		return nil, err
	}

	breadcrumbs := make([]*dto.BreadcrumbDTO, 0, len(rows))
	for _, row := range rows {
		breadcrumbs = append(breadcrumbs, &dto.BreadcrumbDTO{
			ID:        row.ID,
			Name:      row.Name,
			Slug:      row.Slug,
			MetaTitle: pgtypeutils.DecodeText(row.MetaTitle),
			MetaH1:    pgtypeutils.DecodeText(row.MetaH1),
			Depth:     row.Depth,
		})
	}

	return breadcrumbs, nil
}

// GetBreadcrumbsByCategorySlug returns the full breadcrumb path for a category by its slug
// Example: Catalog -> Office -> Print Tech -> Scanners
func (s *Service) GetBreadcrumbsByCategorySlug(ctx context.Context, categorySlug string) ([]*dto.BreadcrumbDTO, error) {
	rows, err := s.storage.CategoryPaths().GetBreadcrumbsByCategorySlug(ctx, categorySlug)
	if err != nil {
		s.logger.Error("failed to get breadcrumbs by category slug", "slug", categorySlug, "error", err)
		return nil, err
	}

	breadcrumbs := make([]*dto.BreadcrumbDTO, 0, len(rows))
	for _, row := range rows {
		breadcrumbs = append(breadcrumbs, &dto.BreadcrumbDTO{
			ID:        row.ID,
			Name:      row.Name,
			Slug:      row.Slug,
			MetaTitle: pgtypeutils.DecodeText(row.MetaTitle),
			MetaH1:    pgtypeutils.DecodeText(row.MetaH1),
			Depth:     row.Depth,
		})
	}

	return breadcrumbs, nil
}

// GetDirectChildren returns direct child categories (depth = 1)
// Useful for displaying subcategories in a category page
func (s *Service) GetDirectChildren(ctx context.Context, categoryID uuid.UUID) ([]*dto.CategoryChildDTO, error) {
	rows, err := s.storage.CategoryPaths().GetDirectChildren(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to get direct children", "id", categoryID, "error", err)
		return nil, err
	}

	children := make([]*dto.CategoryChildDTO, 0, len(rows))
	for _, row := range rows {
		children = append(children, &dto.CategoryChildDTO{
			ID:       row.ID,
			ParentID: row.ParentID,
			Name:     row.Name,
			Slug:     row.Slug,
			IsEnable: row.IsEnable,
		})
	}

	return children, nil
}

func (s *Service) RebuildCategoryPaths(ctx context.Context, categoryID uuid.UUID) error {
	return s.rebuildCategoryPathsWithOpts(ctx, categoryID)
}

func (s *Service) rebuildCategoryPathsWithOpts(ctx context.Context, categoryID uuid.UUID, opts ...repository.Option) error {
	category, err := s.GetCategoryByID(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to get category for rebuild", "id", categoryID, "error", err)
		return err
	}

	err = s.storage.CategoryPaths(opts...).DeleteCategoryPaths(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to delete old category paths", "id", categoryID, "error", err)
		return err
	}

	err = s.storage.CategoryPaths(opts...).InsertCategoryPath(ctx, repository_category_paths.InsertCategoryPathParams{
		AncestorID:   categoryID,
		DescendantID: categoryID,
		Depth:        0,
	})
	if err != nil {
		s.logger.Error("failed to insert self-reference", "id", categoryID, "error", err)
		return err
	}

	if category.ParentID.Valid {
		ancestors, err := s.storage.CategoryPaths(opts...).GetAllAncestors(ctx, category.ParentID.UUID)
		if err != nil {
			s.logger.Error("failed to get ancestors", "parent_id", category.ParentID.UUID, "error", err)
			return err
		}

		for _, ancestor := range ancestors {
			err = s.storage.CategoryPaths(opts...).InsertCategoryPath(ctx, repository_category_paths.InsertCategoryPathParams{
				AncestorID:   ancestor.ID,
				DescendantID: categoryID,
				Depth:        ancestor.Depth + 1,
			})
			if err != nil {
				s.logger.Error("failed to insert ancestor path", "ancestor", ancestor.ID, "category", categoryID, "error", err)
				return err
			}
		}
	}

	children, err := s.storage.CategoryPaths(opts...).GetDirectChildren(ctx, categoryID)
	if err != nil {
		s.logger.Error("failed to get children for rebuild", "id", categoryID, "error", err)
		return err
	}

	for _, child := range children {
		err = s.rebuildCategoryPathsWithOpts(ctx, child.ID, opts...)
		if err != nil {
			s.logger.Error("failed to rebuild child paths", "child_id", child.ID, "error", err)
			return err
		}
	}

	s.logger.Info("successfully rebuilt category paths", "category_id", categoryID)
	return nil
}
