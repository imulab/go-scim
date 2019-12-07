package crud

import (
	"encoding/json"
	"github.com/imulab/go-scim/pkg/core/expr"
	scimJSON "github.com/imulab/go-scim/pkg/core/json"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestCRUD(t *testing.T) {
	s := new(CRUDTestSuite)
	s.resourceBase = "../../../tests/crud_test_suite"
	suite.Run(t, s)
}

type CRUDTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *CRUDTestSuite) TestAdd() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")
	expr.Register(resourceType)

	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
		path        string
		value       interface{}
		expect      func(t *testing.T, r *prop.Resource, err error)
	}{
		{
			name: "add to non existing field yields error",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "foobar",
			value: "foobar",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.NotNil(t, err)
			},
		},
		{
			name: "add to top level simple property",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "userName",
			value: "foobar",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				p, err := r.NewNavigator().FocusName("userName")
				assert.Nil(t, err)
				assert.Equal(t, "foobar", p.Raw())
			},
		},
		{
			name: "add to nested simple property",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "name.givenName",
			value: "Foobar",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)

				nav := r.NewNavigator()
				{
					_, err = nav.FocusName("name")
					assert.Nil(t, err)
					_, err = nav.FocusName("givenName")
					assert.Nil(t, err)
				}
				assert.Equal(t, "Foobar", nav.Current().Raw())
			},
		},
		{
			name: "add to simple multiValued property",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "schemas",
			value: "Foobar",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)

				nav := r.NewNavigator()
				{
					_, err = nav.FocusName("schemas")
					assert.Nil(t, err)
					_, err = nav.FocusIndex(1)
					assert.Nil(t, err)
				}
				assert.Equal(t, "Foobar", nav.Current().Raw())
			},
		},
		{
			name: "add to complex multiValued property",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "emails",
			value: map[string]interface{}{
				"value": "test@test.org",
			},
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)

				nav := r.NewNavigator()
				{
					_, err = nav.FocusName("emails")
					assert.Nil(t, err)
					_, err = nav.FocusIndex(2)
					assert.Nil(t, err)
					_, err = nav.FocusName("value")
					assert.Nil(t, err)
				}
				assert.Equal(t, "test@test.org", nav.Current().Raw())
			},
		},
		{
			name: "add to every simple property inside a complex multiValued property",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "emails.display",
			value: "foobar",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)

				nav := r.NewNavigator()
				{
					_, err = nav.FocusName("emails")
					assert.Nil(t, err)
				}
				_ = nav.Current().(prop.Container).ForEachChild(func(index int, child prop.Property) error {
					assert.Equal(t, "foobar", child.(prop.Container).ChildAtIndex("display").Raw())
					return nil
				})
			},
		},
		{
			name: "add to a selected few simple property inside a complex multiValued property",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "emails[value eq \"imulab@bar.com\"].primary",
			value: true,
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)

				nav := r.NewNavigator()
				{
					_, err = nav.FocusName("emails")
					assert.Nil(t, err)
				}
				{
					_, err = nav.FocusIndex(0)
					assert.Nil(t, err)
					assert.Equal(t, "imulab@foo.com", nav.Current().(prop.Container).ChildAtIndex("value").Raw())
					assert.Nil(t, nav.Current().(prop.Container).ChildAtIndex("primary").Raw())
					nav.Retract()
				}
				{
					_, err = nav.FocusIndex(1)
					assert.Nil(t, err)
					assert.Equal(t, "imulab@bar.com", nav.Current().(prop.Container).ChildAtIndex("value").Raw())
					assert.Equal(t, true, nav.Current().(prop.Container).ChildAtIndex("primary").Raw())
					nav.Retract()
				}
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource(t)
			err := Add(resource, test.path, test.value)
			test.expect(t, resource, err)
		})
	}
}

