package protocol

import (
	"context"
	"github.com/imulab/go-scim/core"
	"github.com/imulab/go-scim/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

type (
	// Dummy filter design to test the syncing between the visited property and its reference. This filter
	// does not consider the situation where refProp is nil and hence must be used in situation where
	// resource == ref.
	testSyncFilter struct {
		t *testing.T
	}
)

func TestFilterVisitorSync(t *testing.T) {
	var (
		resourceType *core.ResourceType
	)
	{
		_ = core.Schemas.MustLoad("../resource/schema/user_schema.json")
		_ = core.Meta.MustLoad("../resource/metadata/default_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../resource/resource_type/user_resource_type.json")
	}

	resource := test.MustResource("../resource/test/test_user_full.json", resourceType)
	ref := test.MustResource("../resource/test/test_user_full.json", resourceType)

	executor := NewFilterFunc([]*core.ResourceType{
		resourceType,
	}, []PropertyFilter{
		&testSyncFilter{t: t},
	})

	err := executor(context.Background(), resource, ref)
	assert.Nil(t, err)
}

func (f *testSyncFilter) Supports(attribute *core.Attribute) bool {
	return true
}

func (f *testSyncFilter) Order(attribute *core.Attribute) int {
	return 0
}

func (f *testSyncFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	return nil
}

func (f *testSyncFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	assert.True(f.t, property.Attribute().Equals(refProp.Attribute()))
	return nil
}
