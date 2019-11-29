package logs

import "github.com/imulab/go-scim/src/protocol"

// Returns a LogProvider that does nothing.
func NoOp() protocol.LogProvider {
	return noProvider{}
}

type noProvider struct {}

func (n noProvider) Info(format string, args ...interface{}) {}

func (n noProvider) Debug(format string, args ...interface{}) {}

func (n noProvider) Error(format string, args ...interface{}) {}

func (n noProvider) Warning(format string, args ...interface{}) {}

func (n noProvider) Fatal(format string, args ...interface{}) {}

