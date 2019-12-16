package groupsync

import (
	"context"
	"encoding/json"
	scimJSON "github.com/imulab/go-scim/pkg/core/json"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"github.com/imulab/go-scim/pkg/protocol/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestGroupSyncRefresh(t *testing.T) {
	s := new(GroupSyncRefreshTestSuite)
	s.resourceBase = "../../tests/group_sync_refresh_test_suite"
	suite.Run(t, s)
}

type GroupSyncRefreshTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *GroupSyncRefreshTestSuite) TestRefresh() {
	_ = s.mustSchema("/user_schema.json")
	_ = s.mustSchema("/group_schema.json")

	userResourceType := s.mustResourceType("/user_resource_type.json")
	groupResourceType := s.mustResourceType("/group_resource_type.json")

	tests := []struct {
		name       string
		getGroupDB func(t *testing.T) db.DB
		getUser    func(t *testing.T) *prop.Resource
		expect     func(t *testing.T, user *prop.Resource, err error)
	}{
		{
			name: "refresh user",
			getGroupDB: func(t *testing.T) db.DB {
				database := db.Memory()
				require.Nil(t, database.Insert(context.Background(), s.mustResource("/group_001.json", groupResourceType)))
				require.Nil(t, database.Insert(context.Background(), s.mustResource("/group_002.json", groupResourceType)))
				return database
			},
			getUser: func(t *testing.T) *prop.Resource {
				return s.mustResource("/user_001.json", userResourceType)
			},
			expect: func(t *testing.T, user *prop.Resource, err error) {
				assert.Nil(t, err)
				nav := user.NewFluentNavigator().FocusName("groups")
				nav.FocusIndex(0).FocusName("display")
				assert.Equal(t, "Group 001", nav.Current().Raw())
				nav.Retract().Retract()
				nav.FocusIndex(1).FocusName("display")
				assert.Equal(t, "Group 002", nav.Current().Raw())
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			database := test.getGroupDB(t)
			user := test.getUser(t)
			err := Refresher(database).Refresh(context.Background(), user)
			test.expect(t, user, err)
		})
	}
}

func (s *GroupSyncRefreshTestSuite) mustResource(filePath string, resourceType *spec.ResourceType) *prop.Resource {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	resource := prop.NewResource(resourceType)
	err = scimJSON.Deserialize(raw, resource)
	s.Require().Nil(err)

	return resource
}

func (s *GroupSyncRefreshTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *GroupSyncRefreshTestSuite) mustSchema(filePath string) *spec.Schema {
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
