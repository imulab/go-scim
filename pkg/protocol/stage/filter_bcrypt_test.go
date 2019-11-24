package stage

import (
	"context"
	"github.com/imulab/go-scim/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewBCryptFilter(t *testing.T) {
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
		getRef      func(t *testing.T) *core.Resource
		getRefProp  func(t *testing.T, ref *core.Resource) core.Property
		assert      func(t *testing.T, resource *core.Resource, property core.Property, err error)
	}{
		{
			name: "password will be hashed",
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("password"), "s3cret")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				passwordProp, err := core.NewNavigator(resource).Focus("password")
				require.Nil(t, err)
				return passwordProp
			},
			getRef: func(t *testing.T) *core.Resource {
				return nil
			},
			getRefProp: func(t *testing.T, ref *core.Resource) core.Property {
				return nil
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.NotEqual(t, "s3cret", property.Raw())
			},
		},
		{
			name: "hashing is skipped with reference has not changed",
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("password"), "hashed_s3cret")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				passwordProp, err := core.NewNavigator(resource).Focus("password")
				require.Nil(t, err)
				return passwordProp
			},
			getRef: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("password"), "hashed_s3cret")
				require.Nil(t, err)
				return resource
			},
			getRefProp: func(t *testing.T, ref *core.Resource) core.Property {
				require.NotNil(t, ref)
				passwordProp, err := core.NewNavigator(ref).Focus("password")
				require.Nil(t, err)
				return passwordProp
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "hashed_s3cret", property.Raw())
			},
		},
		{
			name: "password is hashed when reference does not match",
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("password"), "new_s3cret")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				passwordProp, err := core.NewNavigator(resource).Focus("password")
				require.Nil(t, err)
				return passwordProp
			},
			getRef: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("password"), "hashed_s3cret")
				require.Nil(t, err)
				return resource
			},
			getRefProp: func(t *testing.T, ref *core.Resource) core.Property {
				require.NotNil(t, ref)
				passwordProp, err := core.NewNavigator(ref).Focus("password")
				require.Nil(t, err)
				return passwordProp
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.NotEmpty(t, property.Raw())
				assert.NotEqual(t, "new_s3cret", property.Raw())
			},
		},
	}

	for _, each := range tests {
		t.Run(each.name, func(t *testing.T) {
			var (
				filter   = NewBCryptFilter(10, 0)
				resource = each.getResource(t)
				property = each.getProperty(t, resource)
				ref      = each.getRef(t)
				refProp  = each.getRefProp(t, ref)
				err      error
			)
			assert.True(t, filter.Supports(property.Attribute()))

			if ref == nil || refProp == nil {
				err = filter.FilterOnCreate(context.Background(), resource, property)
			} else {
				err = filter.FilterOnUpdate(context.Background(), resource, property, ref, refProp)
			}
			each.assert(t, resource, property, err)
		})
	}
}
