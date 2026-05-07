package filter

import (
	"context"
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestVisitor(t *testing.T) {
	s := new(VisitorTestSuite)
	suite.Run(t, s)
}

type VisitorTestSuite struct {
	suite.Suite
	resourceType *spec.ResourceType
}

func (s *VisitorTestSuite) TestUnsupportedFilterDoesNotPreventRemainingFiltersFromRunning() {
	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
		expect      func(t *testing.T, propVisited []prop.Property)
	}{
		{
			name: "traverse resource",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"schemas": []interface{}{"A", "B"},
					"id":      "foobar",
					"meta": map[string]interface{}{
						"version": "v1",
					},
					"emails": []interface{}{
						map[string]interface{}{
							"value": "foo@bar.com",
						},
						map[string]interface{}{
							"value": "bar@foo.com",
						},
					},
				}).HasError())
				return r
			},
			expect: func(t *testing.T, propVisited []prop.Property) {
				for i, p := range []string{
					"schemas",
					"schemas$elem",
					"schemas$elem",
					"id",
					"externalId",
					"meta",
					"meta.resourceType",
					"meta.created",
					"meta.lastModified",
					"meta.location",
					"meta.version",
					"urn:ietf:params:scim:schemas:core:2.0:User:userName",
					"urn:ietf:params:scim:schemas:core:2.0:User:name",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.formatted",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.familyName",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.givenName",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.middleName",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.honorificPrefix",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.honorificSuffix",
					"urn:ietf:params:scim:schemas:core:2.0:User:displayName",
					"urn:ietf:params:scim:schemas:core:2.0:User:nickName",
					"urn:ietf:params:scim:schemas:core:2.0:User:profileUrl",
					"urn:ietf:params:scim:schemas:core:2.0:User:title",
					"urn:ietf:params:scim:schemas:core:2.0:User:userType",
					"urn:ietf:params:scim:schemas:core:2.0:User:preferredLanguage",
					"urn:ietf:params:scim:schemas:core:2.0:User:locale",
					"urn:ietf:params:scim:schemas:core:2.0:User:timezone",
					"urn:ietf:params:scim:schemas:core:2.0:User:active",
					"urn:ietf:params:scim:schemas:core:2.0:User:password",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails$elem",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.value",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.type",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.primary",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.display",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails$elem",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.value",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.type",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.primary",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.display",
					"urn:ietf:params:scim:schemas:core:2.0:User:phoneNumbers",
					"urn:ietf:params:scim:schemas:core:2.0:User:ims",
					"urn:ietf:params:scim:schemas:core:2.0:User:photos",
					"urn:ietf:params:scim:schemas:core:2.0:User:addresses",
					"urn:ietf:params:scim:schemas:core:2.0:User:groups",
					"urn:ietf:params:scim:schemas:core:2.0:User:entitlements",
					"urn:ietf:params:scim:schemas:core:2.0:User:roles",
					"urn:ietf:params:scim:schemas:core:2.0:User:x509Certificates",
				} {
					assert.Equal(t, p, propVisited[i].Attribute().ID())
				}
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource(t)
			recordingFilter := recordingPropertyFilter{propHistory: []prop.Property{}}
			err := Visit(context.Background(), resource, alwaysUnsupportedFilter{}, &recordingFilter)
			assert.Nil(t, err)
			test.expect(t, recordingFilter.propHistory)
		})
	}
}

