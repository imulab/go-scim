package crud

import (
	"encoding/json"
	"github.com/imulab/go-scim/core/expr"
	scimJSON "github.com/imulab/go-scim/core/json"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/core/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestSortResource(t *testing.T) {
	s := new(SortResourceTestSuite)
	s.resourceBase = "./internal/sort_resource_test_suite"
	suite.Run(t, s)
}

type SortResourceTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *SortResourceTestSuite) TestSeekSortTarget() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct{
		name		string
		getResource	func() *prop.Resource
		sortBy		string
		expect		func(t *testing.T, target prop.Property, err error)
	}{
		{
			name: 	"simple target",
			getResource: func() *prop.Resource {
				return s.mustResource("/user_000.json", resourceType)
			},
			sortBy: "userName",
			expect: func(t *testing.T, target prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "userName", target.Attribute().Name())
				assert.Equal(t, "imulab", target.Raw())
			},
		},
		{
			name: 	"nested simple target",
			getResource: func() *prop.Resource {
				return s.mustResource("/user_000.json", resourceType)
			},
			sortBy: "name.familyName",
			expect: func(t *testing.T, target prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "familyName", target.Attribute().Name())
				assert.Equal(t, "Qiu", target.Raw())
			},
		},
		{
			name: 	"simple multiValued",
			getResource: func() *prop.Resource {
				return s.mustResource("/user_000.json", resourceType)
			},
			sortBy: "schemas",
			expect: func(t *testing.T, target prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "schemas$elem", target.Attribute().ID())
				assert.Equal(t, "urn:ietf:params:scim:schemas:core:2.0:User", target.Raw())
			},
		},
		{
			name: 	"multiValued with primary",
			getResource: func() *prop.Resource {
				return s.mustResource("/user_000.json", resourceType)
			},
			sortBy: "emails.value",
			expect: func(t *testing.T, target prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "value", target.Attribute().Name())
				assert.Equal(t, "imulab@foo.com", target.Raw())
			},
		},
		{
			name: 	"multiValued without primary",
			getResource: func() *prop.Resource {
				return s.mustResource("/user_000.json", resourceType)
			},
			sortBy: "phoneNumbers.value",
			expect: func(t *testing.T, target prop.Property, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "value", target.Attribute().Name())
				assert.Equal(t, "123-45678", target.Raw())
			},
		},
		{
			name: 	"invalid",
			getResource: func() *prop.Resource {
				return s.mustResource("/user_000.json", resourceType)
			},
			sortBy: "emails",
			expect: func(t *testing.T, target prop.Property, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource()
			by, err := expr.CompilePath(test.sortBy)
			assert.Nil(t, err)
			prop, err := SeekSortTarget(resource, by)
			test.expect(t, prop, err)
		})
	}
}

func (s *SortResourceTestSuite) TestSort() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name         string
		sort         Sort
		getResources func() []*prop.Resource
		expect       func(t *testing.T, resources []*prop.Resource, err error)
	}{
		{
			name: 	"sort sorted",
			sort: 	Sort{
				By:    "userName",
				Order: SortAsc,
			},
			getResources: func() []*prop.Resource {
				return []*prop.Resource{
					s.mustResource("/user_001.json", resourceType),
					s.mustResource("/user_002.json", resourceType),
					s.mustResource("/user_003.json", resourceType),
					s.mustResource("/user_004.json", resourceType),
					s.mustResource("/user_005.json", resourceType),
				}
			},
			expect: func(t *testing.T, resources []*prop.Resource, err error) {
				assert.Nil(t, err)
				userNames := make([]interface{}, 0)
				for _, r := range resources {
					p, err := r.NewNavigator().FocusName("userName")
					assert.Nil(t, err)
					userNames = append(userNames, p.Raw())
				}
				assert.Equal(t, "A", userNames[0])
				assert.Equal(t, "B", userNames[1])
				assert.Equal(t, "C", userNames[2])
				assert.Equal(t, "D", userNames[3])
				assert.Equal(t, "E", userNames[4])
			},
		},
		{
			name: 	"sort random order",
			sort: 	Sort{
				By:    "userName",
				Order: SortAsc,
			},
			getResources: func() []*prop.Resource {
				return []*prop.Resource{
					s.mustResource("/user_003.json", resourceType),
					s.mustResource("/user_001.json", resourceType),
					s.mustResource("/user_005.json", resourceType),
					s.mustResource("/user_002.json", resourceType),
					s.mustResource("/user_004.json", resourceType),
				}
			},
			expect: func(t *testing.T, resources []*prop.Resource, err error) {
				assert.Nil(t, err)
				userNames := make([]interface{}, 0)
				for _, r := range resources {
					p, err := r.NewNavigator().FocusName("userName")
					assert.Nil(t, err)
					userNames = append(userNames, p.Raw())
				}
				assert.Equal(t, "A", userNames[0])
				assert.Equal(t, "B", userNames[1])
				assert.Equal(t, "C", userNames[2])
				assert.Equal(t, "D", userNames[3])
				assert.Equal(t, "E", userNames[4])
			},
		},
		{
			name: 	"sort random order descending",
			sort: 	Sort{
				By:    "userName",
				Order: SortDesc,
			},
			getResources: func() []*prop.Resource {
				return []*prop.Resource{
					s.mustResource("/user_003.json", resourceType),
					s.mustResource("/user_001.json", resourceType),
					s.mustResource("/user_005.json", resourceType),
					s.mustResource("/user_002.json", resourceType),
					s.mustResource("/user_004.json", resourceType),
				}
			},
			expect: func(t *testing.T, resources []*prop.Resource, err error) {
				assert.Nil(t, err)
				userNames := make([]interface{}, 0)
				for _, r := range resources {
					p, err := r.NewNavigator().FocusName("userName")
					assert.Nil(t, err)
					userNames = append(userNames, p.Raw())
				}
				assert.Equal(t, "E", userNames[0])
				assert.Equal(t, "D", userNames[1])
				assert.Equal(t, "C", userNames[2])
				assert.Equal(t, "B", userNames[3])
				assert.Equal(t, "A", userNames[4])
			},
		},
		{
			name: 	"sort reversed",
			sort: 	Sort{
				By:    "userName",
				Order: SortAsc,
			},
			getResources: func() []*prop.Resource {
				return []*prop.Resource{
					s.mustResource("/user_005.json", resourceType),
					s.mustResource("/user_004.json", resourceType),
					s.mustResource("/user_003.json", resourceType),
					s.mustResource("/user_002.json", resourceType),
					s.mustResource("/user_001.json", resourceType),
				}
			},
			expect: func(t *testing.T, resources []*prop.Resource, err error) {
				assert.Nil(t, err)
				userNames := make([]interface{}, 0)
				for _, r := range resources {
					p, err := r.NewNavigator().FocusName("userName")
					assert.Nil(t, err)
					userNames = append(userNames, p.Raw())
				}
				assert.Equal(t, "A", userNames[0])
				assert.Equal(t, "B", userNames[1])
				assert.Equal(t, "C", userNames[2])
				assert.Equal(t, "D", userNames[3])
				assert.Equal(t, "E", userNames[4])
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resources := test.getResources()
			err := test.sort.Sort(resources)
			test.expect(t, resources, err)
		})
	}
}

func (s *SortResourceTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *SortResourceTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *SortResourceTestSuite) mustSchema(filePath string) *spec.Schema {
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
