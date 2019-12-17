package main

import (
	"github.com/imulab/go-scim/cmd/api"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:                   "scim",
		Usage:                  "Simple Cloud Identity Management",
		Commands: []*cli.Command{
			api.Command(),
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}