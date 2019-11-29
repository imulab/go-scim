package logs

import (
	"github.com/imulab/go-scim/src/protocol"
	"log"
	"time"
)

// Returns a LogProvider that uses Golang's native log package for logging. The logger also outputs the
// logging timestamp in RFC3339 format.
func Default() protocol.LogProvider {
	return defaultProvider{}
}

type defaultProvider struct {}

func (d defaultProvider) Info(format string, args ...interface{}) {
	log.Printf("[INFO] %s - " + format, append([]interface{}{time.Now().Format(time.RFC3339)}, args...))
}

func (d defaultProvider) Debug(format string, args ...interface{}) {
	log.Printf("[DEBUG] %s - " + format, append([]interface{}{time.Now().Format(time.RFC3339)}, args...))
}

func (d defaultProvider) Error(format string, args ...interface{}) {
	log.Printf("[ERROR] %s - " + format, append([]interface{}{time.Now().Format(time.RFC3339)}, args...))
}

func (d defaultProvider) Warning(format string, args ...interface{}) {
	log.Printf("[WARNING] %s - " + format, append([]interface{}{time.Now().Format(time.RFC3339)}, args...))
}

func (d defaultProvider) Fatal(format string, args ...interface{}) {
	log.Fatalf("[FATAL] %s - " + format, append([]interface{}{time.Now().Format(time.RFC3339)}, args...))
}

