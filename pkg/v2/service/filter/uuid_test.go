package filter

import (
	"context"
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUUIDFilter(t *testing.T) {
	attr := new(spec.Attribute)
	require.Nil(t, json.Unmarshal([]byte(`
{
  "id": "id",
  "name": "id",
  "type": "string",
  "_annotations": {
    "@UUID": {}
  }
}
`), attr))

	tests := []struct {
		name        string
		getProperty func() prop.Property
		expect      func(t *testing.T, p prop.Property, err error)
	}{
		{
			name: "unassigned property gets uuid",
			getProperty: func() prop.Property {
				return prop.NewProperty(attr)
			},
			expect: func(t *testing.T, p prop.Property, err error) {
				assert.Nil(t, err)
				assert.False(t, p.IsUnassigned())
			},
		},
		{
			name: "assigned property does not get uuid",
			getProperty: func() prop.Property {
				p := prop.NewProperty(attr)
				_, err := p.Replace("foobar")
				assert.Nil(t, err)
				return p
			},
			expect: func(t *testing.T, p prop.Property, err error) {
				assert.Nil(t, err)
				assert.False(t, p.IsUnassigned())
				assert.Equal(t, "foobar", p.Raw())
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := UUIDFilter()
			property := test.getProperty()
			assert.True(t, filter.Supports(property.Attribute()))

			err := filter.Filter(context.Background(), nil, prop.Navigate(property))
			test.expect(t, property, err)
		})
	}
}
