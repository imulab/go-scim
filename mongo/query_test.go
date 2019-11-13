package mongo

import (
	"github.com/imulab/go-scim/test"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"testing"
)

// This test case does not verify much aside from that it doesn't produce error. This is as much as the unit test can
// reliably assert. The effect of the query shall be tested with integration test.
func TestTransformFilter(t *testing.T) {
	_ = test.MustParseSchema(t, "../resource/schema/test_object_schema.json", true)
	_ = test.MustParseSchemaCompanion(t, "../resource/companion/test_object_schema_companion.json", true)
	resourceType := test.MustParseResourceType(t, "../resource/resource_type/test_object_resource_type.json")

	tests := []struct {
		name   string
		filter string
		assert func(t *testing.T, val bsonx.Val, err error)
	}{
		{
			name:   "composite filter",
			filter: "name eq \"david\" and age gt 17",
			assert: func(t *testing.T, val bsonx.Val, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name:   "multiValue equality",
			filter: "tags eq \"foo\"",
			assert: func(t *testing.T, val bsonx.Val, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name:   "duplex path filter",
			filter: "courses.name sw \"10\"",
			assert: func(t *testing.T, val bsonx.Val, err error) {
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
