package protocol

import (
	"log"
	"time"
)

// Returns a LogProvider that uses Golang's native log package for logging. The logger also outputs the
// logging timestamp in RFC3339 format.
func DefaultLogger() LogProvider {
	return defaultLogger{}
}

// Returns a LogProvider that does nothing.
func NoOpLogger() LogProvider {
	return noOpLogger{}
}

type (
	// Service provider for logging.
	LogProvider interface {
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
	defaultLogger struct{}
	noOpLogger    struct{}
)

func (d defaultLogger) Info(format string, args ...interface{}) {
	log.Printf("[INFO] %s - "+format+"\n", append([]interface{}{time.Now().Format(time.RFC3339)}, args...))
}

func (d defaultLogger) Debug(format string, args ...interface{}) {
	log.Printf("[DEBUG] %s - "+format+"\n", append([]interface{}{time.Now().Format(time.RFC3339)}, args...))
}

func (d defaultLogger) Error(format string, args ...interface{}) {
	log.Printf("[ERROR] %s - "+format+"\n", append([]interface{}{time.Now().Format(time.RFC3339)}, args...))
}

func (d defaultLogger) Warning(format string, args ...interface{}) {
	log.Printf("[WARNING] %s - "+format+"\n", append([]interface{}{time.Now().Format(time.RFC3339)}, args...))
}

func (d defaultLogger) Fatal(format string, args ...interface{}) {
	log.Fatalf("[FATAL] %s - "+format+"\n", append([]interface{}{time.Now().Format(time.RFC3339)}, args...))
}

func (n noOpLogger) Info(format string, args ...interface{}) {}

func (n noOpLogger) Debug(format string, args ...interface{}) {}

func (n noOpLogger) Error(format string, args ...interface{}) {}

func (n noOpLogger) Warning(format string, args ...interface{}) {}

func (n noOpLogger) Fatal(format string, args ...interface{}) {}
