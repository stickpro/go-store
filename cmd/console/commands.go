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

func InitCommands(currentAppVersion, appName, _ string) []*cli.Command {
	return []*cli.Command{
		{
			Name:        "start",
			Description: "Go store server",
			Flags:       []cli.Flag{cfgPathsFlag(), kafkaGroupIDFlag()},
			Action: func(ctx context.Context, c *cli.Command) error {
				conf, err := loadConfig(c.Args().Slice(), c.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				if groupID := c.String("kafka-group-id"); groupID != "" {
					conf.Kafka.Consumer.GroupID = groupID
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
		{
			Name:        "search",
			Description: "search index commands",
			Flags:       []cli.Flag{cfgPathsFlag()},
			Commands:    prepareSearchCommands(appName, currentAppVersion),
		},
		{
			Name:        "media",
			Description: "media commands",
			Flags:       []cli.Flag{cfgPathsFlag()},
			Commands:    prepareMediaCommands(appName, currentAppVersion),
		},
		{
			Name:        "user",
			Description: "user account commands",
			Flags:       []cli.Flag{cfgPathsFlag()},
			Commands:    prepareUserCommands(appName, currentAppVersion),
		},
		{
			Name:        "import",
			Description: "one-off data import commands",
			Flags:       []cli.Flag{cfgPathsFlag()},
			Commands:    prepareImportCommands(appName, currentAppVersion),
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

// kafkaGroupIDFlag overrides kafka.consumer.group_id / KAFKA_CONSUMER_GROUP_ID.
// Use it for local runs against a shared broker so the process joins its own
// consumer group instead of splitting partitions with another deployment
// (e.g. production) using the default group id.
func kafkaGroupIDFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:  "kafka-group-id",
		Usage: "override the Kafka consumer group id (default from config/KAFKA_CONSUMER_GROUP_ID); set this to a unique value for local runs against a shared broker",
	}
}

func prepareConfigCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "genenvs",
			Usage: "generate markdown for all envs and config yaml template",
			Action: func(_ context.Context, _ *cli.Command) error {
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
