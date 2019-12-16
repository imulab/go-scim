package services

import (
	"context"
	"encoding/json"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/expr"
	scimJSON "github.com/imulab/go-scim/pkg/core/json"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"github.com/imulab/go-scim/pkg/protocol/crud"
	"github.com/imulab/go-scim/pkg/protocol/db"
	"github.com/imulab/go-scim/pkg/protocol/event"
	"github.com/imulab/go-scim/pkg/protocol/log"
	"github.com/imulab/go-scim/pkg/protocol/services/filter"
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
		Schemas       []string                           `json:"schemas"`
		Operations    []PatchOperation                   `json:"Operations"`
		ResourceID    string                             `json:"-"`
		MatchCriteria func(resource *prop.Resource) bool `json:"-"`
	}
	PatchOperation struct {
		Op    string          `json:"op"`
		Path  string          `json:"path"`
		Value json.RawMessage `json:"value"`
	}
	PatchResponse struct {
		Resource   *prop.Resource
		Location   string
		OldVersion string
		NewVersion string
	}
	PatchService struct {
		Logger                log.Logger
		PrePatchFilters       []filter.ForResource
		PostPatchFilters      []filter.ForResource
		Database              db.DB
		ServiceProviderConfig *spec.ServiceProviderConfig
		Event                 event.Publisher
	}
)

func (s *PatchService) checkSupport() error {
	if !s.ServiceProviderConfig.Patch.Supported {
		return errors.InvalidRequest("patch is not supported")
	}
	return nil
}

func (s *PatchService) PatchResource(ctx context.Context, request *PatchRequest) (*PatchResponse, error) {
	if err := s.checkSupport(); err != nil {
		return nil, err
	}

	s.Logger.Debug("received patch request [id=%s]", request.ResourceID)

	if err := request.Validate(); err != nil {
		s.Logger.Error("patch request for [id=%s] is invalid: %s", request.ResourceID, err.Error())
		return nil, err
	}

	ref, err := s.Database.Get(ctx, request.ResourceID, nil)
	if err != nil {
		return nil, err
	}
	if s.ServiceProviderConfig.ETag.Supported && request.MatchCriteria != nil {
		if !request.MatchCriteria(ref) {
			return nil, errors.PreConditionFailed("resource [id=%s] does not meet pre condition", request.ResourceID)
		}
	}

	resource := ref.Clone()
	for _, f := range s.PrePatchFilters {
		if err := f.FilterRef(ctx, resource, ref); err != nil {
			s.Logger.Error("patch request encounter error during filter for resource [id=%s]: %s", request.ResourceID, err.Error())
			return nil, err
		}
	}

	for _, patchOp := range request.Operations {
		switch patchOp.Op {
		case OpAdd:
			if valueToAdd, err := patchOp.ParseValue(resource); err != nil {
				return nil, err
			} else if err := crud.Add(resource, patchOp.Path, valueToAdd); err != nil {
				return nil, err
			}
		case OpReplace:
			if valueToReplace, err := patchOp.ParseValue(resource); err != nil {
				return nil, err
			} else if err := crud.Replace(resource, patchOp.Path, valueToReplace); err != nil {
				return nil, err
			}
		case OpRemove:
			if err := crud.Delete(resource, patchOp.Path); err != nil {
				return nil, err
			}
		}
	}

	for _, f := range s.PostPatchFilters {
		if err := f.FilterRef(ctx, resource, ref); err != nil {
			s.Logger.Error("patch request encounter error during filter for resource [id=%s]: %s", request.ResourceID, err.Error())
			return nil, err
		}
	}

	// Only replace when version is bumped
	if resource.Version() != ref.Version() {
		err = s.Database.Replace(ctx, resource)
		if err != nil {
			s.Logger.Error("resource [id=%s] failed to save into persistence: %s", request.ResourceID, err.Error())
			return nil, err
		}
		s.Logger.Debug("resource [id=%s] saved in persistence", request.ResourceID)

		if s.Event != nil {
			s.Event.ResourceUpdated(ctx, resource, ref)
		}
	}

	return &PatchResponse{
		Resource:   resource,
		Location:   resource.Location(),
		OldVersion: ref.Version(),
		NewVersion: resource.Version(),
	}, nil
}

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
	var (
		head *expr.Expression
		err  error
	)
	{
		if len(po.Path) > 0 {
			head, err = expr.CompilePath(po.Path)
			if err != nil {
				return nil, err
			}
			if head.IsPath() && head.Token() == resource.ResourceType().ID() {
				head = head.Next()
			}
		}
	}

	attr := po.getTargetAttribute(resource.SuperAttribute(), head)
	if attr == nil {
		return nil, errors.NoTarget("'%s' does not yield any target", po.Path)
	}

	p := prop.New(attr, nil)
	if err := scimJSON.DeserializeProperty(po.Value, p, po.Op == OpAdd); err != nil {
		return nil, err
	}

	return p.Raw(), nil
}

func (po *PatchOperation) getTargetAttribute(parentAttr *spec.Attribute, cursor *expr.Expression) *spec.Attribute {
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
