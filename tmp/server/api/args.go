package api

import (
	"github.com/imulab/go-scim/server/args"
	"github.com/urfave/cli/v2"
)

type arguments struct {
	*args.Scim
	*args.Memory
	*args.Mongo
	*args.Rabbit
	*args.Log
	// http
	httpPort int
}

func (arg *arguments) Flags() []cli.Flag {
	flags := []cli.Flag{
		&cli.IntFlag{
			Name:        "port",
			Aliases:     []string{"p"},
			Usage:       "HTTP port that the server listens on",
			EnvVars:     []string{"HTTP_PORT"},
			Value:       8080,
			Destination: &arg.httpPort,
		},
	}
	flags = append(flags, arg.Scim.Flags()...)
	flags = append(flags, arg.Memory.Flags()...)
	flags = append(flags, arg.Mongo.Flags()...)
	flags = append(flags, arg.Rabbit.Flags()...)
	flags = append(flags, arg.Log.Flags()...)
	return flags
}

