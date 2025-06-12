package console

import (
	"context"
	"fmt"
	"github.com/stickpro/go-store/internal/app"
	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/pkg/cfg"
	"github.com/stickpro/go-store/pkg/logger"
	"github.com/urfave/cli/v3"
)

const (
	defaultConfigPath = "configs/config.yaml"
)

func InitCommands(currentAppVersion, appName, commitHash string) []*cli.Command {
	return []*cli.Command{
		{
			Name:        "start",
			Description: "Go store server",
			Flags:       []cli.Flag{cfgPathsFlag()},
			Action: func(ctx context.Context, c *cli.Command) error {
				conf, err := loadConfig(c.Args().Slice(), c.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				loggerOpts := append(defaultLoggerOpts(appName, currentAppVersion), logger.WithConfig(conf.Log))

				l := logger.NewExtended(loggerOpts...)
				defer func() {
					_ = l.Sync()
				}()
				app.Run(ctx, conf, l)
				return nil
			},
		},
		{
			Name:        "config",
			Description: "validate, gen envs and flags for config",
			Commands:    prepareConfigCommands(),
		}, // config
		{
			Name:        "migrate",
			Description: "migration database schema",
			Flags:       []cli.Flag{cfgPathsFlag()},
			Commands:    prepareMigrationCommands(appName, currentAppVersion),
		}, // migrate
		{
			Name:        "geo",
			Description: "geo commands",
			Flags:       []cli.Flag{cfgPathsFlag()},
			Commands:    prepareGeoCommands(appName, currentAppVersion),
		},
	}
}

func loadConfig(args, configPaths []string) (*config.Config, error) {
	conf := new(config.Config)
	if err := cfg.Load(conf,
		cfg.WithLoaderConfig(cfg.Config{
			Args:       args,
			Files:      configPaths,
			MergeFiles: true,
		}),
	); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return conf, nil
}

func defaultLoggerOpts(appName, version string) []logger.Option {
	return []logger.Option{
		logger.WithAppName(appName),
		logger.WithAppVersion(version),
	}
}

func cfgPathsFlag() *cli.StringSliceFlag {
	return &cli.StringSliceFlag{
		Name:    "configs",
		Aliases: []string{"c"},
		Usage:   "allows you to use your own paths to configuration files, separated by commas (config.yaml,config.prod.yml,.env)",
		Value:   cli.NewStringSlice(defaultConfigPath).Value(),
	}
}

func prepareConfigCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "genenvs",
			Usage: "generate markdown for all envs and config yaml template",
			Action: func(ctx context.Context, _ *cli.Command) error {
				if err := cfg.GenerateMarkdown(new(config.Config), "ENVS.md"); err != nil {
					return fmt.Errorf("failed to generate markdown: %w", err)
				}

				if err := cfg.GenerateYamlTemplate(new(config.Config), "configs/config.template.yaml"); err != nil {
					return fmt.Errorf("failed to generate yaml template: %w", err)
				}
				return nil
			},
		},
		{
			Name:  "flags",
			Usage: "print available config flags",
			Action: func(_ context.Context, _ *cli.Command) error {
				res, err := cfg.GenerateFlags(new(config.Config))
				if err != nil {
					return err
				}

				fmt.Println(res)

				return nil
			},
		},
		{
			Name:  "validate",
			Usage: "validate config without starting the server",
			Flags: []cli.Flag{cfgPathsFlag()},
			Action: func(_ context.Context, cl *cli.Command) error {
				return cfg.ValidateConfig(
					new(config.Config),
					cfg.WithLoaderConfig(cfg.Config{
						Args:  cl.Args().Slice(),
						Files: cl.StringSlice("configs"),
					}),
				)
			},
		},
	}
}
