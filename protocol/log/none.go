package log

// Returns a Logger that does nothing.
func None() Logger {
	return noOpLogger{}
}

type noOpLogger struct{}

func (n noOpLogger) Info(message string, args Args) {}

func (n noOpLogger) Debug(message string, args Args) {}

func (n noOpLogger) Error(message string, args Args) {}

func (n noOpLogger) Warning(message string, args Args) {}

func (n noOpLogger) Fatal(message string, args Args) {}
