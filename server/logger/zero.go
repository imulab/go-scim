package logger

import (
	"github.com/imulab/go-scim/protocol/log"
	"github.com/rs/zerolog"
	"os"
)

func Zero() log.Logger {
	return &zeroLogger{
		logger: zerolog.New(os.Stderr).With().Timestamp().Logger(),
	}
}

type zeroLogger struct {
	logger 	zerolog.Logger
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
