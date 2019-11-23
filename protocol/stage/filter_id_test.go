package stage

import (
	"context"
	"github.com/imulab/go-scim/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIDFilter(t *testing.T) {
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
			name: "generate id on new resource",
			getResource: func(t *testing.T) *core.Resource {
				return core.Resources.New(resourceType)
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				idProp, err := core.NewNavigator(resource).Focus("id")
				require.Nil(t, err)
				return idProp
			},
			getRef: func(t *testing.T) *core.Resource {
				return nil
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				return nil
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.NotEmpty(t, property.Raw())
			},
		},
		{
			name: "filter is no-op when reference is present",
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("id"), "foobar")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				idProp, err := core.NewNavigator(resource).Focus("id")
				require.Nil(t, err)
				return idProp
			},
			getRef: func(t *testing.T) *core.Resource {
				return core.Resources.New(resourceType)
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				require.NotNil(t, ref)
				idProp, err := core.NewNavigator(ref).Focus("id")
				require.Nil(t, err)
				return idProp
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "foobar", property.Raw())
			},
		},
	}

	for _, each := range tests {
		t.Run(each.name, func(t *testing.T) {
			var (
				filter      = NewIDFilter()
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

