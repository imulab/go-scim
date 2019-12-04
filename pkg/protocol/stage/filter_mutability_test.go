package stage

import (
	"context"
	"github.com/imulab/go-scim/pkg/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMutabilityFilter(t *testing.T) {
	tests := []struct {
		name           string
		getResource    func(t *testing.T) *core.Resource
		getProperty    func(t *testing.T, resource *core.Resource) core.Property
		getRef         func(t *testing.T) *core.Resource
		getRefProperty func(t *testing.T, ref *core.Resource) core.Property
		assert         func(t *testing.T, resource *core.Resource, property core.Property, err error)
	}{
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
				filter      = NewMutabilityFilter(0)
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