func (s *CRUDTestSuite) TestReplace() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")
	expr.Register(resourceType)

	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
		path        string
		expect      func(t *testing.T, r *prop.Resource, err error)
	}{
		{
			name: "delete non existing field yields error",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "foobar",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.NotNil(t, err)
			},
		},
		{
			name: "delete top level simple property",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "userName",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				p, err := r.NewNavigator().FocusName("userName")
				assert.Nil(t, err)
				assert.True(t, p.IsUnassigned())
			},
		},
		{
			name: "delete nested simple property",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "name.givenName",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)

				nav := r.NewNavigator()
				{
					_, err = nav.FocusName("name")
					assert.Nil(t, err)
					_, err = nav.FocusName("givenName")
					assert.Nil(t, err)
				}
				assert.True(t, nav.Current().IsUnassigned())
			},
		},
		{
			name: "delete an element of the multiValued property",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "emails[value eq \"imulab@bar.com\"]",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)

				nav := r.NewNavigator()
				{
					_, err = nav.FocusName("emails")
					assert.Nil(t, err)
				}
				assert.Equal(t, 1, nav.Current().(prop.Container).CountChildren())
			},
		},
		{
			name: "delete a selected few simple property inside a complex multiValued property",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "emails[value eq \"imulab@bar.com\"].display",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)

				nav := r.NewNavigator()
				{
					_, err = nav.FocusName("emails")
					assert.Nil(t, err)
				}
				{
					_, err = nav.FocusIndex(0)
					assert.Nil(t, err)
					assert.Equal(t, "imulab@foo.com", nav.Current().(prop.Container).ChildAtIndex("value").Raw())
					assert.NotEmpty(t, nav.Current().(prop.Container).ChildAtIndex("display").Raw())
					nav.Retract()
				}
				{
					_, err = nav.FocusIndex(1)
					assert.Nil(t, err)
					assert.Equal(t, "imulab@bar.com", nav.Current().(prop.Container).ChildAtIndex("value").Raw())
					assert.True(t, nav.Current().(prop.Container).ChildAtIndex("display").IsUnassigned())
					nav.Retract()
				}
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource(t)
			err := Delete(resource, test.path)
			test.expect(t, resource, err)
		})
	}
}

func (s *CRUDTestSuite) TestDelete() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")
	expr.Register(resourceType)

	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
		path        string
		value       interface{}
		expect      func(t *testing.T, r *prop.Resource, err error)
	}{
		{
			name: "replace non existing field yields error",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "foobar",
			value: "foobar",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.NotNil(t, err)
			},
		},
		{
			name: "replace top level simple property",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "userName",
			value: "foobar",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				p, err := r.NewNavigator().FocusName("userName")
				assert.Nil(t, err)
				assert.Equal(t, "foobar", p.Raw())
			},
		},
		{
			name: "replace nested simple property",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "name.givenName",
			value: "Foobar",
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)

				nav := r.NewNavigator()
				{
					_, err = nav.FocusName("name")
					assert.Nil(t, err)
					_, err = nav.FocusName("givenName")
					assert.Nil(t, err)
				}
				assert.Equal(t, "Foobar", nav.Current().Raw())
			},
		},
		{
			name: "replace complex multiValued property",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "emails",
			value: []interface{}{
				map[string]interface{}{
					"value": "test@test.org",
				},
			},
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)

				nav := r.NewNavigator()
				{
					_, err = nav.FocusName("emails")
					assert.Nil(t, err)
				}
				assert.Equal(t, 1, nav.Current().(prop.Container).CountChildren())
			},
		},
		{
			name: "replace a selected few simple property inside a complex multiValued property",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			path:  "emails[value eq \"imulab@bar.com\"].primary",
			value: true,
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)

				nav := r.NewNavigator()
				{
					_, err = nav.FocusName("emails")
					assert.Nil(t, err)
				}
				{
					_, err = nav.FocusIndex(0)
					assert.Nil(t, err)
					assert.Equal(t, "imulab@foo.com", nav.Current().(prop.Container).ChildAtIndex("value").Raw())
					assert.Nil(t, nav.Current().(prop.Container).ChildAtIndex("primary").Raw())
					nav.Retract()
				}
				{
					_, err = nav.FocusIndex(1)
					assert.Nil(t, err)
					assert.Equal(t, "imulab@bar.com", nav.Current().(prop.Container).ChildAtIndex("value").Raw())
					assert.Equal(t, true, nav.Current().(prop.Container).ChildAtIndex("primary").Raw())
					nav.Retract()
				}
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource(t)
			err := Replace(resource, test.path, test.value)
			test.expect(t, resource, err)
		})
	}
}

func (s *CRUDTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *CRUDTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *CRUDTestSuite) mustSchema(filePath string) *spec.Schema {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	sch := new(spec.Schema)
	err = json.Unmarshal(raw, sch)
	s.Require().Nil(err)

	spec.SchemaHub.Put(sch)

	return sch
}
