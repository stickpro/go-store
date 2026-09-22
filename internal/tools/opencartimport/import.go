// Package opencartimport reads categories and products out of a legacy OpenCart
// (3.x/4.x) MySQL database and creates missing ProductVariants on go-store
// Products that already exist, for a one-off catalog migration.
//
// It never creates a Product. OpenCart's `model` column groups several product
// rows (color/size variants sharing manufacturer/price/dimensions) under one
// base item; go-store already has that base item as a Product with
// products.sku == that model. Every OpenCart row in the group becomes one
// ProductVariant on the matched Product. A model with no matching sku in
// go-store is skipped entirely.
package opencartimport

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/pkg/dbutils/pgerror"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/logger"
)

const (
	// listAllPageSize is the max page size the service layer's pagination
	// helper allows (see pkg/dbutils.Pagination); listing all existing
	// variant slugs for dedup has to page through it.
	listAllPageSize uint64 = 100
	// bulkFetchTimeout bounds the whole-table MySQL fetch of product-category
	// links - a remote connection that stalls should fail loudly instead of
	// hanging the run forever.
	bulkFetchTimeout = 20 * time.Minute
)

// Options controls what a Run does.
type Options struct {
	OnlyEnabled bool
	// Limit caps the number of OpenCart model-groups processed (not rows) - useful for a test run.
	Limit  int
	DryRun bool
	// Workers is how many model-groups are processed concurrently.
	Workers int
}

// Report summarizes what a Run did.
type Report struct {
	mu sync.Mutex

	ModelGroups       int
	ProductsMatched   int
	ProductsUnmatched int
	VariantsCreated   int
	VariantsUpdated   int
	Errors            []string
}

func (r *Report) addErr(format string, args ...any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Errors = append(r.Errors, fmt.Sprintf(format, args...))
}

func (r *Report) inc(counter *int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	*counter++
}

// productState is the mutable state shared across concurrent group workers.
type productState struct {
	mu               sync.Mutex
	usedVariantSlugs map[string]bool
}

// reserveNewVariantSlug picks and reserves a slug for a brand-new variant.
// Never call this for a variant that already exists.
func (s *productState) reserveNewVariantSlug(seoKeyword, name string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	slug := resolveSlug(seoKeyword, name, func(c string) bool { return s.usedVariantSlugs[c] })
	s.usedVariantSlugs[slug] = true
	return slug
}

// Run migrates ProductVariants for OpenCart products grouped by `model` onto
// go-store Products matched by products.sku == model. It never creates a
// Product - a model group with no matching sku is skipped and counted as
// unmatched.
func Run(ctx context.Context, l logger.Logger, services *service.Services, oc *Client, opts Options) (*Report, error) {
	report := &Report{}

	l.Info("loading existing go-store product skus")
	skus, err := loadExistingSkus(ctx, services)
	if err != nil {
		return report, fmt.Errorf("load existing skus: %w", err)
	}
	l.Infow("existing skus loaded", "count", len(skus))

	l.Info("fetching matching OpenCart products by model")
	groups, err := fetchModelGroups(ctx, l, oc, skus, opts.OnlyEnabled)
	if err != nil {
		return report, fmt.Errorf("fetch opencart products: %w", err)
	}
	l.Infow("grouped opencart products by model", "groups", len(groups))

	l.Info("fetching product SEO keywords from MySQL")
	seoKeywords, err := oc.FetchSeoKeywords(ctx, "product_id")
	if err != nil {
		return report, fmt.Errorf("fetch product seo keywords: %w", err)
	}

	l.Info("fetching product-category links from MySQL")
	productCategoryLinks, err := fetchWithTimeout(ctx, func(ctx context.Context) (map[int64][]int64, error) {
		return oc.FetchAllProductToCategory(ctx, func(n int) { l.Infow("product-category links progress", "rows", n) })
	})
	if err != nil {
		return report, fmt.Errorf("fetch product-category links: %w", err)
	}

	l.Info("resolving opencart categories against existing go-store categories")
	categoryIDs, err := resolveExistingCategories(ctx, oc, services)
	if err != nil {
		return report, fmt.Errorf("resolve categories: %w", err)
	}
	l.Infow("categories resolved", "matched", len(categoryIDs))

	l.Info("preloading existing variant slugs from go-store")
	usedVariantSlugs, err := preloadVariantSlugs(ctx, services)
	if err != nil {
		return report, err
	}
	state := &productState{usedVariantSlugs: usedVariantSlugs}

	workers := opts.Workers
	if workers < 1 {
		workers = 1
	}
	l.Infow("starting group workers", "workers", workers)

	modelKeys := sortedModelKeys(groups)
	if opts.Limit > 0 && len(modelKeys) > opts.Limit {
		modelKeys = modelKeys[:opts.Limit]
	}

	const progressEvery = 200
	var processed atomic.Int64

	type job struct {
		model string
		rows  []ProductRow
	}
	jobCh := make(chan job, workers*2)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobCh {
				if err := processModelGroup(ctx, l, services, j.model, j.rows, seoKeywords, productCategoryLinks, categoryIDs, state, report, opts.DryRun); err != nil {
					report.addErr("model %q: %v", j.model, err)
				}
				if n := processed.Add(1); n%progressEvery == 0 {
					l.Infow("group processing progress", "processed", n)
				}
			}
		}()
	}

	for _, model := range modelKeys {
		report.inc(&report.ModelGroups)
		jobCh <- job{model: model, rows: groups[model]}
	}
	close(jobCh)
	wg.Wait()

	return report, nil
}

