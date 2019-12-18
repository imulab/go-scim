package log

// Service provider for logging.
type Logger interface {
	// Log message with INFO level.
	Info(message string, args Args)
	// Log message with DEBUG level.
	Debug(message string, args Args)
	// Log message with ERROR level.
	Error(message string, args Args)
	// Log message with WARNING level.
	Warning(message string, args Args)
	// Log message with FATAL level.
	Fatal(message string, args Args)
}
