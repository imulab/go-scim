package log

// Returns a Logger that does nothing.
func None() Logger {
	return noOpLogger{}
}

type noOpLogger struct{}

func (n noOpLogger) Info(format string, args ...interface{}) {}

func (n noOpLogger) Debug(format string, args ...interface{}) {}

func (n noOpLogger) Error(format string, args ...interface{}) {}

func (n noOpLogger) Warning(format string, args ...interface{}) {}

func (n noOpLogger) Fatal(format string, args ...interface{}) {}
