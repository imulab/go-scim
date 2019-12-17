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

func (l *zeroLogger) Info(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l *zeroLogger) Debug(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

func (l *zeroLogger) Error(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

func (l *zeroLogger) Warning(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

func (l *zeroLogger) Fatal(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}
