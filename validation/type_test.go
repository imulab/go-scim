package validation

import (
	"fmt"
	. "github.com/davidiamyou/go-scim/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateType(t *testing.T) {
	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.Nil(t, err)
	require.NotNil(t, sch)

	for _, test := range []struct {
		getResource func(r *Resource) *Resource
		assertion   func(err error)
	}{
		{
			// correct
			func(r *Resource) *Resource {
				return r
			},
			func(err error) {
				assert.Nil(t, err)
			},
		},
		{
			// expected string
			func(r *Resource) *Resource {
				r.Complex["userName"] = 123
				return r
			},
			func(err error) {
				assert.NotNil(t, err)
				assert.IsType(t, &InvalidTypeError{}, err)
				assert.Equal(t, fmt.Sprintf("%s:userName", UserUrn), err.(*InvalidTypeError).Path)
			},
		},
		{
			// expected string (ignored)
			func(r *Resource) *Resource {
				r.Complex["id"] = 123
				return r
			},
			func(err error) {
				assert.Nil(t, err)
			},
		},
		{
			// expected string array
			func(r *Resource) *Resource {
				r.Complex["schemas"] = "foo"
				return r
			},
			func(err error) {
				assert.NotNil(t, err)
				assert.IsType(t, &InvalidTypeError{}, err)
				assert.Equal(t, "schemas", err.(*InvalidTypeError).Path)
			},
		},
		{
			// expected complex
			func(r *Resource) *Resource {
				r.Complex["name"] = "foo"
				return r
			},
			func(err error) {
				assert.NotNil(t, err)
				assert.IsType(t, &InvalidTypeError{}, err)
				assert.Equal(t, fmt.Sprintf("%s:name", UserUrn), err.(*InvalidTypeError).Path)
			},
		},
		{
			// expected complex array
			func(r *Resource) *Resource {
				r.Complex["emails"] = "foo"
				return r
			},
			func(err error) {
				assert.NotNil(t, err)
				assert.IsType(t, &InvalidTypeError{}, err)
				assert.Equal(t, fmt.Sprintf("%s:emails", UserUrn), err.(*InvalidTypeError).Path)
			},
		},
		{
			// expected string inside complex
			func(r *Resource) *Resource {
				p, err := NewPath("name.familyName")
				require.Nil(t, err)

				err = r.Set(p, 123, sch)
				require.Nil(t, err)

				return r
			},
			func(err error) {
				assert.NotNil(t, err)
				assert.IsType(t, &InvalidTypeError{}, err)
				assert.Equal(t, fmt.Sprintf("%s:name.familyName", UserUrn), err.(*InvalidTypeError).Path)
			},
		},
		{
			// no attribute
			func(r *Resource) *Resource {
				r.Complex["foo"] = "bar"
				return r
			},
			func(err error) {
				assert.NotNil(t, err)
				t.Log(err.Error())
				assert.IsType(t, &NoAttributeError{}, err)
				assert.Equal(t, "foo", err.(*NoAttributeError).Path)
			},
		},
	} {
		r, _, err := ParseResource("../resources/tests/user_1.json")
		require.Nil(t, err)
		require.NotNil(t, r)

		test.assertion(ValidateType(test.getResource(r), sch))
	}
}
