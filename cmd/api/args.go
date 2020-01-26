package api

import (
	"github.com/imulab/go-scim/cmd/args"
	"github.com/urfave/cli/v2"
)

type arguments struct {
	args.Scim
	args.MemoryDB
	args.MongoDB
	args.RabbitMQ
	args.Logging
	HttpPort int
}

func (arg arguments) Flags() []cli.Flag {
	flags := []cli.Flag{
		&cli.IntFlag{
			Name:        "port",
			Aliases:     []string{"p"},
			Usage:       "HTTP port that the server listens on",
			EnvVars:     []string{"HTTP_PORT"},
			Value:       8080,
			Destination: &arg.HttpPort,
		},
	}
	flags = append(flags, arg.Scim.Flags()...)
	flags = append(flags, arg.MemoryDB.Flags()...)
	flags = append(flags, arg.MongoDB.Flags()...)
	flags = append(flags, arg.RabbitMQ.Flags()...)
	flags = append(flags, arg.Logging.Flags()...)
	return flags
}