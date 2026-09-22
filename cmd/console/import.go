package console

import (
	"context"
	"fmt"

	"github.com/stickpro/go-store/internal/app"
	"github.com/stickpro/go-store/internal/tools/opencartimport"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/urfave/cli/v3"
)

func prepareImportCommands(appName, currentAppVersion string) []*cli.Command {
	return []*cli.Command{
		{
			Name:        "opencart",
			Description: "one-off import of ProductVariants from a legacy OpenCart (3.x/4.x) MySQL database, onto go-store Products that already exist (matched by products.sku == OpenCart's model)",
			Flags: []cli.Flag{
				cfgPathsFlag(),
				&cli.StringFlag{Name: "oc-host", Value: "127.0.0.1", Usage: "OpenCart MySQL host"},
				&cli.IntFlag{Name: "oc-port", Value: 3306, Usage: "OpenCart MySQL port"},
				&cli.StringFlag{Name: "oc-user", Usage: "OpenCart MySQL user"},
				&cli.StringFlag{Name: "oc-password", Usage: "OpenCart MySQL password"},
				&cli.StringFlag{Name: "oc-database", Usage: "OpenCart MySQL database name"},
				&cli.StringFlag{Name: "oc-prefix", Value: "oc_", Usage: "OpenCart table prefix"},
				&cli.IntFlag{Name: "oc-store-id", Value: 0, Usage: "OpenCart store_id to read SEO URLs for"},
				&cli.IntFlag{Name: "oc-language-id", Value: 1, Usage: "OpenCart language_id to read descriptions/SEO URLs for"},
				&cli.BoolFlag{Name: "only-enabled", Value: true, Usage: "skip disabled OpenCart products"},
				&cli.IntFlag{Name: "limit", Value: 0, Usage: "process at most N OpenCart model-groups (0 = all); useful for a test run"},
				&cli.BoolFlag{Name: "dry-run", Usage: "log what would be created/updated without writing anything"},
				&cli.IntFlag{Name: "workers", Value: 8, Usage: "number of model-groups processed concurrently"},
			},
			Action: func(ctx context.Context, cl *cli.Command) error {
				conf, err := loadConfig(cl.Args().Slice(), cl.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				loggerOpts := append(defaultLoggerOpts(appName, currentAppVersion), logger.WithConfig(conf.Log))
				l := logger.NewExtended(loggerOpts...)
				defer func() { _ = l.Sync() }()

				ocCfg := opencartimport.Config{
					Host:        cl.String("oc-host"),
					Port:        cl.Int("oc-port"),
					User:        cl.String("oc-user"),
					Password:    cl.String("oc-password"),
					Database:    cl.String("oc-database"),
					TablePrefix: cl.String("oc-prefix"),
					StoreID:     int64(cl.Int("oc-store-id")),
					LanguageID:  int64(cl.Int("oc-language-id")),
				}
				opts := opencartimport.Options{
					OnlyEnabled: cl.Bool("only-enabled"),
					Limit:       cl.Int("limit"),
					DryRun:      cl.Bool("dry-run"),
					Workers:     cl.Int("workers"),
				}

				report, err := app.ImportOpenCart(ctx, conf, l, ocCfg, opts)
				if report != nil {
					fmt.Printf("model groups: %d (matched: %d, unmatched: %d)\n",
						report.ModelGroups, report.ProductsMatched, report.ProductsUnmatched)
					fmt.Printf("variants: %d created, %d updated\n", report.VariantsCreated, report.VariantsUpdated)
					if len(report.Errors) > 0 {
						fmt.Printf("%d errors:\n", len(report.Errors))
						for _, e := range report.Errors {
							fmt.Println("  " + e)
						}
					}
				}
				if err != nil {
					return fmt.Errorf("import opencart: %w", err)
				}
				return nil
			},
		},
	}
}
