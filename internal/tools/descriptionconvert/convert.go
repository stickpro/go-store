// Package descriptionconvert is a one-off maintenance tool that rewrites
// ProductVariant.Description from legacy HTML (produced by OpenCart's
// WYSIWYG editor, or otherwise hand-written) into Markdown, since go-store
// now stores and renders variant descriptions as Markdown, not HTML.
package descriptionconvert

import (
	"context"
	"fmt"
	"html"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/pkg/dbutils/pgtypeutils"
	"github.com/stickpro/go-store/pkg/logger"
)

// pageSize is how many variants are fetched per page while scanning the
// table; it stays under the 100-row MaxLimit the variants pagination config
// enforces (repository_product_variants.GetWithPaginate).
const pageSize uint64 = 100

// Options controls what a Run does.
type Options struct {
	DryRun bool
	// Limit caps the number of variants scanned (0 = no cap); useful for a test run.
	Limit int
}

// Report summarizes what a Run did.
type Report struct {
	Scanned   int
	Converted int
	Skipped   int
	Errors    []string
}

// Run pages through every ProductVariant and, for any description that looks
// like HTML (contains a '<'), rewrites it to Markdown. Descriptions that
// don't look like HTML are left untouched. Exact conversion fidelity isn't
// required, so a conversion error is recorded and the row is skipped rather
// than aborting the whole run.
func Run(ctx context.Context, l logger.Logger, services *service.Services, opts Options) (*Report, error) {
	report := &Report{}

	for page := uint64(1); ; page++ {
		size := pageSize
		res, err := services.ProductService.GetVariantsWithPagination(ctx, dto.GetDTO{Page: &page, PageSize: &size})
		if err != nil {
			return report, fmt.Errorf("list variants (page %d): %w", page, err)
		}

		for _, row := range res.Items {
			if opts.Limit > 0 && report.Scanned >= opts.Limit {
				return report, nil
			}
			report.Scanned++

			description := pgtypeutils.DecodeText(row.Description)
			if description == nil {
				report.Skipped++
				continue
			}
			candidate := unescapeIfDoubleEncoded(*description)
			if !strings.Contains(candidate, "<") {
				report.Skipped++
				continue
			}

			md, err := htmltomarkdown.ConvertString(candidate)
			if err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("variant %s (slug %s): convert: %v", row.ID, row.Slug, err))
				continue
			}

			if opts.DryRun {
				l.Infow("would convert variant description", "variant_id", row.ID, "slug", row.Slug)
				report.Converted++
				continue
			}

			if _, err := services.ProductService.UpdateProductVariant(ctx, row.ID, dto.UpdateProductVariantDTO{
				ID:              row.ID,
				Name:            row.Name,
				Slug:            row.Slug,
				CategoryID:      row.CategoryID,
				Description:     &md,
				MetaTitle:       pgtypeutils.DecodeText(row.MetaTitle),
				MetaH1:          pgtypeutils.DecodeText(row.MetaH1),
				MetaDescription: pgtypeutils.DecodeText(row.MetaDescription),
				MetaKeyword:     pgtypeutils.DecodeText(row.MetaKeyword),
				SortOrder:       row.SortOrder,
				IsEnable:        row.IsEnable,
			}); err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("variant %s (slug %s): update: %v", row.ID, row.Slug, err))
				continue
			}
			report.Converted++
		}

		if uint64(len(res.Items)) < pageSize || page >= res.Pagination.LastPage {
			break
		}
	}

	return report, nil
}

// unescapeIfDoubleEncoded undoes one extra layer of HTML-entity encoding.
// Some legacy rows store their description HTML-escaped a second time
// (literal "&lt;span&gt;" instead of "<span>"), which hides every tag from
// an HTML parser. Content with a real, unescaped tag is left untouched.
func unescapeIfDoubleEncoded(s string) string {
	if !strings.Contains(s, "<") && strings.Contains(s, "&lt;") {
		return html.UnescapeString(s)
	}
	return s
}
