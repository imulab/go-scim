package services

import (
	"encoding/json"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
	"github.com/imulab/go-scim/src/core/expr"
	scimJSON "github.com/imulab/go-scim/src/core/json"
	"github.com/imulab/go-scim/src/core/prop"
)

const (
	PatchOpSchema = "urn:ietf:params:scim:api:messages:2.0:PatchOp"

	// Patch operations
	OpAdd     = "add"
	OpReplace = "replace"
	OpRemove  = "remove"
)

type (
	PatchRequest struct {
		Schemas    []string         `json:"schemas"`
		Operations []PatchOperation `json:"Operations"`
	}
	PatchOperation struct {
		Op    string          `json:"op"`
		Path  string          `json:"path"`
		Value json.RawMessage `json:"value"`
	}
)

func (pr *PatchRequest) Validate() error {
	if len(pr.Schemas) != 1 && pr.Schemas[0] != PatchOpSchema {
		return errors.InvalidSyntax("patch request must describe payload with schema '%s'", PatchOpSchema)
	}

	for _, each := range pr.Operations {
		switch each.Op {
		case OpAdd:
			if len(each.Value) == 0 {
				return errors.InvalidSyntax("missing add operation value")
			}
		case OpReplace:
			if len(each.Value) == 0 {
				return errors.InvalidSyntax("missing replace operation value")
			}
		case OpRemove:
			if len(each.Path) == 0 {
				return errors.InvalidSyntax("path is required for remove operation")
			} else if len(each.Value) > 0 {
				return errors.InvalidSyntax("value should not be provided for remove operation")
			}
		default:
			return errors.InvalidSyntax("'%s' is not a valid patch operation", each.Op)
		}
	}

	return nil
}

func (po *PatchOperation) ParseValue(resource *prop.Resource) (interface{}, error) {
	head, err  := expr.CompilePath(po.Path)
	if err != nil {
		return nil, err
	}

	if head.IsPath() && head.Token() == resource.ResourceType().ID() {
		head = head.Next()
	}

	attr := po.getTargetAttribute(resource.SuperAttribute(), head)
	if attr == nil {
		return nil, errors.NoTarget("'%s' does not yield any target", po.Path)
	}

	p := prop.NewProperty(attr, nil)
	if err := scimJSON.DeserializeProperty(po.Value, p, po.Op == OpAdd); err != nil {
		return nil, err
	}

	return p.Raw(), nil
}

func (po *PatchOperation) getTargetAttribute(parentAttr *core.Attribute, cursor *expr.Expression) *core.Attribute {
	if cursor == nil {
		return parentAttr
	}

	if parentAttr == nil {
		return nil
	}

	if cursor.IsRootOfFilter() {
		return po.getTargetAttribute(parentAttr, cursor.Next())
	}

	return po.getTargetAttribute(parentAttr.SubAttributeForName(cursor.Token()), cursor.Next())
}