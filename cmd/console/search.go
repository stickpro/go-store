package console

import (
	"context"
	"fmt"

	"github.com/stickpro/go-store/internal/app"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/urfave/cli/v3"
)

func prepareSearchCommands(appName, currentAppVersion string) []*cli.Command {
	return []*cli.Command{
		{
			Name:        "reindex",
			Description: "drop and rebuild search indexes from the database",
			Flags: []cli.Flag{
				cfgPathsFlag(),
				&cli.BoolFlag{Name: "products", Usage: "reindex product variants only"},
				&cli.BoolFlag{Name: "attributes", Usage: "reindex attributes and attribute groups only"},
				&cli.BoolFlag{Name: "cities", Usage: "reindex cities only"},
				&cli.BoolFlag{Name: "category-paths", Usage: "rebuild the category_paths closure table before reindexing"},
			},
			Action: func(ctx context.Context, cl *cli.Command) error {
				conf, err := loadConfig(cl.Args().Slice(), cl.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				loggerOpts := append(defaultLoggerOpts(appName, currentAppVersion), logger.WithConfig(conf.Log))
				l := logger.NewExtended(loggerOpts...)
				defer func() { _ = l.Sync() }()

				target := app.ReindexTarget{
					Products:      cl.Bool("products"),
					Attributes:    cl.Bool("attributes"),
					Cities:        cl.Bool("cities"),
					CategoryPaths: cl.Bool("category-paths"),
				}
				// no explicit index flag -> reindex every index (category-paths stays opt-in)
				if !target.Products && !target.Attributes && !target.Cities {
					all := app.AllReindexTargets()
					target.Products, target.Attributes, target.Cities = all.Products, all.Attributes, all.Cities
				}

				return app.Reindex(ctx, conf, l, target)
			},
		},
		{
			Name:        "rebuild-category-paths",
			Description: "wipe and recompute the category_paths closure table from categories.parent_id",
			Flags:       []cli.Flag{cfgPathsFlag()},
			Action: func(ctx context.Context, cl *cli.Command) error {
				conf, err := loadConfig(cl.Args().Slice(), cl.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				loggerOpts := append(defaultLoggerOpts(appName, currentAppVersion), logger.WithConfig(conf.Log))
				l := logger.NewExtended(loggerOpts...)
				defer func() { _ = l.Sync() }()

				return app.Reindex(ctx, conf, l, app.ReindexTarget{CategoryPaths: true})
			},
		},
	}
}
