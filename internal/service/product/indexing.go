package product

import (
	"context"

	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/search/searchtypes"
	"github.com/stickpro/go-store/internal/storage/repository/repository_product_attribute_values"
	"github.com/stickpro/go-store/internal/tools"
)

func (s *Service) CreateProductVariantIndex(ctx context.Context, reindex bool) error {
	shouldCreate := reindex

	if !reindex {
		exists, err := s.searchService.CheckIndex(constant.ProductVariantsIndex)
		if err != nil {
			s.logger.Error("Failed to check variant index", "error", err)
			return err
		}
		shouldCreate = !exists
	}

	if !shouldCreate {
		return nil
	}

	filterableAttrs, err := s.getFilterableAttributeSlugs(ctx)
	if err != nil {
		s.logger.Error("Failed to get filterable attributes", "error", err)
		return err
	}

	indexOptions := searchtypes.IndexOptions{
		SearchableAttributes: []string{"name", "description", "meta_keyword", "model"},
		FilterableAttributes: append(
			[]string{"price", "category_id", "manufacturer_id", "is_enable", "stock_status"},
			filterableAttrs...,
		),
		SortableAttributes: []string{"price", "created_at", "name"},
	}

	page := uint64(1)
	pageSize := uint64(100)
	isFirstBatch := true

	for {
		res, err := s.GetEnrichedVariantsWithPagination(ctx, dto.GetDTO{
			Page:     tools.Pointer(page),
			PageSize: tools.Pointer(pageSize),
		})
		if err != nil {
			s.logger.Error("Failed to get variant page", "page", page, "error", err)
			return err
		}

		if len(res.Items) == 0 {
			break
		}

		data, err := s.buildVariantDocuments(ctx, res.Items)
		if err != nil {
			s.logger.Error("Failed to build variant documents", "error", err)
			return err
		}

		if isFirstBatch {
			err = s.searchService.CreateIndex(constant.ProductVariantsIndex, data, indexOptions)
			isFirstBatch = false
		} else {
			err = s.searchService.CreateIndex(constant.ProductVariantsIndex, data)
		}

		if err != nil {
			s.logger.Error("Failed to index variant batch", "page", page, "error", err)
			return err
		}

		s.logger.Debug("Indexed variant batch", "page", page, "count", len(res.Items))

		if page >= res.Pagination.LastPage {
			break
		}

		page++
	}

	s.logger.Info("Product variant index created successfully")

	return nil
}

func (s *Service) IndexVariant(ctx context.Context, variant *models.ProductVariant, product *models.Product) error {
	doc, err := s.buildVariantDocument(ctx, variant, product)
	if err != nil {
		return err
	}
	return s.searchService.UpsertDocument(constant.ProductVariantsIndex, []map[string]any{doc})
}

func (s *Service) buildVariantDocuments(ctx context.Context, variants []*dto.EnrichedVariantDTO) ([]map[string]any, error) {
	attrCache := make(map[uuid.UUID][]*repository_product_attribute_values.GetByProductIDRow)

	docs := make([]map[string]any, 0, len(variants))
	for _, v := range variants {
		attrs, ok := attrCache[v.ProductID]
		if !ok {
			var err error
			attrs, err = s.storage.ProductAttributeValues().GetByProductID(ctx, v.ProductID)
			if err != nil {
				s.logger.Warn("Failed to get attributes for product", "product_id", v.ProductID, "error", err)
			}
			attrCache[v.ProductID] = attrs
		}

		docs = append(docs, s.variantToDocument(v, attrs))
	}

	return docs, nil
}

func (s *Service) buildVariantDocument(ctx context.Context, variant *models.ProductVariant, product *models.Product) (map[string]any, error) {
	enriched := &dto.EnrichedVariantDTO{
		ProductVariant: variant,
		Price:          product.Price,
		ManufacturerID: product.ManufacturerID,
		StockStatus:    product.StockStatus,
		Model:          product.Model,
	}

	attrs, err := s.storage.ProductAttributeValues().GetByProductID(ctx, product.ID)
	if err != nil {
		s.logger.Warn("Failed to get attributes for product", "product_id", product.ID, "error", err)
	}

	return s.variantToDocument(enriched, attrs), nil
}

func (s *Service) variantToDocument(v *dto.EnrichedVariantDTO, attrs []*repository_product_attribute_values.GetByProductIDRow) map[string]any {
	doc := map[string]any{
		"id":              v.ID,
		"product_id":      v.ProductID,
		"category_id":     v.CategoryID,
		"name":            v.Name,
		"slug":            v.Slug,
		"description":     v.Description.String,
		"meta_keyword":    v.MetaKeyword.String,
		"image":           v.Image.String,
		"is_enable":       v.IsEnable,
		"sort_order":      v.SortOrder,
		"viewed":          v.Viewed,
		"created_at":      v.CreatedAt,
		"price":           v.Price,
		"manufacturer_id": v.ManufacturerID,
		"stock_status":    v.StockStatus,
		"model":           v.Model,
	}

	for _, attr := range attrs {
		var fieldValue any
		switch attr.AttributeType {
		case "number":
			if attr.ValueNumeric.Valid {
				val, _ := attr.ValueNumeric.Decimal.Float64()
				fieldValue = val
			}
		case "boolean":
			fieldValue = attr.AttributeValue == "true" || attr.AttributeValue == "1"
		default:
			fieldValue = attr.AttributeValue
		}
		doc[attr.AttributeSlug] = fieldValue
	}

	return doc
}

func (s *Service) getFilterableAttributeSlugs(ctx context.Context) ([]string, error) {
	attributes, err := s.storage.Attributes().GetFilterableAttributes(ctx)
	if err != nil {
		return nil, err
	}

	slugs := make([]string, 0, len(attributes))
	for _, attr := range attributes {
		slugs = append(slugs, attr.Slug)
	}

	return slugs, nil
}
