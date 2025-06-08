package main

import (
	"context"
	"fmt"
	"github.com/stickpro/go-store/cmd/console"
	g0store "github.com/stickpro/go-store/pkg/util/signal"
	"github.com/urfave/cli/v3"
	"os"
	"os/signal"
	"runtime"
)

var (
	appName    = "go-store"
	version    = "local"
	commitHash = "unknown"
)

//	@title			GO-store
//	@version		1.0
//	@description	This is an API for go-store

//	@contact.name	Vladislav B
//	@contact.email	go-store@stick.sh

// @BasePath					/api
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), g0store.Shutdown()...)
	defer cancel()

	app := &cli.Command{
		Name:        appName,
		Description: "Api for go store",
		Version:     getBuildVersion(),
		Suggest:     true,
		Flags: []cli.Flag{
			cli.HelpFlag,
			cli.VersionFlag,
		},
		Commands: console.InitCommands(version, appName, commitHash),
	}

	if err := app.Run(ctx, os.Args); err != nil {
		_, _ = fmt.Println(err.Error())
		os.Exit(1)
	}
}

func getBuildVersion() string {
	return fmt.Sprintf(
		"\n\nrelease: %s\ncommit hash: %s\ngo version: %s",
		version,
		commitHash,
		runtime.Version(),
	)
}
