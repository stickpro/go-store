package console

import (
	"context"
	"fmt"

	"github.com/stickpro/go-store/internal/app"
	"github.com/stickpro/go-store/internal/tools/descriptionconvert"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/urfave/cli/v3"
)

func prepareDescriptionCommands(appName, currentAppVersion string) []*cli.Command {
	return []*cli.Command{
		{
			Name: "html-to-markdown",
			Description: "one-off conversion of legacy HTML ProductVariant descriptions (e.g. from the OpenCart import) " +
				"into Markdown; descriptions that don't look like HTML are left untouched",
			Flags: []cli.Flag{
				cfgPathsFlag(),
				&cli.BoolFlag{Name: "dry-run", Value: true, Usage: "log what would be converted without writing anything"},
				&cli.IntFlag{Name: "limit", Value: 0, Usage: "scan at most N variants (0 = all); useful for a test run"},
			},
			Action: func(ctx context.Context, cl *cli.Command) error {
				conf, err := loadConfig(cl.Args().Slice(), cl.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				loggerOpts := append(defaultLoggerOpts(appName, currentAppVersion), logger.WithConfig(conf.Log))
				l := logger.NewExtended(loggerOpts...)
				defer func() { _ = l.Sync() }()

				opts := descriptionconvert.Options{
					DryRun: cl.Bool("dry-run"),
					Limit:  cl.Int("limit"),
				}

				report, err := app.ConvertVariantDescriptions(ctx, conf, l, opts)
				if report != nil {
					mode := "converted"
					if opts.DryRun {
						mode = "would convert"
					}
					fmt.Printf("scanned: %d, %s: %d, skipped (no html): %d\n", report.Scanned, mode, report.Converted, report.Skipped)
					if len(report.Errors) > 0 {
						fmt.Printf("%d errors:\n", len(report.Errors))
						for _, e := range report.Errors {
							fmt.Println("  " + e)
						}
					}
				}
				if err != nil {
					return fmt.Errorf("convert descriptions: %w", err)
				}
				return nil
			},
		},
	}
}
