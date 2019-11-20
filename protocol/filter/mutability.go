package filter

import (
	"context"
	"github.com/imulab/go-scim/core"
	"github.com/imulab/go-scim/query"
)

type (
	mutabilityFilter struct{}
)

func (f *mutabilityFilter) Supports(attribute *core.Attribute) bool {
	switch attribute.Mutability {
	case core.MutabilityReadOnly:
		return !ContainsAnnotation(attribute, "@skipReadOnly")
	case core.MutabilityImmutable:
		return true
	default:
		return false
	}
}

func (f *mutabilityFilter) Order(attribute *core.Attribute) int {
	return 100
}

func (f *mutabilityFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource) error {
	switch property.Attribute().Mutability {
	case core.MutabilityReadOnly:
		return f.filterReadOnly(ctx, resource, property, ref)
	case core.MutabilityImmutable:
		return f.filterReadOnly(ctx, resource, property, ref)
	default:
		return nil
	}
}

func (f *mutabilityFilter) filterReadOnly(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource) error {
	if ref == nil {
		// When there's no reference, clear any existing value on readOnly properties.
		return property.(core.Crud).Delete(nil)
	}

	// When there's reference, replace any existing value with value's from reference
	var (
		value interface{}
	)
	{
		metadata := core.Meta.Get(property.Attribute().Id, core.DefaultMetadataId)
		head, err := query.CompilePath(metadata.(*core.DefaultMetadata).Path, false)
		if err != nil {
			return f.wrapError(err, property.Attribute())
		}
		value, err = ref.Get(head)
		if err != nil {
			return f.wrapError(err, property.Attribute())
		}
	}

	return property.(core.Crud).Replace(nil, value)
}

func (f *mutabilityFilter) filterImmutable(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource) error {
	if ref == nil {
		// When there's no reference, any value is allowed
		return nil
	}

	// When there's reference, replace any existing value with value's from reference
}

func (f *mutabilityFilter) wrapError(err error, attribute *core.Attribute) error {
	return core.Errors.Internal("failed to check read only on '%s': %s", attribute.DisplayName(), err.Error())
}
