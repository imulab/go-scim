package shared

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestCorrectCase(t *testing.T) {
	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.Nil(t, err)
	require.NotNil(t, sch)

	for _, test := range []struct {
		getResource func(r *Resource) *Resource
	}{
		{
			// identical
			func(r *Resource) *Resource {
				return r
			},
		},
		{
			// first level field capitalize
			func(r *Resource) *Resource {
				m := reflect.ValueOf(r.Complex)
				k0 := reflect.ValueOf("id")
				k1 := reflect.ValueOf("ID")
				v := m.MapIndex(k0)
				m.SetMapIndex(k1, v)
				m.SetMapIndex(k0, reflect.Value{})
				return r
			},
		},
		{
			// nested level field capitalize
			func(r *Resource) *Resource {
				m := reflect.ValueOf(r.Complex["name"])
				if m.Kind() == reflect.Interface {
					m = m.Elem()
				}

				k0 := reflect.ValueOf("familyName")
				k1 := reflect.ValueOf("FamilyName")
				v := m.MapIndex(k0)
				m.SetMapIndex(k1, v)
				m.SetMapIndex(k0, reflect.Value{})

				return r
			},
		},
		{
			// array nested field capitalize
			func(r *Resource) *Resource {
				s := reflect.ValueOf(r.Complex["emails"])
				if s.Kind() == reflect.Interface {
					s = s.Elem()
				}

				for i := 0; i < s.Len(); i++ {
					m := s.Index(i)
					if m.Kind() == reflect.Interface {
						m = m.Elem()
					}

					k0 := reflect.ValueOf("value")
					k1 := reflect.ValueOf("Value")
					v := m.MapIndex(k0)
					m.SetMapIndex(k1, v)
					m.SetMapIndex(k0, reflect.Value{})
				}

				return r
			},
		},
	} {
		r0, _, err := ParseResource("../resources/tests/user_1.json")
		require.Nil(t, err)
		require.NotNil(t, r0)

		r1, _, err := ParseResource("../resources/tests/user_1.json")
		require.Nil(t, err)
		require.NotNil(t, r1)
		r1 = test.getResource(r1)

		ctx := context.Background()
		err = CorrectCase(r1, sch, ctx)

		assert.Nil(t, err)
		assert.True(t, reflect.DeepEqual(r0.Complex, r1.Complex))
	}
}
