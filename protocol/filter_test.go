package protocol

import (
	"context"
	"github.com/imulab/go-scim/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewMutabilityFilter(t *testing.T) {
	var (
		resourceType *core.ResourceType
	)
	{
		_ = core.Schemas.MustLoad("../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../resource/resource_type/user_resource_type.json")
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
					Id:          "4052e48c-e77f-4893-8f48-e2e6ecc47fb2",
					Path:        "foobar",
				})
				return core.Properties.NewStringOf(&core.Attribute{
					Id: 		"4052e48c-e77f-4893-8f48-e2e6ecc47fb2",
					Name: 		"foobar",
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
					Id: 		"4052e48c-e77f-4893-8f48-e2e6ecc47fb2",
					Name: 		"foobar",
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
				err = filter.Filter(context.Background(), resource, property)
			} else {
				err = filter.FilterWithRef(context.Background(), resource, property, ref, refProperty)
			}

			each.assert(t, resource, property, err)
		})
	}
}

func TestNewRequiredFilter(t *testing.T) {
	var (
		resourceType *core.ResourceType
	)
	{
		_ = core.Schemas.MustLoad("../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../resource/resource_type/user_resource_type.json")
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
				filter      = NewRequiredFilter()
				resource    = each.getResource(t)
				property    = each.getProperty(t, resource)
				ref         = each.getRef(t)
				refProperty = each.getRefProperty(t, ref)
				err         error
			)

			assert.True(t, filter.Supports(property.Attribute()))

			if ref == nil || refProperty == nil {
				err = filter.Filter(context.Background(), resource, property)
			} else {
				err = filter.FilterWithRef(context.Background(), resource, property, ref, refProperty)
			}

			each.assert(t, resource, property, err)
		})
	}
}

func TestNewMetaVersionFilter(t *testing.T) {
	var (
		resourceType *core.ResourceType
	)
	{
		_ = core.Schemas.MustLoad("../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../resource/resource_type/user_resource_type.json")
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
			name: "assign version on new resource",
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("id"), "268982ab-f85d-45eb-8f73-920919b6e2a5")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				var (
					versionProp core.Property
					err         error
				)
				{
					navigator := core.NewNavigator(resource)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					versionProp, err = navigator.Focus("version")
					require.Nil(t, err)
				}
				return versionProp
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
			name: "assign new version when reference is present",
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("id"), "268982ab-f85d-45eb-8f73-920919b6e2a5")
				require.Nil(t, err)
				err = resource.Replace(core.Steps.NewPathChain("meta", "version"), "W/\"1\"")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				var (
					versionProp core.Property
					err         error
				)
				{
					navigator := core.NewNavigator(resource)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					versionProp, err = navigator.Focus("version")
					require.Nil(t, err)
				}
				return versionProp
			},
			getRef: func(t *testing.T) *core.Resource {
				return core.Resources.New(resourceType)
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				require.NotNil(t, ref)
				var (
					versionProp core.Property
					err         error
				)
				{
					navigator := core.NewNavigator(ref)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					versionProp, err = navigator.Focus("version")
					require.Nil(t, err)
				}
				return versionProp
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.NotEmpty(t, property.Raw())
				assert.NotEqual(t, "W/\"1\"", property.Raw())
			},
		},
	}

	for _, each := range tests {
		t.Run(each.name, func(t *testing.T) {
			var (
				filter      = NewMetaVersionFilter()
				resource    = each.getResource(t)
				property    = each.getProperty(t, resource)
				ref         = each.getRef(t)
				refProperty = each.getRefProperty(t, ref)
				err         error
			)

			assert.True(t, filter.Supports(property.Attribute()))

			if ref == nil || refProperty == nil {
				err = filter.Filter(context.Background(), resource, property)
			} else {
				err = filter.FilterWithRef(context.Background(), resource, property, ref, refProperty)
			}

			each.assert(t, resource, property, err)
		})
	}
}

