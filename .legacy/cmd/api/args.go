package api

import (
	"github.com/imulab/go-scim/cmd/internal/args"
	"github.com/urfave/cli/v2"
)

func newArgs() *arguments {
	return &arguments{
		Scim:     new(args.Scim),
		MemoryDB: new(args.MemoryDB),
		MongoDB:  new(args.MongoDB),
		RabbitMQ: new(args.RabbitMQ),
		Logging:  new(args.Logging),
	}
}

type arguments struct {
	*args.Scim
	*args.MemoryDB
	*args.MongoDB
	*args.RabbitMQ
	*args.Logging
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
	flags = append(flags, arg.MemoryDB.Flags()...)
	flags = append(flags, arg.MongoDB.Flags()...)
	flags = append(flags, arg.RabbitMQ.Flags()...)
	flags = append(flags, arg.Logging.Flags()...)
	return flags
}

func (arg *arguments) Initialize() *applicationContext {
	return &applicationContext{args: arg}
}
