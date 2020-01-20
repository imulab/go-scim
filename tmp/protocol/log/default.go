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

func (d defaultLogger) Info(message string, args Args) {
	log.Printf("[INFO] %s - %s %s\n", time.Now().Format(time.RFC3339), message, args.String())
}

func (d defaultLogger) Debug(message string, args Args) {
	log.Printf("[DEBUG] %s - %s %s\n", time.Now().Format(time.RFC3339), message, args.String())
}

func (d defaultLogger) Error(message string, args Args) {
	log.Printf("[ERROR] %s - %s %s\n", time.Now().Format(time.RFC3339), message, args.String())
}

func (d defaultLogger) Warning(message string, args Args) {
	log.Printf("[WARNING] %s - %s %s\n", time.Now().Format(time.RFC3339), message, args.String())
}

func (d defaultLogger) Fatal(message string, args Args) {
	log.Fatalf("[FATAL] %s - %s %s\n", time.Now().Format(time.RFC3339), message, args.String())
}
