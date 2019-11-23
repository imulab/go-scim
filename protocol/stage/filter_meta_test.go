package stage

import (
	"context"
	"github.com/imulab/go-scim/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestMetaFilters(t *testing.T) {
	suite.Run(t, new(MetaFiltersTestSuite))
}

type MetaFiltersTestSuite struct {
	suite.Suite
	resourceType *core.ResourceType
}

func (s *MetaFiltersTestSuite) SetupTest() {
	_ = core.Schemas.MustLoad("../../resource/schema/user_schema.json")
	_ = core.Meta.MustLoad("../../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
	s.resourceType = core.ResourceTypes.MustLoad("../../resource/resource_type/user_resource_type.json")
}

func (s *MetaFiltersTestSuite) TestMetaResourceTypeFilter() {
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
				return core.Resources.New(s.resourceType)
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
				assert.Equal(t, s.resourceType.Name, property.Raw())
			},
		},
		{
			name: "filter is no-op when reference is present",
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(s.resourceType)
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
				return core.Resources.New(s.resourceType)
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
		s.T().Run(each.name, func(t *testing.T) {
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
				err = filter.FilterOnCreate(context.Background(), resource, property)
			} else {
				err = filter.FilterOnUpdate(context.Background(), resource, property, ref, refProperty)
			}

			each.assert(t, resource, property, err)
		})
	}
}

func (s *MetaFiltersTestSuite) TestMetaCreatedFilter() {
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
				return core.Resources.New(s.resourceType)
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
				resource := core.Resources.New(s.resourceType)
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
				return core.Resources.New(s.resourceType)
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
		s.T().Run(each.name, func(t *testing.T) {
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
				err = filter.FilterOnCreate(context.Background(), resource, property)
			} else {
				err = filter.FilterOnUpdate(context.Background(), resource, property, ref, refProperty)
			}

			each.assert(t, resource, property, err)
		})
	}
}

func (s *MetaFiltersTestSuite) TestMetaLastModifiedFilter() {
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
				return core.Resources.New(s.resourceType)
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
				resource := core.Resources.New(s.resourceType)
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
				return core.Resources.New(s.resourceType)
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
		s.T().Run(each.name, func(t *testing.T) {
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
				err = filter.FilterOnCreate(context.Background(), resource, property)
			} else {
				err = filter.FilterOnUpdate(context.Background(), resource, property, ref, refProperty)
			}

			each.assert(t, resource, property, err)
		})
	}
}

func (s *MetaFiltersTestSuite) TestMetaLocationFilter() {
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
				s.resourceType.Id: "https://test.org/%s",
			},
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(s.resourceType)
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
				s.resourceType.Id: "https://test.org/%s",
			},
			getResource: func(t *testing.T) *core.Resource {
				resource := core.Resources.New(s.resourceType)
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
				return core.Resources.New(s.resourceType)
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
		s.T().Run(each.name, func(t *testing.T) {
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
				err = filter.FilterOnCreate(context.Background(), resource, property)
			} else {
				err = filter.FilterOnUpdate(context.Background(), resource, property, ref, refProperty)
			}

			each.assert(t, resource, property, err)
		})
	}
}

func (s *MetaFiltersTestSuite) TestMetaVersionFilter() {
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
				resource := core.Resources.New(s.resourceType)
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
				resource := core.Resources.New(s.resourceType)
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
				return core.Resources.New(s.resourceType)
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
		s.T().Run(each.name, func(t *testing.T) {
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
				err = filter.FilterOnCreate(context.Background(), resource, property)
			} else {
				err = filter.FilterOnUpdate(context.Background(), resource, property, ref, refProperty)
			}

			each.assert(t, resource, property, err)
		})
	}
}