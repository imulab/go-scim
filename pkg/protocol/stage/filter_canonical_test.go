package stage

import (
	"context"
	"github.com/imulab/go-scim/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewCanonicalValueFilter(t *testing.T) {
	var (
		resourceType *core.ResourceType
	)
	{
		_ = core.Schemas.MustLoad("../../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../../resource/resource_type/user_resource_type.json")
	}

	tests := []struct {
		name        string
		getResource func(t *testing.T) *core.Resource
		getProperty func(t *testing.T, resource *core.Resource) core.Property
		assert      func(t *testing.T, resource *core.Resource, property core.Property, err error)
	}{
		{
			name: "canonical value check passes if conform to it",
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("userType"), "Contractor")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				userTypeProp, err := core.NewNavigator(resource).Focus("userType")
				require.Nil(t, err)
				return userTypeProp
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "canonical value check fails if illegal value",
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("userType"), "NotLegalValue")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				userTypeProp, err := core.NewNavigator(resource).Focus("userType")
				require.Nil(t, err)
				return userTypeProp
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, each := range tests {
		t.Run(each.name, func(t *testing.T) {
			var (
				filter   = NewCanonicalValueFilter(0)
				resource = each.getResource(t)
				property = each.getProperty(t, resource)
				err      error
			)
			assert.True(t, filter.Supports(property.Attribute()))

			err = filter.FilterOnCreate(context.Background(), resource, property)
			each.assert(t, resource, property, err)
			err = filter.FilterOnUpdate(context.Background(), resource, property, nil, nil)
			each.assert(t, resource, property, err)
		})
	}
}
