package log

import "strings"

type Level int

const (
	LevelFatal Level = iota
	LevelError
	LevelWarning
	LevelInfo
	LevelDebug
)

// Matches the log level of FATAL/ERROR/WARNING/INFO/DEBUG case insensitively.
// Defaults to INFO if mismatches.
func LevelOf(value string) Level {
	switch strings.ToUpper(value) {
	case "FATAL":
		return LevelFatal
	case "ERROR":
		return LevelError
	case "WARNING":
		return LevelWarning
	case "INFO":
		return LevelInfo
	case "DEBUG":
		return LevelDebug
	default:
		return LevelInfo
	}
}