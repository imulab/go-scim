package args

import (
	"github.com/imulab/go-scim/protocol/log"
	"github.com/imulab/go-scim/server/logger"
	"github.com/urfave/cli/v2"
)

type Log struct {
	Level string
}

func (arg *Log) Logger() log.Logger {
	return logger.ZeroL(log.LevelOf(arg.Level))
}

func (arg *Log) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Usage:       "Logger level",
			EnvVars:     []string{"LOG_LEVEL"},
			Value:       "INFO",
			Destination: &arg.Level,
		},
	}
}
