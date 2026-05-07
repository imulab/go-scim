package filter

import (
	"context"
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestBCryptFilter(t *testing.T) {
	attr := new(spec.Attribute)
	require.Nil(t, json.Unmarshal([]byte(`
{
  "id": "password",
  "name": "password",
  "type": "string",
  "_annotations": {
    "@BCrypt": {
      "cost": 5
	}
  }
}
`), attr))

	tests := []struct {
		name         string
		getProperty  func() prop.Property
		getReference func() prop.Property
		expect       func(t *testing.T, p prop.Property, err error)
	}{
		{
			name: "unassigned property does not hash",
			getProperty: func() prop.Property {
				return prop.NewProperty(attr)
			},
			getReference: func() prop.Property {
				return nil
			},
			expect: func(t *testing.T, p prop.Property, err error) {
				assert.Nil(t, err)
				assert.True(t, p.IsUnassigned())
			},
		},
		{
			name: "assigned property is hahsed",
			getProperty: func() prop.Property {
				p := prop.NewProperty(attr)
				_, err := p.Replace("s3cret")
				assert.Nil(t, err)
				return p
			},
			getReference: func() prop.Property {
				return nil
			},
			expect: func(t *testing.T, p prop.Property, err error) {
				assert.Nil(t, err)
				assert.False(t, p.IsUnassigned())
				cost, err := bcrypt.Cost([]byte(p.Raw().(string)))
				assert.Nil(t, err)
				assert.Equal(t, 5, cost)
			},
		},
		{
			name: "same value as reference does not hash",
			getProperty: func() prop.Property {
				p := prop.NewProperty(attr)
				_, err := p.Replace("pretending_to_have_been_hashed")
				assert.Nil(t, err)
				return p
			},
			getReference: func() prop.Property {
				p := prop.NewProperty(attr)
				_, err := p.Replace("pretending_to_have_been_hashed")
				assert.Nil(t, err)
				return p
			},
			expect: func(t *testing.T, p prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "pretending_to_have_been_hashed", p.Raw())
			},
		},
		{
			name: "different value as reference gets hashed",
			getProperty: func() prop.Property {
				p := prop.NewProperty(attr)
				_, err := p.Replace("new_s3cret")
				assert.Nil(t, err)
				return p
			},
			getReference: func() prop.Property {
				p := prop.NewProperty(attr)
				_, err := p.Replace("pretending_to_have_been_hashed")
				assert.Nil(t, err)
				return p
			},
			expect: func(t *testing.T, p prop.Property, err error) {
				assert.Nil(t, err)
				assert.NotEqual(t, "pretending_to_have_been_hashed", p.Raw())
				cost, err := bcrypt.Cost([]byte(p.Raw().(string)))
				assert.Nil(t, err)
				assert.Equal(t, 5, cost)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := BCryptFilter()

			property := test.getProperty()
			reference := test.getReference()
			assert.True(t, filter.Supports(property.Attribute()))

			var err error
			if reference == nil {
				err = filter.Filter(context.Background(),
					nil, prop.Navigate(property))
			} else {
				err = filter.FilterRef(context.Background(),
					nil, prop.Navigate(property), prop.Navigate(reference))
			}

			test.expect(t, property, err)
		})
	}
}
