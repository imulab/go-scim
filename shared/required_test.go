package shared

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateRequired(t *testing.T) {
	for _, test := range []struct {
		getResource func(r *Resource) *Resource
		getSchema   func(sch *Schema) *Schema
		assertion   func(err error)
	}{
		{
			// correct
			func(r *Resource) *Resource {
				return r
			},
			func(sch *Schema) *Schema {
				return sch
			},
			func(err error) {
				assert.Nil(t, err)
			},
		},
		{
			// nil immutable is allowed
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
			func(err error) {
				assert.Nil(t, err)
			},
		},
		{
			// empty immutable is not allowed
			func(r *Resource) *Resource {
				r.Complex["id"] = ""
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
			func(err error) {
				assert.NotNil(t, err)
				assert.IsType(t, &MissingRequiredPropertyError{}, err)
				assert.Equal(t, "id", err.(*MissingRequiredPropertyError).Path)
			},
		},
		{
			// unassigned readOnly is allowed
			func(r *Resource) *Resource {
				r.Complex["meta"] = nil
				return r
			},
			func(sch *Schema) *Schema {
				p, err := NewPath("meta")
				require.Nil(t, err)
				metaAttr := sch.GetAttribute(p, false)
				require.NotNil(t, metaAttr)
				metaAttr.Required = true
				metaAttr.Mutability = ReadOnly
				return sch
			},
			func(err error) {
				assert.Nil(t, err)
			},
		},
		{
			// unassigned required throws error
			func(r *Resource) *Resource {
				r.Complex["userName"] = ""
				return r
			},
			func(sch *Schema) *Schema {
				p, err := NewPath("userName")
				require.Nil(t, err)
				userNameAttr := sch.GetAttribute(p, false)
				require.NotNil(t, userNameAttr)
				userNameAttr.Required = true
				userNameAttr.Mutability = ReadWrite
				return sch
			},
			func(err error) {
				assert.NotNil(t, err)
				assert.IsType(t, &MissingRequiredPropertyError{}, err)
				assert.Equal(t, fmt.Sprintf("%s:userName", UserUrn), err.(*MissingRequiredPropertyError).Path)
			},
		},
	} {
		sch, _, err := ParseSchema("../resources/tests/user_schema.json")
		require.Nil(t, err)
		require.NotNil(t, sch)

		r, _, err := ParseResource("../resources/tests/user_1.json")
		require.Nil(t, err)
		require.NotNil(t, r)

		ctx := context.Background()
		test.assertion(ValidateRequired(test.getResource(r), test.getSchema(sch), ctx))
	}
}