func (s *VisitorTestSuite) TestVisit() {
	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
		expect      func(t *testing.T, propVisited []prop.Property)
	}{
		{
			name: "traverse resource",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"schemas": []interface{}{"A", "B"},
					"id":      "foobar",
					"meta": map[string]interface{}{
						"version": "v1",
					},
					"emails": []interface{}{
						map[string]interface{}{
							"value": "foo@bar.com",
						},
						map[string]interface{}{
							"value": "bar@foo.com",
						},
					},
				}).HasError())
				return r
			},
			expect: func(t *testing.T, propVisited []prop.Property) {
				for i, p := range []string{
					"schemas",
					"schemas$elem",
					"schemas$elem",
					"id",
					"externalId",
					"meta",
					"meta.resourceType",
					"meta.created",
					"meta.lastModified",
					"meta.location",
					"meta.version",
					"urn:ietf:params:scim:schemas:core:2.0:User:userName",
					"urn:ietf:params:scim:schemas:core:2.0:User:name",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.formatted",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.familyName",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.givenName",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.middleName",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.honorificPrefix",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.honorificSuffix",
					"urn:ietf:params:scim:schemas:core:2.0:User:displayName",
					"urn:ietf:params:scim:schemas:core:2.0:User:nickName",
					"urn:ietf:params:scim:schemas:core:2.0:User:profileUrl",
					"urn:ietf:params:scim:schemas:core:2.0:User:title",
					"urn:ietf:params:scim:schemas:core:2.0:User:userType",
					"urn:ietf:params:scim:schemas:core:2.0:User:preferredLanguage",
					"urn:ietf:params:scim:schemas:core:2.0:User:locale",
					"urn:ietf:params:scim:schemas:core:2.0:User:timezone",
					"urn:ietf:params:scim:schemas:core:2.0:User:active",
					"urn:ietf:params:scim:schemas:core:2.0:User:password",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails$elem",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.value",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.type",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.primary",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.display",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails$elem",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.value",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.type",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.primary",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails.display",
					"urn:ietf:params:scim:schemas:core:2.0:User:phoneNumbers",
					"urn:ietf:params:scim:schemas:core:2.0:User:ims",
					"urn:ietf:params:scim:schemas:core:2.0:User:photos",
					"urn:ietf:params:scim:schemas:core:2.0:User:addresses",
					"urn:ietf:params:scim:schemas:core:2.0:User:groups",
					"urn:ietf:params:scim:schemas:core:2.0:User:entitlements",
					"urn:ietf:params:scim:schemas:core:2.0:User:roles",
					"urn:ietf:params:scim:schemas:core:2.0:User:x509Certificates",
				} {
					assert.Equal(t, p, propVisited[i].Attribute().ID())
				}
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource(t)
			filter := recordingPropertyFilter{propHistory: []prop.Property{}}
			err := Visit(context.Background(), resource, &filter)
			assert.Nil(t, err)
			test.expect(t, filter.propHistory)
		})
	}
}

