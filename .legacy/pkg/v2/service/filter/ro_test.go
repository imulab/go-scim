package filter

import (
	"context"
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadOnlyFilter(t *testing.T) {
	tests := []struct {
		name         string
		attrJson     string
		getProperty  func(attr *spec.Attribute) prop.Property
		getReference func(attr *spec.Attribute) prop.Property
		expect       func(t *testing.T, p prop.Property, err error)
	}{
		{
			name: "read only property is cleared",
			attrJson: `
{
  "id": "id",
  "name": "id",
  "type": "string",
  "mutability": "readOnly",
  "_annotations": {
    "@ReadOnly": {
      "reset": true,
      "copy": true
    }
  }
}
`,
			getProperty: func(attr *spec.Attribute) prop.Property {
				p := prop.NewProperty(attr)
				_, err := p.Replace("foobar")
				assert.Nil(t, err)
				return p
			},
			getReference: func(attr *spec.Attribute) prop.Property {
				return nil
			},
			expect: func(t *testing.T, p prop.Property, err error) {
				assert.Nil(t, err)
				assert.True(t, p.IsUnassigned())
			},
		},
		{
			name: "read only property is cleared and copied",
			attrJson: `
{
  "id": "id",
  "name": "id",
  "type": "string",
  "mutability": "readOnly",
  "_annotations": {
    "@ReadOnly": {
      "reset": true,
      "copy": true
    }
  }
}
`,
			getProperty: func(attr *spec.Attribute) prop.Property {
				p := prop.NewProperty(attr)
				_, err := p.Replace("foobar")
				assert.Nil(t, err)
				return p
			},
			getReference: func(attr *spec.Attribute) prop.Property {
				p := prop.NewProperty(attr)
				_, err := p.Replace("dbf6c563-78da-45b8-958e-f2a85562419c")
				assert.Nil(t, err)
				return p
			},
			expect: func(t *testing.T, p prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "dbf6c563-78da-45b8-958e-f2a85562419c", p.Raw())
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attr := new(spec.Attribute)
			assert.Nil(t, json.Unmarshal([]byte(test.attrJson), attr))

			property := test.getProperty(attr)
			reference := test.getReference(attr)

			var err error
			filter := ReadOnlyFilter()
			if reference == nil {
				err = filter.Filter(context.Background(), nil, prop.Navigate(property))
			} else {
				err = filter.FilterRef(context.Background(), nil, prop.Navigate(property), prop.Navigate(reference))
			}

			test.expect(t, property, err)
		})
	}
}