// loadExistingSkus pages through every go-store product and collects its sku
// - the set of OpenCart `model` values the import actually needs to fetch.
func loadExistingSkus(ctx context.Context, services *service.Services) ([]string, error) {
	var skus []string
	for page := uint64(1); ; page++ {
		pageSize := listAllPageSize
		res, err := services.ProductService.GetProductWithPagination(ctx, dto.GetDTO{Page: &page, PageSize: &pageSize})
		if err != nil {
			return nil, fmt.Errorf("list existing products: %w", err)
		}
		for _, p := range res.Items {
			if sku := pgtypeutils.DecodeText(p.Sku); sku != nil && *sku != "" {
				skus = append(skus, *sku)
			}
		}
		if uint64(len(res.Items)) < listAllPageSize || page >= res.Pagination.LastPage {
			break
		}
	}
	return skus, nil
}

// fetchModelGroups fetches only the OpenCart products matching one of skus
// (go-store's existing sku set) and groups the rows by their `model` column.
// Rows with an empty model (nothing to match a go-store sku against) are
// dropped.
func fetchModelGroups(ctx context.Context, l logger.Logger, oc *Client, skus []string, onlyEnabled bool) (map[string][]ProductRow, error) {
	rows, err := fetchWithTimeout(ctx, func(ctx context.Context) ([]ProductRow, error) {
		return oc.FetchProductsByModels(ctx, skus, func(n int) { l.Infow("product fetch progress", "rows", n) })
	})
	if err != nil {
		return nil, err
	}

	groups := map[string][]ProductRow{}
	for _, row := range rows {
		if onlyEnabled && !row.Status {
			continue
		}
		if row.Model == "" {
			continue
		}
		groups[row.Model] = append(groups[row.Model], row)
	}
	return groups, nil
}

