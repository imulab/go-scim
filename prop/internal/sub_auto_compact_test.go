package internal

import (
	"encoding/json"
	"github.com/elvsn/scim.go/prop"
	"github.com/elvsn/scim.go/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAutoCompactSubscriber(t *testing.T) {
	attrFunc := func(t *testing.T) *spec.Attribute {
		attr := new(spec.Attribute)
		require.Nil(t, json.Unmarshal([]byte(`
{
  "id": "schemas",
  "name": "schemas",
  "type": "string",
  "multiValued": true,
  "_path": "schemas",
  "_index": 0,
  "_annotations": {
    "@AutoCompact": {}
  }
}
`), attr))
		return attr
	}

	tests := []struct {
		name        string
		getProperty func(t *testing.T) prop.Property
		modFunc     func(t *testing.T, p prop.Property)
		expect      func(t *testing.T, raw interface{})
	}{
		{
			name: "removing element auto compacts multiValued property",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewMultiOf(attrFunc(t), []interface{}{"A", "B", "C"})
			},
			modFunc: func(t *testing.T, p prop.Property) {
				assert.Nil(t, prop.Navigate(p).At(1).Delete())
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, []interface{}{"A", "C"}, raw)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := test.getProperty(t)
			test.modFunc(t, p)
			test.expect(t, p.Raw())
		})
	}
}
