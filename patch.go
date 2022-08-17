package scim

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	URNPatchOp = "urn:ietf:params:scim:api:messages:2.0:PatchOp"

	opAdd     = "add"
	opReplace = "replace"
	opRemove  = "remove"
)

// PatchRequest is the payload of SCIM patch request. By parsing it through json.Unmarshal, this structure is validated.
// If parsed through other mechanisms, users are responsible for validation.
type PatchRequest struct {
	Schemas    []string
	Operations []*PatchOperation
}

func (r *PatchRequest) UnmarshalJSON(bytes []byte) error {
	var t struct {
		Schemas    []string          `json:"schemas"`
		Operations []*PatchOperation `json:"Operations"`
	}

	if err := json.Unmarshal(bytes, &t); err != nil {
		return fmt.Errorf("%w: invalid patch request", ErrInvalidSyntax)
	}

	switch {
	case len(t.Schemas) != 1 || t.Schemas[0] != URNPatchOp:
		return fmt.Errorf("%w: invalid patch request schema", ErrInvalidSyntax)
	case len(t.Operations) == 0:
		return fmt.Errorf("%w: at least one operation expected for patch request", ErrInvalidSyntax)
	}

	r.Schemas = t.Schemas
	r.Operations = t.Operations

	return nil
}

// PatchOperation is a single patch operation within the PatchRequest.
type PatchOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path,omitempty"`
	Value any    `json:"value,omitempty"`
}

func (r *PatchOperation) validate() error {
	switch strings.ToLower(r.Op) {
	case opAdd, opReplace:
		if r.Value == nil {
			return fmt.Errorf("%w: value required for %s operations", ErrInvalidSyntax, r.Op)
		}
		if len(r.Path) == 0 {
			if _, ok := r.Value.(map[string]any); !ok {
				return fmt.Errorf("%w: value must be a json object when path is absent", ErrInvalidValue)
			}
		}
		return nil
	case opRemove:
		if len(r.Path) == 0 {
			return fmt.Errorf("%w: path is required for remove operations", ErrNoTarget)
		}
		return nil
	default:
		return fmt.Errorf("%w: invalid patch op", ErrInvalidSyntax)
	}
}
