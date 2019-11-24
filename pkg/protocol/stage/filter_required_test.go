package stage

import (
	"context"
	"github.com/imulab/go-scim/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewRequiredFilter(t *testing.T) {
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
			name: "passes required and present value",
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("userName"), "foobar")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				userNameProp, err := core.NewNavigator(resource).Focus("userName")
				require.Nil(t, err)
				require.True(t, userNameProp.Attribute().Required)
				return userNameProp
			},
			getRef: func(t *testing.T) *core.Resource {
				return nil
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				return nil
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "fails required but absent value",
			getResource: func(t *testing.T) *core.Resource {
				return core.Resources.New(resourceType)
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				userNameProp, err := core.NewNavigator(resource).Focus("userName")
				require.Nil(t, err)
				require.True(t, userNameProp.Attribute().Required)
				return userNameProp
			},
			getRef: func(t *testing.T) *core.Resource {
				return nil
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				return nil
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, each := range tests {
		t.Run(each.name, func(t *testing.T) {
			var (
				filter      = NewRequiredFilter(0)
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