func TestNewMetaLocationFilter(t *testing.T) {
	var (
		resourceType *core.ResourceType
	)
	{
		_ = core.Schemas.MustLoad("../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../resource/resource_type/user_resource_type.json")
	}

	tests := []struct {
		name           string
		locationFormat map[string]string
		getResource    func(t *testing.T) *core.Resource
		getProperty    func(t *testing.T, resource *core.Resource) core.Property
		getRef         func(t *testing.T) *core.Resource
		getRefProperty func(t *testing.T, ref *core.Resource) core.Property
		assert         func(t *testing.T, resource *core.Resource, property core.Property, err error)
	}{
		{
			name: "assign location on new resource",
			locationFormat: map[string]string{
				resourceType.Id: "https://test.org/%s",
			},
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPath("id"), "ac2fe73d-9b76-4e0f-a0bd-af2e61cd7969")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				var (
					locationProp core.Property
					err          error
				)
				{
					navigator := core.NewNavigator(resource)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					locationProp, err = navigator.Focus("location")
					require.Nil(t, err)
				}
				return locationProp
			},
			getRef: func(t *testing.T) *core.Resource {
				return nil
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				return nil
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "https://test.org/ac2fe73d-9b76-4e0f-a0bd-af2e61cd7969", property.Raw())
			},
		},
		{
			name: "filter is no op when reference is present",
			locationFormat: map[string]string{
				resourceType.Id: "https://test.org/%s",
			},
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPathChain("meta", "location"), "https://test.org/foobar")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				var (
					locationProp core.Property
					err          error
				)
				{
					navigator := core.NewNavigator(resource)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					locationProp, err = navigator.Focus("location")
					require.Nil(t, err)
				}
				return locationProp
			},
			getRef: func(t *testing.T) *core.Resource {
				return core.Resources.New(resourceType)
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				require.NotNil(t, ref)
				var (
					locationProp core.Property
					err          error
				)
				{
					navigator := core.NewNavigator(ref)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					locationProp, err = navigator.Focus("location")
					require.Nil(t, err)
				}
				return locationProp
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "https://test.org/foobar", property.Raw())
			},
		},
	}

	for _, each := range tests {
		t.Run(each.name, func(t *testing.T) {
			var (
				filter      = NewMetaLocationFilter(each.locationFormat)
				resource    = each.getResource(t)
				property    = each.getProperty(t, resource)
				ref         = each.getRef(t)
				refProperty = each.getRefProperty(t, ref)
				err         error
			)

			assert.True(t, filter.Supports(property.Attribute()))

			if ref == nil || refProperty == nil {
				err = filter.Filter(context.Background(), resource, property)
			} else {
				err = filter.FilterWithRef(context.Background(), resource, property, ref, refProperty)
			}

			each.assert(t, resource, property, err)
		})
	}
}

func TestNewMetaLastModifiedFilter(t *testing.T) {
	var (
		resourceType *core.ResourceType
	)
	{
		_ = core.Schemas.MustLoad("../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../resource/resource_type/user_resource_type.json")
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
			name: "assign lastModified time on new resource",
			getResource: func(t *testing.T) *core.Resource {
				return core.Resources.New(resourceType)
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				var (
					lastModifiedProp core.Property
					err              error
				)
				{
					navigator := core.NewNavigator(resource)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					lastModifiedProp, err = navigator.Focus("lastModified")
					require.Nil(t, err)
				}
				return lastModifiedProp
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
			name: "assign lastModified time on resource when reference is present",
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPathChain("meta", "lastModified"), "2019-11-21T14:30:00")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				var (
					lastModifiedProp core.Property
					err              error
				)
				{
					navigator := core.NewNavigator(resource)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					lastModifiedProp, err = navigator.Focus("lastModified")
					require.Nil(t, err)
				}
				return lastModifiedProp
			},
			getRef: func(t *testing.T) *core.Resource {
				return core.Resources.New(resourceType)
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				require.NotNil(t, ref)
				var (
					lastModifiedProp core.Property
					err              error
				)
				{
					navigator := core.NewNavigator(ref)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					lastModifiedProp, err = navigator.Focus("lastModified")
					require.Nil(t, err)
				}
				return lastModifiedProp
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.NotEmpty(t, property.Raw())
				assert.NotEqual(t, "2019-11-21T14:30:00", property.Raw())
			},
		},
	}

	for _, each := range tests {
		t.Run(each.name, func(t *testing.T) {
			var (
				filter      = NewMetaLastModifiedFilter()
				resource    = each.getResource(t)
				property    = each.getProperty(t, resource)
				ref         = each.getRef(t)
				refProperty = each.getRefProperty(t, ref)
				err         error
			)

			assert.True(t, filter.Supports(property.Attribute()))

			if ref == nil || refProperty == nil {
				err = filter.Filter(context.Background(), resource, property)
			} else {
				err = filter.FilterWithRef(context.Background(), resource, property, ref, refProperty)
			}

			each.assert(t, resource, property, err)
		})
	}
}

