package app

import (
	"context"
	"fmt"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/queue"
)

// ReindexTarget selects which search indexes a Reindex run rebuilds.
type ReindexTarget struct {
	Products   bool
	Attributes bool
	Cities     bool
	// CategoryPaths rebuilds the category_paths closure table before reindexing,
	// so product docs get the correct subtree category_ids.
	CategoryPaths bool
}

// AllReindexTargets rebuilds every search index.
func AllReindexTargets() ReindexTarget {
	return ReindexTarget{Products: true, Attributes: true, Cities: true}
}

// Reindex boots the storage + service layer, drops and rebuilds the selected search
// indexes from the database, then tears everything down. Used by the `search reindex`
// console command; the same rebuild also runs automatically on every server start.
func Reindex(ctx context.Context, conf *config.Config, l logger.Logger, target ReindexTarget) error {
	st, err := storage.InitStore(ctx, conf)
	if err != nil {
		return fmt.Errorf("init store: %w", err)
	}
	defer func() {
		if cerr := st.Close(); cerr != nil {
			l.Error("storage close error", cerr)
		}
	}()

	// Reindex runs no mail worker; give the service layer a throwaway in-memory
	// queue so enqueued mail (none is expected here) has somewhere to go.
	services, err := service.InitService(conf, l, st, queue.NewInMemoryQueue())
	if err != nil {
		return fmt.Errorf("init services: %w", err)
	}
	defer func() {
		if cerr := services.Close(); cerr != nil {
			l.Error("services close error", cerr)
		}
	}()

	if target.CategoryPaths && services.CategoryService != nil {
		l.Info("rebuilding category paths")
		if err := services.CategoryService.RebuildAllCategoryPaths(ctx); err != nil {
			return fmt.Errorf("rebuild category paths: %w", err)
		}
	}

	if target.Cities && services.GeoService != nil {
		l.Info("reindexing cities")
		if err := services.GeoService.CreateCityIndex(ctx, true); err != nil {
			return fmt.Errorf("city index: %w", err)
		}
	}

	if target.Products && services.ProductService != nil {
		l.Info("reindexing product variants")
		if err := services.ProductService.CreateProductVariantIndex(ctx, true); err != nil {
			return fmt.Errorf("product variant index: %w", err)
		}
	}

	if target.Attributes && services.AttributeService != nil {
		l.Info("reindexing attribute groups")
		if err := services.AttributeService.CreateAttributeGroupIndex(ctx, true); err != nil {
			return fmt.Errorf("attribute group index: %w", err)
		}
		l.Info("reindexing attributes")
		if err := services.AttributeService.CreateAttributeIndex(ctx, true); err != nil {
			return fmt.Errorf("attribute index: %w", err)
		}
	}

	l.Info("reindex complete")
	return nil
}
