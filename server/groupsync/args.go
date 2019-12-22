package groupsync

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
	requeueLimit int
}

func (arg *arguments) Flags() []cli.Flag {
	flags := []cli.Flag{
		&cli.IntFlag{
			Name:        "requeue-limit",
			Usage:       "Limit for message re-queues. Messages that has been re-queued more than this limit will be dropped (0 for unlimited).",
			EnvVars:     []string{"REQUEUE_LIMIT"},
			Destination: &arg.requeueLimit,
		},
	}
	flags = append(flags, arg.Scim.Flags()...)
	flags = append(flags, arg.Memory.Flags()...)
	flags = append(flags, arg.Mongo.Flags()...)
	flags = append(flags, arg.Rabbit.Flags()...)
	flags = append(flags, arg.Log.Flags()...)
	return flags
}