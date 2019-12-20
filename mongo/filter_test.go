package mongo

import (
	"encoding/json"
	"github.com/imulab/go-scim/core/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"io/ioutil"
	"os"
	"testing"
)

func TestTransformFilter(t *testing.T) {
	s := new(TransformFilterTestSuite)
	s.resourceBase = "./internal/transform_filter_test_suite"
	suite.Run(t, s)
}

type TransformFilterTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *TransformFilterTestSuite) TestTransform() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name   string
		filter string
		expect func(t *testing.T, doc bsonx.Val, err error)
	}{
		{
			name:   "top level string property pr",
			filter: "userName pr",
			expect: func(t *testing.T, doc bsonx.Val, err error) {
				assert.Nil(t, err)
				raw, err := bson.MarshalExtJSON(doc, true, false)
				assert.Nil(t, err)
				expect := `{"userName":{"$and":[{"$exists":true},{"$ne":null},{"$ne":""}]}}`
				assert.JSONEq(t, expect, string(raw))
			},
		},
		{
			name:   "second level string property pr",
			filter: "name.familyName pr",
			expect: func(t *testing.T, doc bsonx.Val, err error) {
				assert.Nil(t, err)
				raw, err := bson.MarshalExtJSON(doc, true, false)
				assert.Nil(t, err)
				expect := `{"name.familyName":{"$and":[{"$exists":true},{"$ne":null},{"$ne":""}]}}`
				assert.JSONEq(t, expect, string(raw))
			},
		},
		{
			name:   "multiValued pr",
			filter: "emails pr",
			expect: func(t *testing.T, doc bsonx.Val, err error) {
				assert.Nil(t, err)
				raw, err := bson.MarshalExtJSON(doc, true, false)
				assert.Nil(t, err)
				expect := `{"emails":{"$and":[{"$exists":true},{"$ne":null},{"$nor":[{"$size":{"$numberInt":"0"}}]}]}}`
				assert.JSONEq(t, expect, string(raw))
			},
		},
		{
			name:   "multiValued second level pr",
			filter: "emails.value pr",
			expect: func(t *testing.T, doc bsonx.Val, err error) {
				assert.Nil(t, err)
				raw, err := bson.MarshalExtJSON(doc, true, false)
				assert.Nil(t, err)
				expect := `{"emails":{"$elemMatch":{"value":{"$and":[{"$exists":true},{"$ne":null},{"$ne":""}]}}}}`
				assert.JSONEq(t, expect, string(raw))
			},
		},
		{
			name:   "multiValued eq",
			filter: "schemas eq \"foobar\"",
			expect: func(t *testing.T, doc bsonx.Val, err error) {
				assert.Nil(t, err)
				raw, err := bson.MarshalExtJSON(doc, true, false)
				assert.Nil(t, err)
				expect := `{"schemas":{"$elemMatch":{"$eq":"foobar"}}}`
				assert.JSONEq(t, expect, string(raw))
			},
		},
		{
			name:   "multiValued second level eq",
			filter: "emails.value eq \"foo@bar.com\"",
			expect: func(t *testing.T, doc bsonx.Val, err error) {
				assert.Nil(t, err)
				raw, err := bson.MarshalExtJSON(doc, true, false)
				assert.Nil(t, err)
				expect := `{"emails":{"$elemMatch":{"value":{"$regularExpression":{"pattern":"^foo@bar.com$","options":"i"}}}}}`
				assert.JSONEq(t, expect, string(raw))
			},
		},
		{
			name:   "top level ne",
			filter: "userName ne \"imulab\"",
			expect: func(t *testing.T, doc bsonx.Val, err error) {
				assert.Nil(t, err)
				raw, err := bson.MarshalExtJSON(doc, true, false)
				assert.Nil(t, err)
				expect := `{"userName":{"$regularExpression":{"pattern":"^((?!imulab$).)","options":"i"}}}`
				assert.JSONEq(t, expect, string(raw))
			},
		},
		{
			name:   "second level ne",
			filter: "name.familyName ne \"Q\"",
			expect: func(t *testing.T, doc bsonx.Val, err error) {
				assert.Nil(t, err)
				raw, err := bson.MarshalExtJSON(doc, true, false)
				assert.Nil(t, err)
				expect := `{"name.familyName":{"$regularExpression":{"pattern":"^((?!Q$).)","options":"i"}}}`
				assert.JSONEq(t, expect, string(raw))
			},
		},
		{
			name:   "multiValued ne",
			filter: "schemas ne \"foobar\"",
			expect: func(t *testing.T, doc bsonx.Val, err error) {
				assert.Nil(t, err)
				raw, err := bson.MarshalExtJSON(doc, true, false)
				assert.Nil(t, err)
				expect := `{"schemas":{"$elemMatch":{"$ne":"foobar"}}}`
				assert.JSONEq(t, expect, string(raw))
			},
		},
		{
			name:   "multiValued second level ne",
			filter: "emails.value ne \"foo@bar.com\"",
			expect: func(t *testing.T, doc bsonx.Val, err error) {
				assert.Nil(t, err)
				raw, err := bson.MarshalExtJSON(doc, true, false)
				assert.Nil(t, err)
				expect := `{"emails":{"$elemMatch":{"value":{"$regularExpression":{"pattern":"^((?!foo@bar.com$).)","options":"i"}}}}}`
				assert.JSONEq(t, expect, string(raw))
			},
		},
		{
			name:   "dateTime gt",
			filter: "meta.created gt \"2019-12-20T04:40:00\"",
			expect: func(t *testing.T, doc bsonx.Val, err error) {
				assert.Nil(t, err)
				raw, err := bson.MarshalExtJSON(doc, true, false)
				assert.Nil(t, err)
				expect := `{"meta.created":{"$gt":{"$date":{"$numberLong":"1576816800000"}}}}`
				assert.JSONEq(t, expect, string(raw))
			},
		},
		{
			name:   "logical operator",
			filter: "(userName eq \"imulab\") and (meta.created gt \"2019-12-20T04:40:00\")",
			expect: func(t *testing.T, doc bsonx.Val, err error) {
				assert.Nil(t, err)
				raw, err := bson.MarshalExtJSON(doc, true, false)
				assert.Nil(t, err)
				expect := `{"$and":[{"userName":{"$regularExpression":{"pattern":"^imulab$","options":"i"}}},{"meta.created":{"$gt":{"$date":{"$numberLong":"1576816800000"}}}}]}`
				assert.JSONEq(t, expect, string(raw))
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			v, err := TransformFilter(test.filter, resourceType)
			test.expect(t, v, err)
		})
	}
}

func (s *TransformFilterTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *TransformFilterTestSuite) mustSchema(filePath string) *spec.Schema {
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
