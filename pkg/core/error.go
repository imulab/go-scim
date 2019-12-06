package core

import "fmt"

// A SCIM error.
type Error struct {
	Status  int    `json:"status"`
	Type    string `json:"error_type"`
	Message string `json:"error_message"`
}

// Return the formatted error information for display.
func (s Error) Error() string {
	if len(s.Message) == 0 {
		return s.Type
	}
	return fmt.Sprintf("%s: %s", s.Type, s.Message)
}
