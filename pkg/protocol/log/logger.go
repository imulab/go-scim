package log

// Service provider for logging.
type Logger interface {
	// Log message with INFO level.
	Info(format string, args ...interface{})
	// Log message with DEBUG level.
	Debug(format string, args ...interface{})
	// Log message with ERROR level.
	Error(format string, args ...interface{})
	// Log message with WARNING level.
	Warning(format string, args ...interface{})
	// Log message with FATAL level.
	Fatal(format string, args ...interface{})
}
