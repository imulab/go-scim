package filter

import (
	"context"
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

func TestFromForProperty(t *testing.T) {
	s := new(FromForPropertyTestSuite)
	s.resourceBase = "./internal/from_for_property_test_suite"
	suite.Run(t, s)
}

type (
	FromForPropertyTestSuite struct {
		suite.Suite
		resourceBase string
	}
	testForPropertyFilter struct {
		t         *testing.T
		refAssert func(t *testing.T, prop prop.Property, refProp prop.Property)
	}
)

func (s *FromForPropertyTestSuite) TestFilterRef() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name        string
		getResource func() *prop.Resource
		getRef      func() *prop.Resource
		expect      func(t *testing.T, prop prop.Property, refProp prop.Property)
	}{
		{
			name: "filter with identical resources",
			getResource: func() *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			getRef: func() *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			expect: func(t *testing.T, prop prop.Property, refProp prop.Property) {
				assert.Equal(t, prop.Attribute().ID(), refProp.Attribute().ID())
				assert.True(t, prop.Matches(refProp))
			},
		},
		{
			name: "filter with modified resources",
			getResource: func() *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
			getRef: func() *prop.Resource {
				return s.mustResource("/user_002.json", resourceType)
			},
			expect: func(t *testing.T, prop prop.Property, refProp prop.Property) {
				switch prop.Attribute().ID() {
				case "schemas",
					"schemas$elem",
					"id",
					"externalId",
					"urn:ietf:params:scim:schemas:core:2.0:User:userName",
					"urn:ietf:params:scim:schemas:core:2.0:User:name",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.formatted",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.familyName",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.givenName",
					"urn:ietf:params:scim:schemas:core:2.0:User:name.honorificPrefix",
					"urn:ietf:params:scim:schemas:core:2.0:User:displayName",
					"urn:ietf:params:scim:schemas:core:2.0:User:profileUrl",
					"urn:ietf:params:scim:schemas:core:2.0:User:userType",
					"urn:ietf:params:scim:schemas:core:2.0:User:preferredLanguage",
					"urn:ietf:params:scim:schemas:core:2.0:User:locale",
					"urn:ietf:params:scim:schemas:core:2.0:User:timezone",
					"urn:ietf:params:scim:schemas:core:2.0:User:active":
					assert.NotNil(t, refProp)
					assert.Equal(t, prop.Attribute().ID(), refProp.Attribute().ID())
					assert.True(t, prop.Matches(refProp))
				case "meta",
					"meta.resourceType",
					"meta.created",
					"meta.lastModified",
					"meta.version",
					"urn:ietf:params:scim:schemas:core:2.0:User:emails":
					assert.NotNil(t, refProp)
					assert.Equal(t, prop.Attribute().ID(), refProp.Attribute().ID())
					assert.False(t, prop.Matches(refProp))
				case "urn:ietf:params:scim:schemas:core:2.0:User:emails.value":
					if "imulab@foo.com" == prop.Raw() {
						assert.NotNil(t, refProp)
						assert.Equal(t, prop.Attribute().ID(), refProp.Attribute().ID())
						assert.True(t, prop.Matches(refProp))
					} else {
						assert.Nil(t, refProp)
					}
				}
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			f := FromForProperty(&testForPropertyFilter{t: t, refAssert: test.expect})
			err := f.FilterRef(context.Background(), test.getResource(), test.getRef())
			assert.Nil(s.T(), err)
		})
	}
}

func (s *FromForPropertyTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *FromForPropertyTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *FromForPropertyTestSuite) mustSchema(filePath string) *spec.Schema {
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

func (tf *testForPropertyFilter) Supports(attribute *spec.Attribute) bool {
	return true
}

func (tf *testForPropertyFilter) Filter(ctx context.Context, resource *prop.Resource, property prop.Property) error {
	return nil
}

func (tf *testForPropertyFilter) FieldRef(ctx context.Context, resource *prop.Resource, property prop.Property,
	refResource *prop.Resource, refProperty prop.Property) error {
	tf.refAssert(tf.t, property, refProperty)
	return nil
}
