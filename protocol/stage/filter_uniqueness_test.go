package stage

import (
	"context"
	"github.com/imulab/go-scim/core"
	"github.com/imulab/go-scim/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewUniquenessFilter(t *testing.T) {
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
		prepare     func(t *testing.T, p persistence.Provider)
		getResource func(t *testing.T) *core.Resource
		getProperty func(t *testing.T, resource *core.Resource) core.Property
		assert      func(t *testing.T, resource *core.Resource, property core.Property, err error)
	}{
		{
			name: "unique value will pass",
			prepare: func(t *testing.T, p persistence.Provider) {
				return
			},
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("id"), "802ed529-1d9d-4855-82b2-11e9bbcb65d8")
				require.Nil(t, err)
				err = resource.Replace(core.Steps.NewPath("userName"), "foobar")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				userNameProp, err := core.NewNavigator(resource).Focus("userName")
				require.Nil(t, err)
				require.Equal(t, core.UniquenessServer, userNameProp.Attribute().Uniqueness)
				return userNameProp
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "non-unique value will fail",
			prepare: func(t *testing.T, p persistence.Provider) {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("id"), "802ed529-1d9d-4855-82b2-11e9bbcb65d8")
				require.Nil(t, err)
				err = resource.Replace(core.Steps.NewPath("userName"), "foobar")
				require.Nil(t, err)
				err = p.InsertOne(context.Background(), resource)
				require.Nil(t, err)
				return
			},
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("id"), "5cab2734-0198-462c-8444-78982083deea")
				require.Nil(t, err)
				err = resource.Replace(core.Steps.NewPath("userName"), "foobar")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				userNameProp, err := core.NewNavigator(resource).Focus("userName")
				require.Nil(t, err)
				require.Equal(t, core.UniquenessServer, userNameProp.Attribute().Uniqueness)
				return userNameProp
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, each := range tests {
		t.Run(each.name, func(t *testing.T) {
			var (
				provider = persistence.NewMemoryProvider()
				filter   = NewUniquenessFilter([]persistence.Provider{provider})
				resource = each.getResource(t)
				property = each.getProperty(t, resource)
				err      error
			)
			each.prepare(t, provider)
			assert.True(t, filter.Supports(property.Attribute()))
			err = filter.FilterOnCreate(context.Background(), resource, property)
			each.assert(t, resource, property, err)
		})
	}
}
