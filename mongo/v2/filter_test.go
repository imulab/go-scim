package v2

import (
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"os"
	"testing"
)

func TestTransformFilter(t *testing.T) {
	s := new(TransformFilterTestSuite)
	suite.Run(t, s)
}

type TransformFilterTestSuite struct {
	suite.Suite
	resourceType *spec.ResourceType
}

func (s *TransformFilterTestSuite) TestTransform() {
	tests := []struct {
		name   string
		filter string
		expect func(t *testing.T, extJson string, err error)
	}{
		{
			name:   "top level string property pr",
			filter: "userName pr",
			expect: func(t *testing.T, extJson string, err error) {
				assert.Nil(t, err)
				expect := `{"$and":[{"userName":{"$exists":true}},{"userName":{"$ne":null}},{"userName":{"$ne":""}}]}`
				assert.JSONEq(t, expect, extJson)
			},
		},
		{
			name:   "second level string property pr",
			filter: "name.familyName pr",
			expect: func(t *testing.T, extJson string, err error) {
				assert.Nil(t, err)
				expect := `{"$and":[{"name.familyName":{"$exists":true}},{"name.familyName":{"$ne":null}},{"name.familyName":{"$ne":""}}]}`
				assert.JSONEq(t, expect, extJson)
			},
		},
		{
			name:   "multiValued pr",
			filter: "emails pr",
			expect: func(t *testing.T, extJson string, err error) {
				assert.Nil(t, err)
				expect := `{"$and":[{"emails":{"$exists":true}},{"emails":{"$ne":null}},{"emails":{"$nor":[{"$size":{"$numberInt":"0"}}]}}]}`
				assert.JSONEq(t, expect, extJson)
			},
		},
		{
			name:   "multiValued second level pr",
			filter: "emails.value pr",
			expect: func(t *testing.T, extJson string, err error) {
				assert.Nil(t, err)
				expect := `{"emails":{"$elemMatch":{"$and":[{"value":{"$exists":true}},{"value":{"$ne":null}},{"value":{"$ne":""}}]}}}`
				assert.JSONEq(t, expect, extJson)
			},
		},
		{
			name:   "multiValued eq",
			filter: "schemas eq \"foobar\"",
			expect: func(t *testing.T, extJson string, err error) {
				assert.Nil(t, err)
				expect := `{"schemas":{"$elemMatch":{"$eq":"foobar"}}}`
				assert.JSONEq(t, expect, extJson)
			},
		},
		{
			name:   "multiValued second level eq",
			filter: "emails.value eq \"foo@bar.com\"",
			expect: func(t *testing.T, extJson string, err error) {
				assert.Nil(t, err)
				expect := `{"emails":{"$elemMatch":{"value":{"$regularExpression":{"pattern":"^foo@bar.com$","options":"i"}}}}}`
				assert.JSONEq(t, expect, extJson)
			},
		},
		{
			name:   "top level ne",
			filter: "userName ne \"imulab\"",
			expect: func(t *testing.T, extJson string, err error) {
				assert.Nil(t, err)
				expect := `{"userName":{"$regularExpression":{"pattern":"^((?!imulab$).)","options":"i"}}}`
				assert.JSONEq(t, expect, extJson)
			},
		},
		{
			name:   "second level ne",
			filter: "name.familyName ne \"Q\"",
			expect: func(t *testing.T, extJson string, err error) {
				assert.Nil(t, err)
				expect := `{"name.familyName":{"$regularExpression":{"pattern":"^((?!Q$).)","options":"i"}}}`
				assert.JSONEq(t, expect, extJson)
			},
		},
		{
			name:   "multiValued ne",
			filter: "schemas ne \"foobar\"",
			expect: func(t *testing.T, extJson string, err error) {
				assert.Nil(t, err)
				expect := `{"schemas":{"$elemMatch":{"$ne":"foobar"}}}`
				assert.JSONEq(t, expect, extJson)
			},
		},
		{
			name:   "multiValued second level ne",
			filter: "emails.value ne \"foo@bar.com\"",
			expect: func(t *testing.T, extJson string, err error) {
				assert.Nil(t, err)
				expect := `{"emails":{"$elemMatch":{"value":{"$regularExpression":{"pattern":"^((?!foo@bar.com$).)","options":"i"}}}}}`
				assert.JSONEq(t, expect, extJson)
			},
		},
		{
			name:   "dateTime gt",
			filter: "meta.created gt \"2019-12-20T04:40:00\"",
			expect: func(t *testing.T, extJson string, err error) {
				assert.Nil(t, err)
				expect := `{"meta.created":{"$gt":{"$date":{"$numberLong":"1576816800000"}}}}`
				assert.JSONEq(t, expect, extJson)
			},
		},
		{
			name:   "logical operator",
			filter: "(userName eq \"imulab\") and (meta.created gt \"2019-12-20T04:40:00\")",
			expect: func(t *testing.T, extJson string, err error) {
				assert.Nil(t, err)
				expect := `{"$and":[{"userName":{"$regularExpression":{"pattern":"^imulab$","options":"i"}}},{"meta.created":{"$gt":{"$date":{"$numberLong":"1576816800000"}}}}]}`
				assert.JSONEq(t, expect, extJson)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			v, err := TransformFilter(test.filter, s.resourceType)
			assert.Nil(t, err)
			raw, err := bson.MarshalExtJSON(v, true, false)
			assert.Nil(t, err)
			test.expect(t, string(raw), err)
		})
	}
}

func (s *TransformFilterTestSuite) SetupSuite() {
	for _, each := range []struct {
		filepath  string
		structure interface{}
		post      func(parsed interface{})
	}{
		{
			filepath:  "../../public/schemas/core_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../public/schemas/user_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../public/resource_types/user_resource_type.json",
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