func sortedModelKeys(groups map[string][]ProductRow) []string {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// resolveExistingCategories maps OpenCart category_id -> go-store category ID
// for categories that already exist in go-store (matched by the same slug
// scheme the original catalog import used: sanitized SEO keyword, else
// slugified name). It never creates a category - a miss is simply left out.
func resolveExistingCategories(ctx context.Context, oc *Client, services *service.Services) (map[int64]uuid.UUID, error) {
	rows, err := oc.FetchCategories(ctx)
	if err != nil {
		return nil, err
	}
	keywords, err := oc.FetchSeoKeywords(ctx, "category_id")
	if err != nil {
		return nil, fmt.Errorf("fetch category seo keywords: %w", err)
	}

	result := map[int64]uuid.UUID{}
	for _, row := range rows {
		slug := SanitizeKeyword(keywords[row.ID])
		if slug == "" {
			slug = Slugify(row.Name)
		}
		cat, err := services.CategoryService.GetCategoryBySlug(ctx, slug)
		if err != nil {
			if isNotFound(err) {
				continue
			}
			return nil, fmt.Errorf("lookup category %d (%s): %w", row.ID, row.Name, err)
		}
		result[row.ID] = cat.ID
	}
	return result, nil
}

// preloadVariantSlugs reads every existing product_variants.slug once up front.
// product_variants.slug has no unique DB constraint, so this - not a live
// per-variant check - is what keeps concurrent workers (and reruns) from
// generating two variants with the same slug.
func preloadVariantSlugs(ctx context.Context, services *service.Services) (map[string]bool, error) {
	used := map[string]bool{}
	for page := uint64(1); ; page++ {
		pageSize := listAllPageSize
		res, err := services.ProductService.GetVariantsWithPagination(ctx, dto.GetDTO{Page: &page, PageSize: &pageSize})
		if err != nil {
			return nil, fmt.Errorf("list existing variant slugs: %w", err)
		}
		for _, v := range res.Items {
			used[v.Slug] = true
		}
		if uint64(len(res.Items)) < listAllPageSize || page >= res.Pagination.LastPage {
			break
		}
	}
	return used, nil
}

// processModelGroup handles every OpenCart row sharing one `model` value: it
// looks up the go-store Product whose sku equals that model and, only if
// found, creates or updates one ProductVariant per row.
func processModelGroup(
	ctx context.Context,
	l logger.Logger,
	services *service.Services,
	model string,
	rows []ProductRow,
	seoKeywords map[int64]string,
	productCategoryLinks map[int64][]int64,
	categoryIDs map[int64]uuid.UUID,
	state *productState,
	report *Report,
	dryRun bool,
) error {
	prd, err := services.ProductService.GetProductBySku(ctx, model)
	if err != nil {
		if isNotFound(err) {
			report.inc(&report.ProductsUnmatched)
			l.Debugw("no matching product for opencart model, skipping", "model", model)
			return nil
		}
		return fmt.Errorf("lookup product by sku: %w", err)
	}
	report.inc(&report.ProductsMatched)

	existingVariants, err := services.ProductService.GetProductVariants(ctx, prd.ID)
	if err != nil {
		return fmt.Errorf("get existing variants: %w", err)
	}
	existingByModel := make(map[string]*models.ProductVariant, len(existingVariants))
	for _, v := range existingVariants {
		existingByModel[v.Model] = v
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })

	for _, row := range rows {
		variantModel := row.Sku
		if variantModel == "" {
			variantModel = fmt.Sprintf("%s-%d", model, row.ID)
		}

		var mappedCategoryIDs []uuid.UUID
		for _, oldID := range productCategoryLinks[row.ID] {
			if newID, ok := categoryIDs[oldID]; ok {
				mappedCategoryIDs = append(mappedCategoryIDs, newID)
			}
		}
		var primaryCategoryID uuid.NullUUID
		if len(mappedCategoryIDs) > 0 {
			primaryCategoryID = uuid.NullUUID{UUID: mappedCategoryIDs[0], Valid: true}
		}

		existing, matched := existingByModel[variantModel]

		if dryRun {
			action := "create"
			if matched {
				action = "update"
			}
			l.Infow("would "+action+" variant", "model", variantModel, "product_sku", model, "name", row.Name)
			continue
		}

		if matched {
			if _, err := services.ProductService.UpdateProductVariant(ctx, existing.ID, dto.UpdateProductVariantDTO{
				ID:              existing.ID,
				Name:            row.Name,
				Slug:            existing.Slug,
				Model:           variantModel,
				CategoryID:      primaryCategoryID,
				Description:     nonEmpty(row.Description),
				MetaTitle:       nonEmpty(row.MetaTitle),
				MetaH1:          nonEmpty(row.MetaH1),
				MetaDescription: nonEmpty(row.MetaDescription),
				MetaKeyword:     nonEmpty(row.MetaKeyword),
				SortOrder:       row.SortOrder,
				IsEnable:        row.Status,
			}); err != nil {
				report.addErr("variant %d (model %s): update: %v", row.ID, variantModel, err)
				continue
			}
			report.inc(&report.VariantsUpdated)
			continue
		}

		slug := state.reserveNewVariantSlug(seoKeywords[row.ID], row.Name)
		if _, err := services.ProductService.CreateProductVariant(ctx, prd.ID, dto.CreateProductVariantDTO{
			Name:            row.Name,
			Slug:            slug,
			Model:           variantModel,
			CategoryID:      primaryCategoryID,
			Description:     nonEmpty(row.Description),
			MetaTitle:       nonEmpty(row.MetaTitle),
			MetaH1:          nonEmpty(row.MetaH1),
			MetaDescription: nonEmpty(row.MetaDescription),
			MetaKeyword:     nonEmpty(row.MetaKeyword),
			SortOrder:       row.SortOrder,
			IsEnable:        row.Status,
		}); err != nil {
			report.addErr("variant %d (model %s): create: %v", row.ID, variantModel, err)
			continue
		}
		report.inc(&report.VariantsCreated)
	}

	return nil
}

// resolveSlug prefers a sanitized OpenCart SEO keyword, falling back to a
// slugified name, uniquified against exists.
func resolveSlug(seoKeyword, name string, exists func(string) bool) string {
	base := SanitizeKeyword(seoKeyword)
	if base == "" {
		base = Slugify(name)
	}
	return Uniquify(base, exists)
}

// fetchWithTimeout bounds a whole-table MySQL fetch so a stalled remote
// connection errors out instead of hanging the run forever.
func fetchWithTimeout[T any](ctx context.Context, fn func(context.Context) (T, error)) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, bulkFetchTimeout)
	defer cancel()
	return fn(ctx)
}

func nonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func isNotFound(err error) bool {
	if errors.Is(err, pgx.ErrNoRows) {
		return true
	}
	var nf *pgerror.NotFoundError
	return errors.As(err, &nf)
}
