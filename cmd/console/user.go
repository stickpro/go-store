package console

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/stickpro/go-store/internal/service/user"
	"github.com/stickpro/go-store/internal/storage"
	"github.com/stickpro/go-store/pkg/logger"
)

func prepareUserCommands(appName, currentAppVersion string) []*cli.Command {
	emailFlag := &cli.StringFlag{Name: "email", Aliases: []string{"e"}, Usage: "account email", Required: true}
	passwordFlag := &cli.StringFlag{Name: "password", Aliases: []string{"p"}, Usage: "account password", Required: true}

	return []*cli.Command{
		{
			Name:        "create-admin",
			Description: "create an admin account (email + password login)",
			Flags:       []cli.Flag{emailFlag, passwordFlag, cfgPathsFlag()},
			Action: func(ctx context.Context, cl *cli.Command) error {
				return withUserService(ctx, cl, appName, currentAppVersion, func(svc user.IUserService) error {
					email := normalizeEmail(cl.String("email"))
					u, err := svc.CreateAdmin(ctx, email, cl.String("password"))
					if err != nil {
						return fmt.Errorf("create admin: %w", err)
					}
					fmt.Printf("admin created: %s (%s)\n", u.Email, u.ID)
					return nil
				})
			},
		},
		{
			Name:        "set-password",
			Description: "set or replace the password of an existing account",
			Flags:       []cli.Flag{emailFlag, passwordFlag, cfgPathsFlag()},
			Action: func(ctx context.Context, cl *cli.Command) error {
				return withUserService(ctx, cl, appName, currentAppVersion, func(svc user.IUserService) error {
					email := normalizeEmail(cl.String("email"))
					if err := svc.SetPassword(ctx, email, cl.String("password")); err != nil {
						return fmt.Errorf("set password: %w", err)
					}
					fmt.Printf("password updated for %s\n", email)
					return nil
				})
			},
		},
	}
}

func withUserService(ctx context.Context, cl *cli.Command, appName, version string, fn func(user.IUserService) error) error {
	conf, err := loadConfig(cl.Args().Slice(), cl.StringSlice("configs"))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	loggerOpts := append(defaultLoggerOpts(appName, version), logger.WithConfig(conf.Log))
	l := logger.NewExtended(loggerOpts...)
	defer func() { _ = l.Sync() }()

	st, err := storage.InitStore(ctx, conf)
	if err != nil {
		return fmt.Errorf("failed to init storage: %w", err)
	}
	defer func() { _ = st.Close() }()

	return fn(user.New(conf, l, st))
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