func TestNewMetaCreatedFilter(t *testing.T) {
	var (
		resourceType *core.ResourceType
	)
	{
		_ = core.Schemas.MustLoad("../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../resource/resource_type/user_resource_type.json")
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
			name: "assign creation time on new resource",
			getResource: func(t *testing.T) *core.Resource {
				return core.Resources.New(resourceType)
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				var (
					createdProp core.Property
					err         error
				)
				{
					navigator := core.NewNavigator(resource)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					createdProp, err = navigator.Focus("created")
					require.Nil(t, err)
				}
				return createdProp
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
				err := resource.Replace(core.Steps.NewPathChain("meta", "created"), "2019-11-21T14:30:00")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				var (
					createdProp core.Property
					err         error
				)
				{
					navigator := core.NewNavigator(resource)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					createdProp, err = navigator.Focus("created")
					require.Nil(t, err)
				}
				return createdProp
			},
			getRef: func(t *testing.T) *core.Resource {
				return core.Resources.New(resourceType)
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				require.NotNil(t, ref)
				var (
					createdProp core.Property
					err         error
				)
				{
					navigator := core.NewNavigator(ref)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					createdProp, err = navigator.Focus("created")
					require.Nil(t, err)
				}
				return createdProp
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "2019-11-21T14:30:00", property.Raw())
			},
		},
	}

	for _, each := range tests {
		t.Run(each.name, func(t *testing.T) {
			var (
				filter      = NewMetaCreatedFilter()
				resource    = each.getResource(t)
				property    = each.getProperty(t, resource)
				ref         = each.getRef(t)
				refProperty = each.getRefProperty(t, ref)
				err         error
			)

			assert.True(t, filter.Supports(property.Attribute()))

			if ref == nil || refProperty == nil {
				err = filter.Filter(context.Background(), resource, property)
			} else {
				err = filter.FilterWithRef(context.Background(), resource, property, ref, refProperty)
			}

			each.assert(t, resource, property, err)
		})
	}
}

func TestNewMetaResourceTypeFilter(t *testing.T) {
	var (
		resourceType *core.ResourceType
	)
	{
		_ = core.Schemas.MustLoad("../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../resource/resource_type/user_resource_type.json")
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
			name: "assign resource type on new resource",
			getResource: func(t *testing.T) *core.Resource {
				return core.Resources.New(resourceType)
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				var (
					resourceTypeProp core.Property
					err              error
				)
				{
					navigator := core.NewNavigator(resource)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					resourceTypeProp, err = navigator.Focus("resourceType")
					require.Nil(t, err)
				}
				return resourceTypeProp
			},
			getRef: func(t *testing.T) *core.Resource {
				return nil
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				return nil
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, resourceType.Name, property.Raw())
			},
		},
		{
			name: "filter is no-op when reference is present",
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(resourceType)
				err := resource.Replace(core.Steps.NewPathChain("meta", "resourceType"), "FooBar")
				require.Nil(t, err)
				return resource
			},
			getProperty: func(t *testing.T, resource *core.Resource) core.Property {
				require.NotNil(t, resource)
				var (
					resourceTypeProp core.Property
					err              error
				)
				{
					navigator := core.NewNavigator(resource)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					resourceTypeProp, err = navigator.Focus("resourceType")
					require.Nil(t, err)
				}
				return resourceTypeProp
			},
			getRef: func(t *testing.T) *core.Resource {
				return core.Resources.New(resourceType)
			},
			getRefProperty: func(t *testing.T, ref *core.Resource) core.Property {
				require.NotNil(t, ref)
				var (
					resourceTypeProp core.Property
					err              error
				)
				{
					navigator := core.NewNavigator(ref)
					_, err = navigator.Focus("meta")
					require.Nil(t, err)
					resourceTypeProp, err = navigator.Focus("resourceType")
					require.Nil(t, err)
				}
				return resourceTypeProp
			},
			assert: func(t *testing.T, resource *core.Resource, property core.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "FooBar", property.Raw())
			},
		},
	}

	for _, each := range tests {
		t.Run(each.name, func(t *testing.T) {
			var (
				filter      = NewMetaResourceTypeFilter()
				resource    = each.getResource(t)
				property    = each.getProperty(t, resource)
				ref         = each.getRef(t)
				refProperty = each.getRefProperty(t, ref)
				err         error
			)

			assert.True(t, filter.Supports(property.Attribute()))

			if ref == nil || refProperty == nil {
				err = filter.Filter(context.Background(), resource, property)
			} else {
				err = filter.FilterWithRef(context.Background(), resource, property, ref, refProperty)
			}

			each.assert(t, resource, property, err)
		})
	}
}

func TestNewIDFilter(t *testing.T) {
	var (
		resourceType *core.ResourceType
	)
	{
		_ = core.Schemas.MustLoad("../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../resource/resource_type/user_resource_type.json")
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
				err = filter.Filter(context.Background(), resource, property)
			} else {
				err = filter.FilterWithRef(context.Background(), resource, property, ref, refProperty)
			}

			each.assert(t, resource, property, err)
		})
	}
}
