package validation

import (
	. "github.com/davidiamyou/go-scim/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestValidateMutability(t *testing.T) {
	for _, test := range []struct {
		getResource  func(r *Resource) *Resource
		getReference func(r *Resource) *Resource
		getSchema    func(sch *Schema) *Schema
		assertion    func(subj, ref *Resource, err error)
	}{
		{
			// correct (nothing changed)
			func(r *Resource) *Resource {
				return r
			},
			func(r *Resource) *Resource {
				return r
			},
			func(sch *Schema) *Schema {
				return sch
			},
			func(subj, ref *Resource, err error) {
				assert.Nil(t, err)
			},
		},
		{
			// correct (with readOnly copied over)
			func(r *Resource) *Resource {
				r.Complex["id"] = ""
				return r
			},
			func(r *Resource) *Resource {
				return r
			},
			func(sch *Schema) *Schema {
				p, err := NewPath("id")
				require.Nil(t, err)
				idAttr := sch.GetAttribute(p, false)
				require.NotNil(t, idAttr)
				idAttr.Required = true
				idAttr.Mutability = ReadOnly
				return sch
			},
			func(subj, ref *Resource, err error) {
				assert.Nil(t, err)
				assert.Equal(t, subj.Complex["id"], ref.Complex["id"])
			},
		},
		{
			// immutable missing
			func(r *Resource) *Resource {
				r.Complex["id"] = ""
				return r
			},
			func(r *Resource) *Resource {
				return r
			},
			func(sch *Schema) *Schema {
				p, err := NewPath("id")
				require.Nil(t, err)
				idAttr := sch.GetAttribute(p, false)
				require.NotNil(t, idAttr)
				idAttr.Required = true
				idAttr.Mutability = Immutable
				return sch
			},
			func(subj, ref *Resource, err error) {
				assert.NotNil(t, err)
				assert.IsType(t, &MutabilityViolationError{}, err)
				assert.Equal(t, "id", err.(*MutabilityViolationError).Path)
			},
		},
		{
			// immutable newly assigned
			func(r *Resource) *Resource {
				return r
			},
			func(r *Resource) *Resource {
				r.Complex["id"] = nil
				return r
			},
			func(sch *Schema) *Schema {
				p, err := NewPath("id")
				require.Nil(t, err)
				idAttr := sch.GetAttribute(p, false)
				require.NotNil(t, idAttr)
				idAttr.Required = true
				idAttr.Mutability = Immutable
				return sch
			},
			func(subj, ref *Resource, err error) {
				assert.Nil(t, err)
			},
		},
		{
			// complex nested (with readOnly copied over)
			func(r *Resource) *Resource {
				r.Complex["name"].(map[string]interface{})["formatted"] = "foo"
				return r
			},
			func(r *Resource) *Resource {
				r.Complex["name"].(map[string]interface{})["formatted"] = "bar"
				return r
			},
			func(sch *Schema) *Schema {
				p, err := NewPath("name.formatted")
				require.Nil(t, err)
				formattedAttr := sch.GetAttribute(p, true)
				require.NotNil(t, formattedAttr)
				formattedAttr.Mutability = ReadOnly
				return sch
			},
			func(subj, ref *Resource, err error) {
				assert.Nil(t, err)
				sv := subj.Complex["name"].(map[string]interface{})["formatted"]
				rv := ref.Complex["name"].(map[string]interface{})["formatted"]
				assert.Equal(t, rv, sv)
			},
		},
		{
			// readWrite complex array with readOnly sub attributes (advanced)
			func(r *Resource) *Resource {
				r.Complex["groups"] = []interface{}{
					map[string]interface{}{
						"value":   "A",
						"display": "A~",
					},
					map[string]interface{}{
						"value":   "D",
						"display": "D",
					},
					map[string]interface{}{
						"value":   "C",
						"display": "C",
					},
				}
				return r
			},
			func(r *Resource) *Resource {
				r.Complex["groups"] = []interface{}{
					map[string]interface{}{
						"value":   "A",
						"display": "A",
						"type":    "direct",
					},
					map[string]interface{}{
						"value":   "B",
						"display": "B",
						"type":    "direct",
					},
					map[string]interface{}{
						"value":   "C",
						"display": "C",
						"type":    "direct",
					},
				}
				return r
			},
			func(sch *Schema) *Schema {
				p, err := NewPath("groups")
				require.Nil(t, err)
				groupsAttr := sch.GetAttribute(p, true)
				require.NotNil(t, groupsAttr)
				groupsAttr.Mutability = ReadWrite
				return sch
			},
			func(subj, ref *Resource, err error) {
				assert.Nil(t, err)
				g := subj.Complex["groups"].([]interface{})
				assert.True(t, reflect.DeepEqual(g, []interface{}{
					map[string]interface{}{
						"value":   "A",
						"display": "A",
						"type":    "direct",
					},
					map[string]interface{}{
						"value":   "D",
						"display": "D",
					},
					map[string]interface{}{
						"value":   "C",
						"display": "C",
						"type":    "direct",
					},
				}))
			},
		},
	} {
		sch, _, err := ParseSchema("../resources/tests/user_schema.json")
		require.Nil(t, err)
		require.NotNil(t, sch)

		r0, _, err := ParseResource("../resources/tests/user_1.json")
		require.Nil(t, err)
		require.NotNil(t, r0)

		r1, _, err := ParseResource("../resources/tests/user_1.json")
		require.Nil(t, err)
		require.NotNil(t, r1)

		subj := test.getResource(r0)
		ref := test.getReference(r1)
		test.assertion(subj, ref, ValidateMutability(subj, ref, test.getSchema(sch)))
	}
}
