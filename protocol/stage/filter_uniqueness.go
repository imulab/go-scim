package stage

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/core"
	"github.com/imulab/go-scim/persistence"
)

// Create an uniqueness filter. This filter is responsible for checking properties whose attribute's uniqueness constraint
// has value 'server'. It will make sure, for the given value, no other resource in the database has that value.
func NewUniquenessFilter(providers []persistence.Provider) PropertyFilter {
	return &uniquenessFilter{
		providers: providers,
	}
}

var (
	_ PropertyFilter = (*uniquenessFilter)(nil)
)

type uniquenessFilter struct {
	providers []persistence.Provider
}

func (f *uniquenessFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Uniqueness == core.UniquenessServer
}

func (f *uniquenessFilter) Order() int {
	return 203
}

func (f *uniquenessFilter) FilterOnCreate(ctx context.Context, resource *core.Resource, property core.Property) error {
	return f.unique(ctx, resource, property)
}

func (f *uniquenessFilter) FilterOnUpdate(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	return f.unique(ctx, resource, property)
}

func (f *uniquenessFilter) unique(ctx context.Context, resource *core.Resource, property core.Property) error {
	if property.Attribute().MultiValued || property.Attribute().Type == core.TypeComplex {
		// uniqueness check does not apply to multiValued or complex attribute
		return nil
	}

	if property.IsUnassigned() {
		// no need to check uniqueness for unassigned values
		return nil
	}

	var provider persistence.Provider
	{
		for _, each := range f.providers {
			if each.IsResourceTypeSupported(resource.GetResourceType()) && each.IsFilterSupported() {
				provider = each
				break
			}
		}
		// no provider configured to check uniqueness
		// silently exit
		if provider == nil {
			return nil
		}
	}

	var (
		id         string
		path       string
		scimFilter string
	)
	{
		path = core.Meta.Get(property.Attribute().Id, core.DefaultMetadataId).(*core.DefaultMetadata).Path
		if path == "id" {
			// special case: because we will use id in other queries
			scimFilter = fmt.Sprintf("id eq \"%s\"", property.Raw())
		} else {
			if v, err := resource.Get(core.Steps.NewPath("id")); err != nil {
				return core.Errors.Internal("failed to obtain id")
			} else if v == nil || len(v.(string)) == 0 {
				return core.Errors.Internal("invalid resource id")
			} else {
				id = v.(string)
			}

			switch property.Attribute().Type {
			case core.TypeString, core.TypeReference, core.TypeBinary, core.TypeDateTime:
				scimFilter = fmt.Sprintf("(id ne \"%s\") and (%s eq \"%s\")", id, path, property.Raw())
			case core.TypeInteger:
				scimFilter = fmt.Sprintf("(id ne \"%s\") and (%s eq %d)", id, path, property.Raw())
			case core.TypeDecimal:
				scimFilter = fmt.Sprintf("(id ne \"%s\") and (%s eq %f)", id, path, property.Raw())
			case core.TypeBoolean:
				scimFilter = fmt.Sprintf("(id ne \"%s\") and (%s eq %t)", id, path, property.Raw())
			default:
				panic("invalid attribute")
			}
		}
	}

	n, err := provider.Count(ctx, scimFilter)
	if err != nil {
		return core.Errors.Uniqueness("failed to check uniqueness for '%s': %s",
			property.Attribute().DisplayName(),
			err.Error(),
		)
	} else if n > 0 {
		return core.Errors.Uniqueness("value of '%s' does not satisfy constraint uniqueness=%s",
			property.Attribute().DisplayName(),
			property.Attribute().Uniqueness.String(),
		)
	}

	return nil
}