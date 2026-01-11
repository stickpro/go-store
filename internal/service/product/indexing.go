package product

import (
	"context"

	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service/search/searchtypes"
	"github.com/stickpro/go-store/internal/storage/repository/repository_products"
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

	// Get filterable attribute slugs for index configuration
	filterableAttrs, err := s.getFilterableAttributeSlugs(ctx)
	if err != nil {
		s.logger.Error("Failed to get filterable attributes", "error", err)
		return err
	}

	// Configure index with filterable attributes
	indexOptions := searchtypes.IndexOptions{
		SearchableAttributes: []string{"name", "description", "model", "meta_keyword"},
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
		d := dto.GetDTO{
			Page:     tools.Pointer(page),
			PageSize: tools.Pointer(pageSize),
		}

		res, err := s.GetProductWithPagination(ctx, d)
		if err != nil {
			s.logger.Error("Failed to get product page", "page", page, "error", err)
			return err
		}

		if len(res.Items) == 0 {
			break
		}

		// Build product documents with attributes
		data, err := s.buildProductDocuments(ctx, res.Items)
		if err != nil {
			s.logger.Error("Failed to build product documents", "error", err)
			return err
		}

		// Pass IndexOptions only on first batch
		if isFirstBatch {
			err = s.searchService.CreateIndex(constant.ProductsIndex, data, indexOptions)
			isFirstBatch = false
		} else {
			err = s.searchService.CreateIndex(constant.ProductsIndex, data)
		}

		if err != nil {
			s.logger.Error("Failed to index product batch ", "page ", page, "error ", err)
			return err
		}

		s.logger.Debug("Indexed product batch ", "page", page, "count ", len(res.Items))

		if page >= res.Pagination.LastPage {
			break
		}

		page++
	}

	s.logger.Info("Product index created successfully")

	return nil
}

func (s *Service) IndexProduct(ctx context.Context, product *models.Product) error {
	// Build document with attributes
	docs, err := s.buildProductDocuments(ctx, []*repository_products.FindRow{
		{Product: *product},
	})
	if err != nil {
		return err
	}

	return s.searchService.UpsertDocument(constant.ProductsIndex, docs)
}

// getFilterableAttributeSlugs retrieves all filterable attribute slugs from database
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

// buildProductDocuments converts product rows to search documents with attributes
func (s *Service) buildProductDocuments(ctx context.Context, products []*repository_products.FindRow) ([]map[string]interface{}, error) {
	// Convert products to maps
	baseDocs, err := utils.StructToMap(products)
	if err != nil {
		return nil, err
	}

	// Enrich each document with product attributes
	for i, product := range products {
		// Get product attributes
		attrRows, err := s.storage.ProductAttributeValues().GetByProductID(ctx, product.ID)
		if err != nil {
			s.logger.Warn("Failed to get attributes for product", "product_id", product.ID, "error", err)
			continue
		}

		// Add attributes to document using slug as key
		for _, attr := range attrRows {
			fieldName := attr.AttributeSlug

			// Convert value based on attribute type
			var fieldValue interface{}
			switch attr.AttributeType {
			case "number":
				if attr.ValueNumeric.Valid {
					val, _ := attr.ValueNumeric.Decimal.Float64()
					fieldValue = val
				}
			case "boolean":
				// Parse boolean from string value
				fieldValue = attr.AttributeValue == "true" || attr.AttributeValue == "1"
			default: // select, text
				fieldValue = attr.AttributeValue
			}

			baseDocs[i][fieldName] = fieldValue
		}
	}

	return baseDocs, nil
}
