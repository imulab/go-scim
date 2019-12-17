package log

import (
	"log"
	"time"
)

// Returns a Logger that uses Golang's native log package for logging. The logger also outputs the
// logging timestamp in RFC3339 format.
func Default() Logger {
	return defaultLogger{}
}

type defaultLogger struct{}

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
