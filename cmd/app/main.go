package main

import (
	"fmt"
	"github.com/stickpro/go-store/cmd/console"
	"github.com/urfave/cli/v2"
	"os"
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
	app := &cli.App{
		Name:                 appName,
		Description:          "Api for go store",
		Version:              getBuildVersion(),
		Suggest:              true,
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			cli.HelpFlag,
			cli.VersionFlag,
			cli.BashCompletionFlag,
		},
		Commands: console.InitCommands(version, appName, commitHash),
	}

	if err := app.Run(os.Args); err != nil {
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
