package mongo

import (
	"encoding/json"
	scimJSON "github.com/imulab/go-scim/core/json"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/core/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"os"
	"testing"
)

func TestSerialize(t *testing.T) {
	s := new(MongoSerializerTestSuite)
	s.resourceBase = "./internal/mongo_serializer_test_suite"
	suite.Run(t, s)
}

type MongoSerializerTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *MongoSerializerTestSuite) TestSerialize() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct{
		name		string
		getResource func(t *testing.T) *prop.Resource
	}{
		{
			name: 	"serialize user",
			getResource: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", resourceType)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			raw, err := newBsonAdapter(test.getResource(t)).MarshalBSON()
			assert.Nil(t, err)
			assert.Nil(t, bson.Raw(raw).Validate())
		})
	}
}

func (s *MongoSerializerTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *MongoSerializerTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *MongoSerializerTestSuite) mustSchema(filePath string) *spec.Schema {
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
