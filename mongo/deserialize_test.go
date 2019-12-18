package mongo

import (
	"encoding/json"
	scimJSON "github.com/imulab/go-scim/core/json"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/core/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestDeserialize(t *testing.T) {
	s := new(MongoDeserializerTestSuite)
	s.resourceBase = "./internal/mongo_deserializer_test_suite"
	suite.Run(t, s)
}

type MongoDeserializerTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *MongoDeserializerTestSuite) TestDeserialize() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name        string
		getResource func(t *testing.T) *prop.Resource
		expect      func(t *testing.T, r *prop.Resource, err error)
	}{
		{
			name: "user resource",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			expect: func(t *testing.T, r *prop.Resource, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, r)

				type focus func(nav *prop.FluentNavigator) *prop.FluentNavigator
				for _, each := range []struct {
					expect interface{}
					focus  focus
				}{
					{
						expect: "urn:ietf:params:scim:schemas:core:2.0:User",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("schemas").FocusIndex(0)
						},
					},
					{
						expect: nil,
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("id")
						},
					},
					{
						expect: "imulab",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("userName")
						},
					},
					{
						expect: "Mr. Weinan Qiu",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("name").FocusName("formatted")
						},
					},
					{
						expect: "Weinan",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("name").FocusName("givenName")
						},
					},
					{
						expect: "Qiu",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("name").FocusName("familyName")
						},
					},
					{
						expect: "Mr.",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("name").FocusName("honorificPrefix")
						},
					},
					{
						expect: "Weinan",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("displayName")
						},
					},
					{
						expect: "https://identity.imulab.io/profiles/3cc032f5-2361-417f-9e2f-bc80adddf4a3",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("profileUrl")
						},
					},
					{
						expect: "Employee",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("userType")
						},
					},
					{
						expect: "zh_CN",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("preferredLanguage")
						},
					},
					{
						expect: "zh_CN",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("locale")
						},
					},
					{
						expect: "Asia/Shanghai",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("timezone")
						},
					},
					{
						expect: true,
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("active")
						},
					},
					{
						expect: "imulab@foo.com",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("emails").FocusIndex(0).FocusName("value")
						},
					},
					{
						expect: "work",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("emails").FocusIndex(0).FocusName("type")
						},
					},
					{
						expect: true,
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("emails").FocusIndex(0).FocusName("primary")
						},
					},
					{
						expect: "imulab@foo.com",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("emails").FocusIndex(0).FocusName("display")
						},
					},
					{
						expect: "imulab@bar.com",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("emails").FocusIndex(1).FocusName("value")
						},
					},
					{
						expect: "home",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("emails").FocusIndex(1).FocusName("type")
						},
					},
					{
						expect: nil,
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("emails").FocusIndex(1).FocusName("primary")
						},
					},
					{
						expect: "imulab@bar.com",
						focus: func(nav *prop.FluentNavigator) *prop.FluentNavigator {
							return nav.FocusName("emails").FocusIndex(1).FocusName("display")
						},
					},
				} {
					actual := each.focus(r.NewFluentNavigator()).Current().Raw()
					if each.expect == nil {
						assert.Nil(t, actual)
					} else {
						assert.Equal(t, each.expect, actual)
					}
				}
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource(t)
			raw, err := newBsonAdapter(resource).MarshalBSON()
			assert.Nil(t, err)
			um := newResourceUnmarshaler(resource.ResourceType())
			err = um.UnmarshalBSON(raw)
			test.expect(t, um.Resource(), err)
		})
	}
}

func (s *MongoDeserializerTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *MongoDeserializerTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *MongoDeserializerTestSuite) mustSchema(filePath string) *spec.Schema {
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
