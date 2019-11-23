package stage

import (
	"context"
	"github.com/imulab/go-scim/core"
	"github.com/imulab/go-scim/persistence"
	"github.com/imulab/go-scim/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilterIntegration(t *testing.T) {
	var resourceType *core.ResourceType
	{
		_ = core.Schemas.MustLoad("../../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../../resource/resource_type/user_resource_type.json")
	}

	tests := []struct {
		name        string
		prepare     func(t *testing.T, p persistence.Provider)
		getResource func(t *testing.T) *core.Resource
		getRef      func(t *testing.T) *core.Resource
		assert      func(t *testing.T, resource *core.Resource, err error)
	}{
		{
			name: "valid user resource on create",
			prepare: func(t *testing.T, p persistence.Provider) {
				return
			},
			getResource: func(t *testing.T) *core.Resource {
				return test.MustResource("../../resource/test/test_valid_user_create_payload.json", resourceType)
			},
			getRef: func(t *testing.T) *core.Resource {
				return nil
			},
			assert: func(t *testing.T, resource *core.Resource, err error) {
				assert.Nil(t, err)

				id, err := resource.GetID()
				assert.Nil(t, err)
				assert.NotEmpty(t, id)

				metaResourceType, err := resource.Get(core.Steps.NewPathChain("meta", "resourceType"))
				assert.Nil(t, err)
				assert.Equal(t, resourceType.Name, metaResourceType)

				metaCreated, err := resource.Get(core.Steps.NewPathChain("meta", "created"))
				assert.Nil(t, err)
				assert.NotEmpty(t, metaCreated)

				metaLastModified, err := resource.Get(core.Steps.NewPathChain("meta", "lastModified"))
				assert.Nil(t, err)
				assert.NotEmpty(t, metaLastModified)

				metaLocation, err := resource.Get(core.Steps.NewPathChain("meta", "location"))
				assert.Nil(t, err)
				assert.NotEmpty(t, metaLocation)

				metaVersion, err := resource.Get(core.Steps.NewPathChain("meta", "version"))
				assert.Nil(t, err)
				assert.NotEmpty(t, metaVersion)

				password, err := resource.Get(core.Steps.NewPath("password"))
				assert.Nil(t, err)
				assert.NotEmpty(t, password)
				assert.NotEqual(t, "s3cret", password)
			},
		},
	}

	for _, each := range tests {
		t.Run(each.name, func(t *testing.T) {
			provider := persistence.NewMemoryProvider(resourceType)
			each.prepare(t, provider)

			stage := NewFilterStage([]*core.ResourceType{
				resourceType,
			}, []PropertyFilter{
				NewIDFilter(),
				NewMetaResourceTypeFilter(),
				NewMetaCreatedFilter(),
				NewMetaLastModifiedFilter(),
				NewMetaLocationFilter(map[string]string{resourceType.Id: "https://test.org/%s"}),
				NewMetaVersionFilter(),
				NewMutabilityFilter(),
				NewRequiredFilter(),
				NewCanonicalValueFilter(),
				NewUniquenessFilter([]persistence.Provider{provider}),
				NewBCryptFilter(10),
			})

			resource := each.getResource(t)
			ref := each.getRef(t)
			err := stage(context.Background(), resource, ref)
			each.assert(t, resource, err)
		})
	}
}

func BenchmarkFilterIntegration(b *testing.B) {
	var resourceType *core.ResourceType
	{
		_ = core.Schemas.MustLoad("../../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../../resource/resource_type/user_resource_type.json")
	}

	provider := persistence.NewMemoryProvider(resourceType)

	stage := NewFilterStage([]*core.ResourceType{
		resourceType,
	}, []PropertyFilter{
		NewIDFilter(),
		NewMetaResourceTypeFilter(),
		NewMetaCreatedFilter(),
		NewMetaLastModifiedFilter(),
		NewMetaLocationFilter(map[string]string{resourceType.Id: "https://test.org/%s"}),
		NewMetaVersionFilter(),
		NewMutabilityFilter(),
		NewRequiredFilter(),
		NewCanonicalValueFilter(),
		NewUniquenessFilter([]persistence.Provider{provider}),
		NewBCryptFilter(10),
	})

	resource := test.MustResource("../../resource/test/test_valid_user_create_payload.json", resourceType)
	b.ResetTimer()
	_ = stage(context.Background(), resource, nil)
}

func TestFilterVisitorSync(t *testing.T) {
	var (
		resourceType *core.ResourceType
	)
	{
		_ = core.Schemas.MustLoad("../../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../../resource/resource_type/user_resource_type.json")
	}

	resource := test.MustResource("../../resource/test/test_user_full.json", resourceType)
	ref := test.MustResource("../../resource/test/test_user_full.json", resourceType)

	executor := NewFilterStage([]*core.ResourceType{
		resourceType,
	}, []PropertyFilter{
		&testSyncFilter{t: t},
	})

	err := executor(context.Background(), resource, ref)
	assert.Nil(t, err)
}

type (
	// Dummy filter design to test the syncing between the visited property and its reference. This filter
	// does not consider the situation where refProp is nil and hence must be used in situation where
	// resource == ref.
	testSyncFilter struct {
		t *testing.T
	}
)

func (f *testSyncFilter) Supports(attribute *core.Attribute) bool {
	return true
}

func (f *testSyncFilter) Order() int {
	return 0
}

func (f *testSyncFilter) FilterOnCreate(ctx context.Context, resource *core.Resource, property core.Property) error {
	return nil
}

func (f *testSyncFilter) FilterOnUpdate(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	assert.True(f.t, property.Attribute().Equals(refProp.Attribute()))
	return nil
}
