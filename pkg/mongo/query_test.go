package mongo

import (
	"github.com/imulab/go-scim/pkg/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

// This test case does not verify much aside from that it doesn't produce error. This is as much as the unit test can
// reliably assert. The effect of the query shall be tested with integration test.
func TestTransformFilter(t *testing.T) {
	var resourceType *core.ResourceType
	{
		_ = core.Schemas.MustLoad("../resource/schema/test_object_schema.json")
		_ = core.Meta.MustLoad("../resource/metadata/test_metadata.json", new(core.DefaultMetadataProvider))
		resourceType = core.ResourceTypes.MustLoad("../resource/resource_type/test_object_resource_type.json")
	}

	tests := []struct {
		name   string
		filter string
		assert func(t *testing.T, val interface{}, err error)
	}{
		{
			name:   "composite filter",
			filter: "name eq \"david\" and age gt 17",
			assert: func(t *testing.T, val interface{}, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name:   "multiValue equality",
			filter: "tags eq \"foo\"",
			assert: func(t *testing.T, val interface{}, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name:   "duplex path filter",
			filter: "courses.name sw \"10\"",
			assert: func(t *testing.T, val interface{}, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name:   "single cardinality filter",
			filter: "id pr",
			assert: func(t *testing.T, val interface{}, err error) {
				assert.Nil(t, err)
			},
		},
	}

	for _, each := range tests {
		t.Run(each.name, func(t *testing.T) {
			v, err := TransformFilter(each.filter, resourceType)
			each.assert(t, v, err)
		})

	}
}
