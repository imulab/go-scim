package prop

import (
	"encoding/json"
	"github.com/imulab/go-scim/src/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

type (
	ResourceTestSuite struct {
		suite.Suite
		resourceBase string
	}
	testVisitor struct {
		trails []string
	}
)

func TestResource(t *testing.T) {
	s := new(ResourceTestSuite)
	s.resourceBase = "../../tests/resource_test_suite"
	suite.Run(t, s)
}

func (s *ResourceTestSuite) TestNewResourceOf() {
	_ = s.mustSchema("/user_schema.json")
	resource := NewResourceOf(s.mustResourceType("/user_resource_type.json"), map[string]interface{}{
		"schemas": []interface{}{
			"urn:ietf:params:scim:schemas:core:2.0:User",
		},
		"id": "3cc032f5-2361-417f-9e2f-bc80adddf4a3",
		"meta": map[string]interface{}{
			"resourceType": "User",
			"created":      "2019-11-20T13:09:00",
			"lastModified": "2019-11-20T13:09:00",
			"location":     "https://identity.imulab.io/Users/3cc032f5-2361-417f-9e2f-bc80adddf4a3",
			"version":      "W/\"1\"",
		},
		"userName": "imulab",
		"name": map[string]interface{}{
			"formatted":       "Mr. Weinan Qiu",
			"familyName":      "Qiu",
			"givenName":       "Weinan",
			"honorificPrefix": "Mr.",
		},
		"displayName":       "Weinan",
		"profileUrl":        "https://identity.imulab.io/profiles/3cc032f5-2361-417f-9e2f-bc80adddf4a3",
		"userType":          "Employee",
		"preferredLanguage": "zh_CN",
		"locale":            "zh_CN",
		"timezone":          "Asia/Shanghai",
		"active":            true,
		"emails": []interface{}{
			map[string]interface{}{
				"value":   "imulab@foo.com",
				"type":    "work",
				"primary": true,
				"display": "imulab@foo.com",
			},
			map[string]interface{}{
				"value":   "imulab@bar.com",
				"type":    "home",
				"display": "imulab@bar.com",
			},
		},
		"phoneNumbers": []interface{}{
			map[string]interface{}{
				"value":   "123-45678",
				"type":    "work",
				"primary": true,
				"display": "123-45678",
			},
			map[string]interface{}{
				"value":   "123-45679",
				"type":    "work",
				"display": "123-45679",
			},
		},
	})
	assert.NotNil(s.T(), resource)
	nav := resource.NewNavigator()

	_, err := nav.FocusName("schemas")
	assert.Nil(s.T(), err)
	_, err = nav.FocusIndex(0)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), "urn:ietf:params:scim:schemas:core:2.0:User", nav.Current().Raw())
	nav.Retract()
	nav.Retract()

	_, err = nav.FocusName("id")
	assert.Equal(s.T(), "3cc032f5-2361-417f-9e2f-bc80adddf4a3", nav.Current().Raw())
	nav.Retract()
}

func (s *ResourceTestSuite) TestVisit() {
	_ = s.mustSchema("/user_schema.json")
	resource := NewResource(s.mustResourceType("/user_resource_type.json"))
	visitor := new(testVisitor)
	err := resource.Visit(visitor)
	assert.Nil(s.T(), err)
	visitor.assertTrail(s.T(), []string{
		"<",
		"schemas",
		"<",
		">",
		"id",
		"externalId",
		"meta",
		"<",
		"meta.resourceType",
		"meta.created",
		"meta.lastModified",
		"meta.location",
		"meta.version",
		">",
		"urn:ietf:params:scim:schemas:core:2.0:User:userName",
		"urn:ietf:params:scim:schemas:core:2.0:User:name",
		"<",
		"urn:ietf:params:scim:schemas:core:2.0:User:name.formatted",
		"urn:ietf:params:scim:schemas:core:2.0:User:name.familyName",
		"urn:ietf:params:scim:schemas:core:2.0:User:name.givenName",
		"urn:ietf:params:scim:schemas:core:2.0:User:name.middleName",
		"urn:ietf:params:scim:schemas:core:2.0:User:name.honorificPrefix",
		"urn:ietf:params:scim:schemas:core:2.0:User:name.honorificSuffix",
		">",
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
		"<",
		">",
		"urn:ietf:params:scim:schemas:core:2.0:User:phoneNumbers",
		"<",
		">",
		"urn:ietf:params:scim:schemas:core:2.0:User:ims",
		"<",
		">",
		"urn:ietf:params:scim:schemas:core:2.0:User:photos",
		"<",
		">",
		"urn:ietf:params:scim:schemas:core:2.0:User:addresses",
		"<",
		">",
		"urn:ietf:params:scim:schemas:core:2.0:User:groups",
		"<",
		">",
		"urn:ietf:params:scim:schemas:core:2.0:User:entitlements",
		"<",
		">",
		"urn:ietf:params:scim:schemas:core:2.0:User:roles",
		"<",
		">",
		"urn:ietf:params:scim:schemas:core:2.0:User:x509Certificates",
		"<",
		">",
		">",
	})
}

func (s *ResourceTestSuite) mustResourceType(filePath string) *core.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(core.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *ResourceTestSuite) mustSchema(filePath string) *core.Schema {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	sch := new(core.Schema)
	err = json.Unmarshal(raw, sch)
	s.Require().Nil(err)

	core.SchemaHub.Put(sch)

	return sch
}

func (v *testVisitor) ShouldVisit(property core.Property) bool {
	return true
}

func (v *testVisitor) Visit(property core.Property) error {
	v.trails = append(v.trails, property.Attribute().ID())
	return nil
}

func (v *testVisitor) BeginChildren(container core.Container) {
	v.trails = append(v.trails, "<")
}

func (v *testVisitor) EndChildren(container core.Container) {
	v.trails = append(v.trails, ">")
}

func (v *testVisitor) assertTrail(t *testing.T, expect []string) {
	assert.Equal(t, len(expect), len(v.trails))
	for i := range v.trails {
		assert.Equal(t, expect[i], v.trails[i])
	}
}
