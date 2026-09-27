package app

import (
	"context"
	"fmt"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/tools/descriptionconvert"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/queue"
)

// ConvertVariantDescriptions boots the storage + service layer, rewrites
// ProductVariant descriptions that look like HTML into Markdown, then tears
// everything down. Used by the `description html-to-markdown` console
// command for a one-off cleanup of legacy (e.g. OpenCart-imported) data.
func ConvertVariantDescriptions(ctx context.Context, conf *config.Config, l logger.Logger, opts descriptionconvert.Options) (*descriptionconvert.Report, error) {
	st, err := storage.InitStore(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("init store: %w", err)
	}
	defer func() {
		if cerr := st.Close(); cerr != nil {
			l.Error("storage close error", cerr)
		}
	}()

	// This maintenance run sends no mail; give the service layer a throwaway
	// in-memory queue so enqueued mail (none is expected here) has somewhere to go.
	services, err := service.InitService(conf, l, st, queue.NewInMemoryQueue())
	if err != nil {
		return nil, fmt.Errorf("init services: %w", err)
	}
	defer func() {
		if cerr := services.Close(); cerr != nil {
			l.Error("services close error", cerr)
		}
	}()

	return descriptionconvert.Run(ctx, l, services, opts)
}
