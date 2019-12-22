package logger

import (
	"github.com/imulab/go-scim/protocol/log"
	"github.com/rs/zerolog"
	"os"
)

func Zero() log.Logger {
	return ZeroL(log.LevelInfo)
}

func ZeroL(lvl log.Level) log.Logger {
	l := zerolog.New(os.Stderr).With().Timestamp().Logger()
	switch lvl {
	case log.LevelInfo:
		l = l.Level(zerolog.InfoLevel)
	case log.LevelDebug:
		l = l.Level(zerolog.DebugLevel)
	case log.LevelError:
		l = l.Level(zerolog.ErrorLevel)
	case log.LevelWarning:
		l = l.Level(zerolog.WarnLevel)
	case log.LevelFatal:
		l = l.Level(zerolog.FatalLevel)
	default:
		l = l.Level(zerolog.InfoLevel)
	}
	return &zeroLogger{logger: l}
}

type zeroLogger struct {
	logger zerolog.Logger
}

func (l *zeroLogger) Info(message string, args log.Args) {
	l.logger.Info().Fields(args).Msg(message)
}

func (l *zeroLogger) Debug(message string, args log.Args) {
	l.logger.Debug().Fields(args).Msg(message)
}

func (l *zeroLogger) Error(message string, args log.Args) {
	l.logger.Error().Fields(args).Msg(message)
}

func (l *zeroLogger) Warning(message string, args log.Args) {
	l.logger.Warn().Fields(args).Msg(message)
}

func (l *zeroLogger) Fatal(message string, args log.Args) {
	l.logger.Fatal().Fields(args).Msg(message)
}
