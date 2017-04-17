package shared

// Common abstraction for property providers
type PropertySource interface {
	Get(key string) interface{}
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
}

// Common abstraction for logging providers
type Logger interface {
	Info(template string, args ...interface{})
	Debug(template string, args ...interface{})
	Error(template string, args ...interface{})
}
