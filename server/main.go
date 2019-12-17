package main

import (
	"github.com/imulab/go-scim/server/api"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "scim",
		Usage: "Simple Cloud Identity Management",
		Commands: []*cli.Command{
			api.Command(),
		},
		HideVersion: true,
		Authors: []*cli.Author{
			{
				Name:  "Weinan Qiu",
				Email: "davidiamyou@gmail.com",
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
