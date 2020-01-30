package prop

import (
	"encoding/json"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/spec"
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
		getProperty func(t *testing.T) Property
		modFunc     func(t *testing.T, p Property)
		expect      func(t *testing.T, raw interface{})
	}{
		{
			name: "removing element auto compacts multiValued property",
			getProperty: func(t *testing.T) Property {
				return NewMultiOf(attrFunc(t), []interface{}{"A", "B", "C"})
			},
			modFunc: func(t *testing.T, p Property) {
				assert.False(t, Navigate(p).At(1).Delete().HasError())
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

func TestExclusivePrimarySubscriber(t *testing.T) {
	attrFunc := func(t *testing.T) *spec.Attribute {
		attr := new(spec.Attribute)
		require.Nil(t, json.Unmarshal([]byte(`
{
  "id": "emails",
  "name": "emails",
  "type": "complex",
  "multiValued": true,
  "subAttributes": [
    {
      "id": "emails.value",
      "name": "value",
      "type": "string",
      "_path": "emails.value",
      "_index": 0,
      "_annotations": {
        "@Identity": {}
      }
    },
    {
      "id": "emails.primary",
      "name": "primary",
      "type": "boolean",
      "_path": "emails.primary",
      "_index": 1,
      "_annotations": {
        "@Primary": {}
      }
    }
  ],
  "_path": "emails",
  "_index": 0,
  "_annotations": {
    "@ExclusivePrimary": {}
  }
}
`), attr))
		return attr
	}

	tests := []struct {
		name        string
		getProperty func(t *testing.T) Property
		modFunc     func(t *testing.T, p Property)
		expect      func(t *testing.T, raw interface{})
	}{
		{
			name: "assigning new primary turns off old primary",
			getProperty: func(t *testing.T) Property {
				return NewMultiOf(attrFunc(t), []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value": "bar",
					},
				})
			},
			modFunc: func(t *testing.T, p Property) {
				assert.False(t, Navigate(p).At(1).Dot("primary").Replace(true).HasError())
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": nil,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": true,
					},
				}, raw)
			},
		},
		{
			name: "assigning old primary has no side effect",
			getProperty: func(t *testing.T) Property {
				return NewMultiOf(attrFunc(t), []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value": "bar",
					},
				})
			},
			modFunc: func(t *testing.T, p Property) {
				assert.False(t, Navigate(p).At(0).Dot("primary").Replace(true).HasError())
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": nil,
					},
				}, raw)
			},
		},
		{
			name: "turning off old primary has no side effect",
			getProperty: func(t *testing.T) Property {
				return NewMultiOf(attrFunc(t), []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value": "bar",
					},
				})
			},
			modFunc: func(t *testing.T, p Property) {
				assert.False(t, Navigate(p).At(0).Dot("primary").Delete().HasError())
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": nil,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": nil,
					},
				}, raw)
			},
		},
		{
			name: "assigning other primary to false has no side effect",
			getProperty: func(t *testing.T) Property {
				return NewMultiOf(attrFunc(t), []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value": "bar",
					},
				})
			},
			modFunc: func(t *testing.T, p Property) {
				assert.False(t, Navigate(p).At(1).Dot("primary").Replace(false).HasError())
			},
			expect: func(t *testing.T, raw interface{}) {
				assert.Equal(t, []interface{}{
					map[string]interface{}{
						"value":   "foo",
						"primary": true,
					},
					map[string]interface{}{
						"value":   "bar",
						"primary": false,
					},
				}, raw)
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

func TestSchemaSyncSubscriber(t *testing.T) {
	var (
		coreSchema      = new(spec.Schema)
		mainSchema      = new(spec.Schema)
		extensionSchema = new(spec.Schema)
		resourceType    = new(spec.ResourceType)
	)
	{
		require.Nil(t, json.Unmarshal([]byte(`
{
  "id": "core",
  "name": "core",
  "attributes": [
    {
      "id": "schemas",
      "name": "schemas",
      "type": "string",
      "multiValued": true,
      "_path": "schemas",
      "_annotations": {
        "@AutoCompact": {}
      }
    }
  ]
}
`), coreSchema))
		spec.Schemas().Register(coreSchema)

		require.Nil(t, json.Unmarshal([]byte(`
{
  "id": "main",
  "name": "main",
  "attributes": [
    {
      "id": "name",
      "name": "name",
      "type": "string",
      "_path": "name"
    }
  ]
}
`), mainSchema))
		spec.Schemas().Register(mainSchema)

		require.Nil(t, json.Unmarshal([]byte(`
{
  "id": "extension",
  "name": "extension",
  "attributes": [
    {
      "id": "text",
      "name": "text",
      "type": "string",
      "_path": "extension.text"
    }
  ]
}
`), extensionSchema))
		spec.Schemas().Register(extensionSchema)

		require.Nil(t, json.Unmarshal([]byte(`
{
  "id": "Test",
  "name": "Test",
  "schema": "main",
  "schemaExtensions": [
    {
      "schema": "extension",
      "required": false
    }
  ]
}
`), resourceType))
	}

	tests := []struct {
		name        string
		getProperty func(t *testing.T) Property
		modFunc     func(t *testing.T, p Property)
		expect      func(t *testing.T, schemas interface{})
	}{
		{
			name: "assign to unassigned extension",
			getProperty: func(t *testing.T) Property {
				return NewComplexOf(resourceType.SuperAttribute(true), map[string]interface{}{
					"schemas": []interface{}{"main"},
					"name":    "foobar",
				})
			},
			modFunc: func(t *testing.T, p Property) {
				assert.False(t, Navigate(p).Dot("extension").Dot("text").Replace("hello world").HasError())
			},
			expect: func(t *testing.T, schemas interface{}) {
				assert.Equal(t, []interface{}{"main", "extension"}, schemas)
			},
		},
		{
			name: "assign to assigned extension does not add extension",
			getProperty: func(t *testing.T) Property {
				return NewComplexOf(resourceType.SuperAttribute(true), map[string]interface{}{
					"schemas": []interface{}{"main", "extension"},
					"name":    "foobar",
					"extension": map[string]interface{}{
						"text": "hello world",
					},
				})
			},
			modFunc: func(t *testing.T, p Property) {
				assert.False(t, Navigate(p).Dot("extension").Replace(map[string]interface{}{
					"text": "see ya later",
				}).HasError())
			},
			expect: func(t *testing.T, schemas interface{}) {
				assert.Equal(t, []interface{}{"main", "extension"}, schemas)
			},
		},
		{
			name: "unassigned extension has schema removed",
			getProperty: func(t *testing.T) Property {
				return NewComplexOf(resourceType.SuperAttribute(true), map[string]interface{}{
					"schemas": []interface{}{"main", "extension"},
					"name":    "foobar",
					"extension": map[string]interface{}{
						"text": "hello world",
					},
				})
			},
			modFunc: func(t *testing.T, p Property) {
				assert.False(t, Navigate(p).Dot("extension").Delete().HasError())
			},
			expect: func(t *testing.T, schemas interface{}) {
				assert.Equal(t, []interface{}{"main"}, schemas)
			},
		},
		{
			name: "deleting unassigned extension does not change schema",
			getProperty: func(t *testing.T) Property {
				return NewComplexOf(resourceType.SuperAttribute(true), map[string]interface{}{
					"schemas":   []interface{}{"main"},
					"name":      "foobar",
					"extension": map[string]interface{}{},
				})
			},
			modFunc: func(t *testing.T, p Property) {
				assert.False(t, Navigate(p).Dot("extension").Delete().HasError())
			},
			expect: func(t *testing.T, schemas interface{}) {
				assert.Equal(t, []interface{}{"main"}, schemas)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := test.getProperty(t)
			test.modFunc(t, p)
			test.expect(t, Navigate(p).Dot("schemas").Current().Raw())
		})
	}
}

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
		getProperty func(t *testing.T) Property
		modFunc     func(t *testing.T, p Property)
		expect      func(t *testing.T, events *Events)
	}{
		{
			name: "assign value to unassigned complex property yields assign event",
			getProperty: func(t *testing.T) Property {
				return NewComplex(attrFunc(t))
			},
			modFunc: func(t *testing.T, p Property) {
				assert.False(t, Navigate(p).Dot("givenName").Replace("David").HasError())
			},
			expect: func(t *testing.T, events *Events) {
				assert.Equal(t, 2, events.Count())
				assert.NotNil(t, events.FindEvent(func(ev *Event) bool {
					return ev.Source().Attribute().ID() == "urn:ietf:params:scim:schemas:core:2.0:User:name" &&
						ev.Type() == EventAssigned
				}))
			},
		},
		{
			name: "assign value to assigned complex property does not yield new event",
			getProperty: func(t *testing.T) Property {
				p := NewComplex(attrFunc(t))
				assert.False(t, Navigate(p).Add(map[string]interface{}{
					"givenName": "David",
				}).HasError())
				return p
			},
			modFunc: func(t *testing.T, p Property) {
				assert.False(t, Navigate(p).Dot("familyName").Replace("Q").HasError())
			},
			expect: func(t *testing.T, events *Events) {
				assert.Equal(t, 1, events.Count())
				assert.Nil(t, events.FindEvent(func(ev *Event) bool {
					return ev.Type() == EventAssigned &&
						ev.Source().Attribute().ID() == "urn:ietf:params:scim:schemas:core:2.0:User:name"
				}))
			},
		},
		{
			name: "delete assigned complex property yield new event",
			getProperty: func(t *testing.T) Property {
				p := NewComplex(attrFunc(t))
				assert.False(t, Navigate(p).Add(map[string]interface{}{
					"givenName": "David",
				}).HasError())
				return p
			},
			modFunc: func(t *testing.T, p Property) {
				// must use navigate to modify in order to trigger event propagation
				assert.False(t, Navigate(p).Delete().HasError())
			},
			expect: func(t *testing.T, events *Events) {
				assert.Equal(t, 1, events.Count())
				assert.NotNil(t, events.FindEvent(func(ev *Event) bool {
					return ev.Type() == EventUnassigned &&
						ev.Source().Attribute().ID() == "urn:ietf:params:scim:schemas:core:2.0:User:name"
				}))
			},
		},
		{
			name: "delete unassigned complex property does not yield new event",
			getProperty: func(t *testing.T) Property {
				return NewComplex(attrFunc(t))
			},
			modFunc: func(t *testing.T, p Property) {
				// must use navigate to modify in order to trigger event propagation
				assert.False(t, Navigate(p).Delete().HasError())
			},
			expect: func(t *testing.T, events *Events) {
				assert.Nil(t, events)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rs := recordingSubscriber{}
			SubscriberFactory().Register(fmt.Sprintf("@%s", t.Name()), func(_ Property, _ map[string]interface{}) Subscriber {
				return &rs
			})
			p := test.getProperty(t)
			test.modFunc(t, p)
			test.expect(t, rs.events)
		})
	}
}

// Internal implementation of Subscriber used in tests.
type recordingSubscriber struct {
	events *Events
}

func (s *recordingSubscriber) Notify(_ Property, events *Events) error {
	s.events = events
	return nil
}