func (s *VisitorTestSuite) TestVisitWithRef() {
	tests := []struct {
		name         string
		description  string
		getResource  func(t *testing.T) *prop.Resource
		getReference func(t *testing.T) *prop.Resource
		expect       func(t *testing.T, propVisited []prop.Property, refVisited []prop.Property)
	}{
		{
			name: "traverse with identical reference",
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"schemas": []interface{}{"A", "B"},
					"id":      "foobar",
					"meta": map[string]interface{}{
						"version": "v1",
					},
					"emails": []interface{}{
						map[string]interface{}{
							"value": "foo@bar.com",
						},
						map[string]interface{}{
							"value": "bar@foo.com",
						},
					},
				}).HasError())
				return r
			},
			getReference: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"schemas": []interface{}{"A", "B"},
					"id":      "foobar",
					"meta": map[string]interface{}{
						"version": "v1",
					},
					"emails": []interface{}{
						map[string]interface{}{
							"value": "foo@bar.com",
						},
						map[string]interface{}{
							"value": "bar@foo.com",
						},
					},
				}).HasError())
				return r
			},
			expect: func(t *testing.T, propVisited []prop.Property, refVisited []prop.Property) {
				assert.Equal(t, len(propVisited), len(refVisited))
				for i, p := range propVisited {
					assert.Equal(t, p.Attribute().ID(), refVisited[i].Attribute().ID())
					assert.True(t, p != refVisited[i])
				}
			},
		},
		{
			name: "traverse with different reference",
			description: `
- reference schema is missing A
- reference does not have meta
- reference does not have foo@bar.com as email, instead, it has baz@bar.com
`,
			getResource: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"schemas": []interface{}{"A", "B"},
					"id":      "foobar",
					"meta": map[string]interface{}{
						"version": "v1",
					},
					"emails": []interface{}{
						map[string]interface{}{
							"value": "foo@bar.com",
						},
						map[string]interface{}{
							"value": "bar@foo.com",
						},
					},
				}).HasError())
				return r
			},
			getReference: func(t *testing.T) *prop.Resource {
				r := prop.NewResource(s.resourceType)
				assert.False(t, r.Navigator().Replace(map[string]interface{}{
					"schemas": []interface{}{"B"},
					"id":      "foobar",
					"emails": []interface{}{
						map[string]interface{}{
							"value": "bar@foo.com",
						},
						map[string]interface{}{
							"value": "baz@bar.com",
						},
					},
				}).HasError())
				return r
			},
			expect: func(t *testing.T, propVisited []prop.Property, refVisited []prop.Property) {
				assert.Equal(t, len(propVisited), len(refVisited))

				type compare struct {
					prop string
					ref  string
				}

				visited := make([]compare, 0)
				for i, p := range propVisited {
					c := compare{prop: p.Attribute().ID()}
					if IsOutOfSync(refVisited[i]) {
						c.ref = "outOfSync"
					} else {
						c.ref = refVisited[i].Attribute().ID()
					}
					visited = append(visited, c)
				}

				for i, c := range []compare{
					{prop: "schemas", ref: "schemas"},
					{prop: "schemas$elem", ref: "outOfSync"},
					{prop: "schemas$elem", ref: "schemas$elem"},
					{prop: "id", ref: "id"},
					{prop: "externalId", ref: "externalId"},
					{prop: "meta", ref: "meta"},
					{prop: "meta.resourceType", ref: "meta.resourceType"},
					{prop: "meta.created", ref: "meta.created"},
					{prop: "meta.lastModified", ref: "meta.lastModified"},
					{prop: "meta.location", ref: "meta.location"},
					{prop: "meta.version", ref: "meta.version"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:userName", ref: "urn:ietf:params:scim:schemas:core:2.0:User:userName"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:name", ref: "urn:ietf:params:scim:schemas:core:2.0:User:name"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:name.formatted", ref: "urn:ietf:params:scim:schemas:core:2.0:User:name.formatted"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:name.familyName", ref: "urn:ietf:params:scim:schemas:core:2.0:User:name.familyName"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:name.givenName", ref: "urn:ietf:params:scim:schemas:core:2.0:User:name.givenName"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:name.middleName", ref: "urn:ietf:params:scim:schemas:core:2.0:User:name.middleName"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:name.honorificPrefix", ref: "urn:ietf:params:scim:schemas:core:2.0:User:name.honorificPrefix"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:name.honorificSuffix", ref: "urn:ietf:params:scim:schemas:core:2.0:User:name.honorificSuffix"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:displayName", ref: "urn:ietf:params:scim:schemas:core:2.0:User:displayName"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:nickName", ref: "urn:ietf:params:scim:schemas:core:2.0:User:nickName"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:profileUrl", ref: "urn:ietf:params:scim:schemas:core:2.0:User:profileUrl"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:title", ref: "urn:ietf:params:scim:schemas:core:2.0:User:title"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:userType", ref: "urn:ietf:params:scim:schemas:core:2.0:User:userType"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:preferredLanguage", ref: "urn:ietf:params:scim:schemas:core:2.0:User:preferredLanguage"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:locale", ref: "urn:ietf:params:scim:schemas:core:2.0:User:locale"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:timezone", ref: "urn:ietf:params:scim:schemas:core:2.0:User:timezone"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:active", ref: "urn:ietf:params:scim:schemas:core:2.0:User:active"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:password", ref: "urn:ietf:params:scim:schemas:core:2.0:User:password"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:emails", ref: "urn:ietf:params:scim:schemas:core:2.0:User:emails"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:emails$elem", ref: "outOfSync"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:emails.value", ref: "outOfSync"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:emails.type", ref: "outOfSync"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:emails.primary", ref: "outOfSync"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:emails.display", ref: "outOfSync"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:emails$elem", ref: "urn:ietf:params:scim:schemas:core:2.0:User:emails$elem"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:emails.value", ref: "urn:ietf:params:scim:schemas:core:2.0:User:emails.value"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:emails.type", ref: "urn:ietf:params:scim:schemas:core:2.0:User:emails.type"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:emails.primary", ref: "urn:ietf:params:scim:schemas:core:2.0:User:emails.primary"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:emails.display", ref: "urn:ietf:params:scim:schemas:core:2.0:User:emails.display"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:phoneNumbers", ref: "urn:ietf:params:scim:schemas:core:2.0:User:phoneNumbers"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:ims", ref: "urn:ietf:params:scim:schemas:core:2.0:User:ims"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:photos", ref: "urn:ietf:params:scim:schemas:core:2.0:User:photos"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:addresses", ref: "urn:ietf:params:scim:schemas:core:2.0:User:addresses"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:groups", ref: "urn:ietf:params:scim:schemas:core:2.0:User:groups"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:entitlements", ref: "urn:ietf:params:scim:schemas:core:2.0:User:entitlements"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:roles", ref: "urn:ietf:params:scim:schemas:core:2.0:User:roles"},
					{prop: "urn:ietf:params:scim:schemas:core:2.0:User:x509Certificates", ref: "urn:ietf:params:scim:schemas:core:2.0:User:x509Certificates"},
				} {
					assert.Equal(t, c.prop, visited[i].prop)
					assert.Equal(t, c.ref, visited[i].ref)
				}
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource(t)
			ref := test.getReference(t)
			filter := recordingPropertyFilter{propHistory: []prop.Property{}, refHistory: []prop.Property{}}
			err := VisitWithRef(context.Background(), resource, ref, &filter)
			assert.Nil(t, err)
			test.expect(t, filter.propHistory, filter.refHistory)
		})
	}
}

