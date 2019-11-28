package expr

import (
	"encoding/json"
	"github.com/imulab/go-scim/src/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestPathAncestry(t *testing.T) {
	s := new(AncestryTestSuite)
	s.resourceBase = "../../tests/ancestry_test_suite"
	suite.Run(t, s)
}

type AncestryTestSuite struct {
	suite.Suite
	resourceBase 	string
}

func (s *AncestryTestSuite) TestIsMember() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct{
		name 		string
		addPath func(t *testing.T, family *PathAncestry)
		getTestPath func(t *testing.T) *Expression
		expect		func(t *testing.T, r bool)
	}{
		{
			name: 	"added single path is member",
			addPath: func(t *testing.T, family *PathAncestry) {
				family.Add(NewPath("emails"))
			},
			getTestPath: func(t *testing.T) *Expression{
				return NewPath("emails")
			},
			expect: func(t *testing.T, r bool) {
				assert.True(t, r)
			},
		},
		{
			name: 	"never added single path is not member",
			addPath: func(t *testing.T, family *PathAncestry) {
				family.Add(NewPath("emails"))
			},
			getTestPath: func(t *testing.T) *Expression{
				return NewPath("phoneNumbers")
			},
			expect: func(t *testing.T, r bool) {
				assert.False(t, r)
			},
		},
		{
			name: 	"added duplex path is member",
			addPath: func(t *testing.T, family *PathAncestry) {
				family.Add(NewPath("emails", "value"))
			},
			getTestPath: func(t *testing.T) *Expression{
				return NewPath("emails", "value")
			},
			expect: func(t *testing.T, r bool) {
				assert.True(t, r)
			},
		},
		{
			name: 	"never added duplex path is not member",
			addPath: func(t *testing.T, family *PathAncestry) {
				family.Add(NewPath("emails", "value"))
			},
			getTestPath: func(t *testing.T) *Expression{
				return NewPath("emails", "primary")
			},
			expect: func(t *testing.T, r bool) {
				assert.False(t, r)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			family := NewPathFamily(resourceType)
			test.addPath(t, family)
			result := family.IsMember(test.getTestPath(t))
			test.expect(t, result)
		})
	}
}

func (s *AncestryTestSuite) TestIsAncestor() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct{
		name 		string
		addPath func(t *testing.T, family *PathAncestry)
		getTestPath func(t *testing.T) *Expression
		expect		func(t *testing.T, r bool)
	}{
		{
			name: 	"added single path is not ancestor",
			addPath: func(t *testing.T, family *PathAncestry) {
				family.Add(NewPath("emails"))
			},
			getTestPath: func(t *testing.T) *Expression{
				return NewPath("emails")
			},
			expect: func(t *testing.T, r bool) {
				assert.False(t, r)
			},
		},
		{
			name: 	"added duplex path is not ancestor",
			addPath: func(t *testing.T, family *PathAncestry) {
				family.Add(NewPath("emails", "value"))
			},
			getTestPath: func(t *testing.T) *Expression{
				return NewPath("emails", "value")
			},
			expect: func(t *testing.T, r bool) {
				assert.False(t, r)
			},
		},
		{
			name: 	"container of added path is ancestor",
			addPath: func(t *testing.T, family *PathAncestry) {
				family.Add(NewPath("emails", "value"))
			},
			getTestPath: func(t *testing.T) *Expression{
				return NewPath("emails")
			},
			expect: func(t *testing.T, r bool) {
				assert.True(t, r)
			},
		},
		{
			name: 	"offspring is not ancestor",
			addPath: func(t *testing.T, family *PathAncestry) {
				family.Add(NewPath("emails"))
			},
			getTestPath: func(t *testing.T) *Expression{
				return NewPath("emails", "primary")
			},
			expect: func(t *testing.T, r bool) {
				assert.False(t, r)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			family := NewPathFamily(resourceType)
			test.addPath(t, family)
			result := family.IsAncestor(test.getTestPath(t))
			test.expect(t, result)
		})
	}
}

func (s *AncestryTestSuite) TestIsOffspring() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct{
		name 		string
		addPath func(t *testing.T, family *PathAncestry)
		getTestPath func(t *testing.T) *Expression
		expect		func(t *testing.T, r bool)
	}{
		{
			name: 	"added single path is not offspring",
			addPath: func(t *testing.T, family *PathAncestry) {
				family.Add(NewPath("emails"))
			},
			getTestPath: func(t *testing.T) *Expression{
				return NewPath("emails")
			},
			expect: func(t *testing.T, r bool) {
				assert.False(t, r)
			},
		},
		{
			name: 	"added duplex path is not offspring",
			addPath: func(t *testing.T, family *PathAncestry) {
				family.Add(NewPath("emails", "value"))
			},
			getTestPath: func(t *testing.T) *Expression{
				return NewPath("emails", "value")
			},
			expect: func(t *testing.T, r bool) {
				assert.False(t, r)
			},
		},
		{
			name: 	"child of added path is offspring",
			addPath: func(t *testing.T, family *PathAncestry) {
				family.Add(NewPath("emails"))
			},
			getTestPath: func(t *testing.T) *Expression{
				return NewPath("emails", "value")
			},
			expect: func(t *testing.T, r bool) {
				assert.True(t, r)
			},
		},
		{
			name: 	"ancestor is not offspring",
			addPath: func(t *testing.T, family *PathAncestry) {
				family.Add(NewPath("emails", "value"))
			},
			getTestPath: func(t *testing.T) *Expression{
				return NewPath("emails")
			},
			expect: func(t *testing.T, r bool) {
				assert.False(t, r)
			},
		},
		{
			name: 	"no offspring in empty family",
			addPath: func(t *testing.T, family *PathAncestry) {},
			getTestPath: func(t *testing.T) *Expression {
				return NewPath("emails")
			},
			expect: func(t *testing.T, r bool) {
				assert.False(t, r)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			family := NewPathFamily(resourceType)
			test.addPath(t, family)
			result := family.IsOffspring(test.getTestPath(t))
			test.expect(t, result)
		})
	}
}

func (s *AncestryTestSuite) mustResourceType(filePath string) *core.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(core.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *AncestryTestSuite) mustSchema(filePath string) *core.Schema {
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