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
			[]string{"price", "category_id", "category_ids", "manufacturer_id", "is_enable", "stock_status", "model"},
			filterableAttrs...,
		),
		SortableAttributes: []string{"price", "created_at", "name", "sort_order", "viewed"},
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

		data := s.buildVariantDocuments(ctx, res.Items)

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
	doc := s.buildVariantDocument(ctx, variant, product)
	return s.searchService.UpsertDocument(constant.ProductVariantsIndex, []map[string]any{doc})
}

func (s *Service) buildVariantDocuments(ctx context.Context, variants []*dto.EnrichedVariantDTO) []map[string]any {
	s.attachCategoryIDs(ctx, variants)
	s.attachMainImage(ctx, variants)

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

	return docs
}

func (s *Service) buildVariantDocument(ctx context.Context, variant *models.ProductVariant, product *models.Product) map[string]any {
	enriched := &dto.EnrichedVariantDTO{
		ProductVariant: variant,
		PriceRetail:    product.PriceRetail,
		PriceBusiness:  product.PriceBusiness,
		PriceWholesale: product.PriceWholesale,
		ManufacturerID: product.ManufacturerID,
		StockStatus:    product.StockStatus,
	}

	s.attachCategoryIDs(ctx, []*dto.EnrichedVariantDTO{enriched})
	s.attachMainImage(ctx, []*dto.EnrichedVariantDTO{enriched})

	attrs, err := s.storage.ProductAttributeValues().GetByProductID(ctx, product.ID)
	if err != nil {
		s.logger.Warn("Failed to get attributes for product", "product_id", product.ID, "error", err)
	}

	return s.variantToDocument(enriched, attrs)
}

func (s *Service) variantToDocument(v *dto.EnrichedVariantDTO, attrs []*repository_product_attribute_values.GetByProductIDRow) map[string]any {
	// price is stored as a plain number (decimal marshals to a JSON string, which Meili
	// cannot filter/sort/facet numerically) — retail price is the storefront-facing one.
	priceRetail, _ := v.PriceRetail.Float64()

	doc := map[string]any{
		"id":              v.ID,
		"product_id":      v.ProductID,
		"category_id":     v.CategoryID,
		"category_ids":    v.CategoryIDs,
		"image":           v.Image,
		"name":            v.Name,
		"slug":            v.Slug,
		"description":     v.Description.String,
		"meta_keyword":    v.MetaKeyword.String,
		"is_enable":       v.IsEnable,
		"sort_order":      v.SortOrder,
		"viewed":          v.Viewed,
		"created_at":      v.CreatedAt,
		"price":           priceRetail,
		"price_retail":    v.PriceRetail,
		"price_business":  v.PriceBusiness,
		"price_wholesale": v.PriceWholesale,
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

// attachMainImage sets EnrichedVariantDTO.Image to the product's first gallery image
// (product_media ordered by sort_order), so the search document carries the ready
// ImageDTO for listings. Requires a reindex when the preset scheme changes.
func (s *Service) attachMainImage(ctx context.Context, variants []*dto.EnrichedVariantDTO) {
	if len(variants) == 0 {
		return
	}

	productIDs := make([]uuid.UUID, 0, len(variants))
	seen := make(map[uuid.UUID]struct{}, len(variants))
	for _, v := range variants {
		if _, ok := seen[v.ProductID]; !ok {
			seen[v.ProductID] = struct{}{}
			productIDs = append(productIDs, v.ProductID)
		}
	}

	rows, err := s.storage.Products().GetMainMediaByProductIDs(ctx, productIDs)
	if err != nil {
		s.logger.Warn("Failed to load main media for index", "error", err)
		return
	}

	presets := s.cfg.Images.ResolvedPresets()
	byProduct := make(map[uuid.UUID]dto.ImageDTO, len(rows))
	for _, r := range rows {
		byProduct[r.ProductID] = dto.NewImageDTO(r.ID, r.Path, r.Width, r.Height, "", presets)
	}

	for _, v := range variants {
		if img, ok := byProduct[v.ProductID]; ok {
			img.Alt = v.Name
			v.Image = &img
		}
	}
}

// attachCategoryIDs fills EnrichedVariantDTO.CategoryIDs for each variant with the
// closure of its primary category and every additional (junction) category — i.e.
// each of those categories plus all of their ancestors. This lets the search index
// match a whole subtree with a single `category_ids = <id>` filter.
func (s *Service) attachCategoryIDs(ctx context.Context, variants []*dto.EnrichedVariantDTO) {
	if len(variants) == 0 {
		return
	}

	variantIDs := make([]uuid.UUID, 0, len(variants))
	for _, v := range variants {
		variantIDs = append(variantIDs, v.ID)
	}

	junction := make(map[uuid.UUID][]uuid.UUID)
	rows, err := s.storage.ProductVariantCategories().GetByVariantIDs(ctx, variantIDs)
	if err != nil {
		s.logger.Warn("Failed to load variant categories for index", "error", err)
	}
	for _, r := range rows {
		junction[r.ProductVariantID] = append(junction[r.ProductVariantID], r.CategoryID)
	}

	seedSet := make(map[uuid.UUID]struct{})
	for _, v := range variants {
		if v.CategoryID.Valid {
			seedSet[v.CategoryID.UUID] = struct{}{}
		}
		for _, c := range junction[v.ID] {
			seedSet[c] = struct{}{}
		}
	}
	if len(seedSet) == 0 {
		return
	}

	seeds := make([]uuid.UUID, 0, len(seedSet))
	for id := range seedSet {
		seeds = append(seeds, id)
	}

	// descendant category id -> its ancestor ids (the closure table includes the self row at depth 0)
	ancestors := make(map[uuid.UUID][]uuid.UUID)
	pathRows, err := s.storage.CategoryPaths().GetCategoryPathsBatch(ctx, seeds)
	if err != nil {
		s.logger.Warn("Failed to load category paths for index", "error", err)
	}
	for _, r := range pathRows {
		ancestors[r.DescendantID] = append(ancestors[r.DescendantID], r.AncestorID)
	}

	for _, v := range variants {
		set := make(map[uuid.UUID]struct{})
		collect := func(id uuid.UUID) {
			set[id] = struct{}{}
			for _, a := range ancestors[id] {
				set[a] = struct{}{}
			}
		}
		if v.CategoryID.Valid {
			collect(v.CategoryID.UUID)
		}
		for _, c := range junction[v.ID] {
			collect(c)
		}

		ids := make([]uuid.UUID, 0, len(set))
		for id := range set {
			ids = append(ids, id)
		}
		v.CategoryIDs = ids
	}
}

func (s *Service) getFilterableAttributeSlugs(ctx context.Context) ([]string, error) {
	attributes, err := s.filterableAttributes(ctx)
	if err != nil {
		return nil, err
	}

	slugs := make([]string, 0, len(attributes))
	for _, attr := range attributes {
		slugs = append(slugs, attr.Slug)
	}

	return slugs, nil
}