func (s *VisitorTestSuite) SetupSuite() {
	for _, each := range []struct {
		filepath  string
		structure interface{}
		post      func(parsed interface{})
	}{
		{
			filepath:  "../../../../public/schemas/core_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../../../public/schemas/user_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../../../public/resource_types/user_resource_type.json",
			structure: new(spec.ResourceType),
			post: func(parsed interface{}) {
				s.resourceType = parsed.(*spec.ResourceType)
			},
		},
	} {
		f, err := os.Open(each.filepath)
		require.Nil(s.T(), err)

		raw, err := ioutil.ReadAll(f)
		require.Nil(s.T(), err)

		err = json.Unmarshal(raw, each.structure)
		require.Nil(s.T(), err)

		if each.post != nil {
			each.post(each.structure)
		}
	}
}

type recordingPropertyFilter struct {
	propHistory []prop.Property
	refHistory  []prop.Property
}

func (f *recordingPropertyFilter) Supports(_ *spec.Attribute) bool {
	return true
}

func (f *recordingPropertyFilter) Filter(_ context.Context, _ *spec.ResourceType, nav prop.Navigator) error {
	f.propHistory = append(f.propHistory, nav.Current())
	return nil
}

func (f *recordingPropertyFilter) FilterRef(_ context.Context, _ *spec.ResourceType, nav prop.Navigator, refNav prop.Navigator) error {
	f.propHistory = append(f.propHistory, nav.Current())
	f.refHistory = append(f.refHistory, refNav.Current())
	return nil
}

type alwaysUnsupportedFilter struct{}

func (f alwaysUnsupportedFilter) Supports(_ *spec.Attribute) bool {
	return false
}

func (f alwaysUnsupportedFilter) Filter(_ context.Context, _ *spec.ResourceType, _ prop.Navigator) error {
	return nil
}

func (f alwaysUnsupportedFilter) FilterRef(_ context.Context, _ *spec.ResourceType, _ prop.Navigator, _ prop.Navigator) error {
	return nil
}
