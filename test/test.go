package test

import "os"

// Return environment variable value with a default value in case environment variable does not exist.
func EnvOrDefault(envVar string, defaultVal string) string {
	env := os.Getenv(envVar)
	if len(env) == 0 {
		return defaultVal
	}
	return env
}

// Return true if the environment variable is not empty
func EnvExists(envVar string) bool {
	return len(os.Getenv(envVar)) > 0
}