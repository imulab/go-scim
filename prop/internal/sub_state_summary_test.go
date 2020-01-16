package internal

import (
	"encoding/json"
	"fmt"
	"github.com/elvsn/scim.go/prop"
	"github.com/elvsn/scim.go/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestComplexStateSummarySubscriber(t *testing.T) {
	attrFunc := func(t *testing.T) *spec.Attribute {
		attr := new(spec.Attribute)
		err := json.Unmarshal([]byte(fmt.Sprintf(`
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User:name",
  "name": "name",
  "type": "complex",
  "_path": "name",
  "subAttributes": [
    {
      "id": "urn:ietf:params:scim:schemas:core:2.0:User:name.givenName",
      "name": "givenName",
      "type": "string",
      "_path": "name.givenName",
      "_index": 0
    },
    {
      "id": "urn:ietf:params:scim:schemas:core:2.0:User:name.familyName",
      "name": "familyName",
      "type": "string",
      "_path": "name.familyName",
      "_index": 1
    }
  ],
  "_annotations": {
    "@StateSummary": {},
    "@%s": {}
  }
}
`, t.Name())), attr)
		require.Nil(t, err)
		return attr
	}

	tests := []struct {
		name        string
		getProperty func(t *testing.T) prop.Property
		modFunc     func(t *testing.T, p prop.Property)
		expect      func(t *testing.T, events *prop.Events)
	}{
		{
			name: "assign value to unassigned complex property yields assign event",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewComplex(attrFunc(t))
			},
			modFunc: func(t *testing.T, p prop.Property) {
				nav := prop.Navigate(p).Dot("givenName")
				require.Nil(t, nav.Error())
				assert.Nil(t, nav.Replace("David"))
			},
			expect: func(t *testing.T, events *prop.Events) {
				assert.Equal(t, 2, events.Count())
				assert.NotNil(t, events.FindEvent(func(ev *prop.Event) bool {
					return ev.Source().Attribute().ID() == "urn:ietf:params:scim:schemas:core:2.0:User:name" &&
						ev.Type() == prop.EventAssigned
				}))
			},
		},
		{
			name: "assign value to assigned complex property does not yield new event",
			getProperty: func(t *testing.T) prop.Property {
				p := prop.NewComplex(attrFunc(t))
				err := prop.Navigate(p).Add(map[string]interface{}{
					"givenName": "David",
				})
				assert.Nil(t, err)
				return p
			},
			modFunc: func(t *testing.T, p prop.Property) {
				// must use navigate to modify in order to trigger event propagation
				nav := prop.Navigate(p).Dot("familyName")
				require.Nil(t, nav.Error())
				assert.Nil(t, nav.Replace("Q"))
			},
			expect: func(t *testing.T, events *prop.Events) {
				assert.Equal(t, 1, events.Count())
				assert.Nil(t, events.FindEvent(func(ev *prop.Event) bool {
					return ev.Type() == prop.EventAssigned &&
						ev.Source().Attribute().ID() == "urn:ietf:params:scim:schemas:core:2.0:User:name"
				}))
			},
		},
		{
			name: "delete assigned complex property yield new event",
			getProperty: func(t *testing.T) prop.Property {
				p := prop.NewComplex(attrFunc(t))
				err := prop.Navigate(p).Add(map[string]interface{}{
					"givenName": "David",
				})
				assert.Nil(t, err)
				return p
			},
			modFunc: func(t *testing.T, p prop.Property) {
				// must use navigate to modify in order to trigger event propagation
				assert.Nil(t, prop.Navigate(p).Delete())
			},
			expect: func(t *testing.T, events *prop.Events) {
				assert.Equal(t, 1, events.Count())
				assert.NotNil(t, events.FindEvent(func(ev *prop.Event) bool {
					return ev.Type() == prop.EventUnassigned &&
						ev.Source().Attribute().ID() == "urn:ietf:params:scim:schemas:core:2.0:User:name"
				}))
			},
		},
		{
			name: "delete unassigned complex property does not yield new event",
			getProperty: func(t *testing.T) prop.Property {
				return prop.NewComplex(attrFunc(t))
			},
			modFunc: func(t *testing.T, p prop.Property) {
				// must use navigate to modify in order to trigger event propagation
				assert.Nil(t, prop.Navigate(p).Delete())
			},
			expect: func(t *testing.T, events *prop.Events) {
				assert.Nil(t, events)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rs := recordingSubscriber{}
			prop.SubscriberFactory().Register(fmt.Sprintf("@%s", t.Name()), func(_ prop.Property, _ map[string]interface{}) prop.Subscriber {
				return &rs
			})
			p := test.getProperty(t)
			test.modFunc(t, p)
			test.expect(t, rs.events)
		})
	}
}
