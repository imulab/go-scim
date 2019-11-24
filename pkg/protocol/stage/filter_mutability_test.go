package stage

import (
	"context"
	"github.com/imulab/go-scim/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewMutabilityFilter(t *testing.T) {
	var (
		resourceType *core.ResourceType
	)
	{
		_ = core.Schemas.MustLoad("../../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../../resource/resource_type/user_resource_type.json")
	}

	tests := []struct {
		name           string
		getResource    func(t *testing.T) *core.Resource
		getProperty    func(t *testing.T, resource *core.Resource) core.Property
		getRef         func(t *testing.T) *core.Resource
		getRefProperty func(t *testing.T, ref *core.Resource) core.Property
		assert         func(t *testing.T, resource *core.Resource, property core.Property, err error)
	}{
		{
			name: "readOnly property without reference is cleared",
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("meta"), map[string]interface{}{
					"version": "foobar",
				})
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				metaProp, err := core.NewNavigator(resource).Focus("meta")
				require.Nil(t, err)
				require.False(t, metaProp.IsUnassigned())
				require.Equal(t, core.MutabilityReadOnly, metaProp.Attribute().Mutability)
				return metaProp
			},
			getRef: func(t *testing.T) *core.Resource {
				return nil
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				return nil
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.True(t, property.IsUnassigned())
			},
		},
		{
			name: "readOnly property with reference is directly replaced",
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("id"), "changed")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				idProp, err := core.NewNavigator(resource).Focus("id")
				require.Nil(t, err)
				require.Equal(t, core.MutabilityReadOnly, idProp.Attribute().Mutability)
				return idProp
			},
			getRef: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("id"), "original")
				require.Nil(t, err)
				return resource
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				require.NotNil(t, ref)
				idProp, err := core.NewNavigator(ref).Focus("id")
				require.Nil(t, err)
				require.Equal(t, core.MutabilityReadOnly, idProp.Attribute().Mutability)
				return idProp
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "original", property.Raw())
			},
		},
		{
			name: "immutable property without reference is ignored",
			getResource: func(t *testing.T) *core.Resource {
				// Here we cheat the white box a bit by not providing a resource because no
				// schema with immutable attribute is readily available. But because resource
				// is not used anyways, we will be fine.
				return nil
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				return core.Properties.NewStringOf(&core.Attribute{
					Type:       core.TypeString,
					Mutability: core.MutabilityImmutable,
				}, "foobar")
			},
			getRef: func(t *testing.T) *core.Resource {
				return nil
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				return nil
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foobar", property.Raw())
			},
		},
		{
			name: "immutable property with reference is compared against",
			getResource: func(t *testing.T) *core.Resource {
				// Here we cheat the white box a bit by not providing a resource because no
				// schema with immutable attribute is readily available. But because resource
				// is not used anyways, we will be fine.
				return nil
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				core.Meta.AddSingle(core.DefaultMetadataId, &core.DefaultMetadata{
					Id:   "4052e48c-e77f-4893-8f48-e2e6ecc47fb2",
					Path: "foobar",
				})
				return core.Properties.NewStringOf(&core.Attribute{
					Id:         "4052e48c-e77f-4893-8f48-e2e6ecc47fb2",
					Name:       "foobar",
					Type:       core.TypeString,
					Mutability: core.MutabilityImmutable,
				}, "foobar")
			},
			getRef: func(t *testing.T) *core.Resource {
				// Same cheat as above, but provide a non-nil resource to enter FilterWithRef
				return &core.Resource{}
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				return core.Properties.NewStringOf(&core.Attribute{
					Id:         "4052e48c-e77f-4893-8f48-e2e6ecc47fb2",
					Name:       "foobar",
					Type:       core.TypeString,
					Mutability: core.MutabilityImmutable,
				}, "original")
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, each := range tests {
		t.Run(each.name, func(t *testing.T) {
			var (
				filter      = NewMutabilityFilter()
				resource    = each.getResource(t)
				property    = each.getProperty(t, resource)
				ref         = each.getRef(t)
				refProperty = each.getRefProperty(t, ref)
				err         error
			)

			assert.True(t, filter.Supports(property.Attribute()))

			if ref == nil || refProperty == nil {
				err = filter.FilterOnCreate(context.Background(), resource, property)
			} else {
				err = filter.FilterOnUpdate(context.Background(), resource, property, ref, refProperty)
			}

			each.assert(t, resource, property, err)
		})
	}
}