package errors

import (
	"encoding/json"
	"fmt"
)

const (
	// Schema URN defined for SCIM error messages
	// https://tools.ietf.org/html/rfc7644#section-3.12
	Schema = "urn:ietf:params:scim:api:messages:2.0:Error"
)

// A SCIM error message. It is recommended to not directly create this structure, but to
// use constructors defined in factory.
type Error struct {
	Status  int
	Type    string
	Message string
}

// Marshal SCIM error message to JSON
func (s Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Schemas  []string `json:"schemas"`
		Status   string   `json:"status"`
		ScimType string   `json:"scimType"`
		Detail   string   `json:"detail"`
	}{
		Schemas:  []string{Schema},
		Status:   fmt.Sprintf("%d", s.Status),
		ScimType: s.Type,
		Detail:   s.Message,
	})
}

func (s Error) Error() string {
	if len(s.Message) == 0 {
		return s.Type
	}
	return s.Message
}

var (
	_ json.Marshaler = (*Error)(nil)
)
