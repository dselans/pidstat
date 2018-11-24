package main

import (
	"fmt"
	"os"

	"github.com/dselans/go-pidstat/deps"
	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/dselans/go-pidstat/api"
	"github.com/dselans/go-pidstat/util"
)

const (
	DefaultRunMode       = "web"
	DefaultListenAddress = ":8787"
)

var (
	sugar         *zap.SugaredLogger
	version       string
	listenAddress string
)

func init() {
	logger, err := util.CreateLogger(false, map[string]interface{}{"pkg": "main"})
	if err != nil {
		panic(fmt.Sprintf("unable to setup logger: %v", err))
	}

	sugar = logger.Sugar()
}

func main() {
	// Setup CLI app
	app := cli.NewApp()
	app.Name = "pidstat"
	app.Author = "daniel.selans@gmail.com"
	app.Action = func(c *cli.Context) error {
		return c.App.Command(DefaultRunMode).Run(c)
	}

	if version != "" {
		app.Version = version
	}

	app.Commands = []cli.Command{
		{
			Name:    "web",
			Aliases: []string{"w"},
			Usage:   "start pidstat in web mode",
			Action:  runWeb,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "address",
					Value:       DefaultListenAddress,
					Usage:       "bind the server to a specific server",
					Destination: &listenAddress,
				},
			},
		},
		{
			Name:    "cli",
			Aliases: []string{"c"},
			Usage:   "start pidstat in cli mode",
			Action:  runCLI,
		},
	}

	if err := app.Run(os.Args); err != nil {
		sugar.Fatal(err)
	}
}

// Launch the app in web mode
func runWeb(ctx *cli.Context) error {
	// Setup dependencies
	d, err := deps.New()
	if err != nil {
		sugar.Fatalf("unable to instantiate dependencies: %v", err)
	}

	// Setup API server
	a, err := api.New(listenAddress, ctx.App.Version, d)
	if err != nil {
		sugar.Fatalf("unable to instantiate API: %v", err)
	}

	// Run API server
	sugar.Fatal(a.Run())

	return nil
}

// Launch the app in CLI mode
func runCLI(ctx *cli.Context) error {
	sugar.Error("CLI mode not implemented yet")
	return nil
}
