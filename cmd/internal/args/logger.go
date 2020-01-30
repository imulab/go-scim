package args

import (
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"os"
)

// Logging is the configuration options related to logging
type Logging struct {
	Level string
}

func (arg *Logging) Logger() *zerolog.Logger {
	var level zerolog.Level
	switch arg.Level {
	case "INFO":
		level = zerolog.InfoLevel
	case "ERROR":
		level = zerolog.ErrorLevel
	case "DEBUG":
		level = zerolog.DebugLevel
	case "WARN":
		level = zerolog.WarnLevel
	case "FATAL":
		level = zerolog.FatalLevel
	default:
		level = zerolog.InfoLevel
	}

	l := zerolog.
		New(os.Stderr).
		Level(level).
		With().Timestamp().
		Logger()
	return &l
}

func (arg *Logging) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Usage:       "Specify logger output level to `[INFO|ERROR|DEBUG|WARN|FATAL]`. Value defaults `INFO`",
			EnvVars:     []string{"LOG_LEVEL"},
			Value:       "INFO",
			Destination: &arg.Level,
		},
	}
}
