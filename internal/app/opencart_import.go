package app

import (
	"context"
	"fmt"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/service"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/internal/tools/opencartimport"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/stickpro/go-store/pkg/queue"
)

// ImportOpenCart boots the storage + service layer, creates ProductVariants
// on already-existing go-store Products from an OpenCart MySQL database
// (matched by products.sku == OpenCart's model), then tears everything down.
// Used by the `import opencart` console command for a one-off catalog migration.
func ImportOpenCart(ctx context.Context, conf *config.Config, l logger.Logger, ocCfg opencartimport.Config, opts opencartimport.Options) (*opencartimport.Report, error) {
	oc, err := opencartimport.NewClient(ctx, ocCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to opencart database: %w", err)
	}
	defer func() {
		if cerr := oc.Close(); cerr != nil {
			l.Error("opencart connection close error", cerr)
		}
	}()

	st, err := storage.InitStore(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("init store: %w", err)
	}
	defer func() {
		if cerr := st.Close(); cerr != nil {
			l.Error("storage close error", cerr)
		}
	}()

	// Import runs no mail worker; give the service layer a throwaway in-memory
	// queue so enqueued mail (none is expected here) has somewhere to go.
	services, err := service.InitService(conf, l, st, queue.NewInMemoryQueue())
	if err != nil {
		return nil, fmt.Errorf("init services: %w", err)
	}
	defer func() {
		if cerr := services.Close(); cerr != nil {
			l.Error("services close error", cerr)
		}
	}()

	return opencartimport.Run(ctx, l, services, oc, opts)
}
