package groupsync

import (
	"context"
	"encoding/json"
	"github.com/imulab/go-scim/pkg/v2/db"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestSyncService(t *testing.T) {
	s := new(SyncServiceTestSuite)
	suite.Run(t, s)
}

type SyncServiceTestSuite struct {
	suite.Suite
	userResourceType  *spec.ResourceType
	groupResourceType *spec.ResourceType
}

func (s *SyncServiceTestSuite) TestSyncGroupPropertyForUser() {
	tests := []struct {
		name       string
		getUser    func(t *testing.T) *prop.Resource
		getGroupDB func(t *testing.T) db.DB
		expect     func(t *testing.T, user *prop.Resource, err error)
	}{
		{
			name: "default",
			getUser: func(t *testing.T) *prop.Resource {
				u := prop.NewResource(s.userResourceType)
				assert.False(t, u.Navigator().Replace(map[string]interface{}{
					"schemas": []interface{}{"urn:ietf:params:scim:schemas:core:2.0:User"},
					"id":      "u1",
				}).HasError())
				return u
			},
			getGroupDB: func(t *testing.T) db.DB {
				database := db.Memory()
				for _, data := range []map[string]interface{}{
					{
						"schemas": []interface{}{"urn:ietf:params:scim:schemas:core:2.0:Group"},
						"id":      "g1",
						"members": []interface{}{
							map[string]interface{}{
								"value":   "u1",
								"$ref":    "/Users/u1",
								"display": "u1",
							},
							map[string]interface{}{
								"value":   "u2",
								"$ref":    "/Users/u2",
								"display": "u2",
							},
						},
					},
					{
						"schemas": []interface{}{"urn:ietf:params:scim:schemas:core:2.0:Group"},
						"id":      "g2",
						"members": []interface{}{
							map[string]interface{}{
								"value":   "u3",
								"$ref":    "/Users/u2",
								"display": "u3",
							},
							map[string]interface{}{
								"value":   "g1",
								"$ref":    "/Groups/g1",
								"display": "g1",
							},
						},
					},
				} {
					g := prop.NewResource(s.groupResourceType)
					assert.False(t, g.Navigator().Replace(data).HasError())
					assert.Nil(t, database.Insert(context.Background(), g))
				}
				return database
			},
			expect: func(t *testing.T, user *prop.Resource, err error) {
				assert.Nil(t, err)

				groupIndex := map[string]struct{}{}
				{
					_ = user.Navigator().Dot("groups").ForEachChild(func(_ int, child prop.Property) error {
						p, _ := child.ChildAtIndex("value")
						groupIndex[p.Raw().(string)] = struct{}{}
						return nil
					})
				}
				assert.Len(t, groupIndex, 2)

				_, hasG1 := groupIndex["g1"]
				assert.True(t, hasG1)

				_, hasG2 := groupIndex["g2"]
				assert.True(t, hasG2)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			user := test.getUser(t)
			service := NewSyncService(test.getGroupDB(t))
			err := service.SyncGroupPropertyForUser(context.Background(), user)
			test.expect(t, user, err)
		})
	}
}

func (s *SyncServiceTestSuite) SetupSuite() {
	for _, each := range []struct {
		filepath  string
		structure interface{}
		post      func(parsed interface{})
	}{
		{
			filepath:  "../../../public/schemas/core_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../../public/schemas/user_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../../public/schemas/group_schema.json",
			structure: new(spec.Schema),
			post: func(parsed interface{}) {
				spec.Schemas().Register(parsed.(*spec.Schema))
			},
		},
		{
			filepath:  "../../../public/resource_types/user_resource_type.json",
			structure: new(spec.ResourceType),
			post: func(parsed interface{}) {
				s.userResourceType = parsed.(*spec.ResourceType)
			},
		},
		{
			filepath:  "../../../public/resource_types/group_resource_type.json",
			structure: new(spec.ResourceType),
			post: func(parsed interface{}) {
				s.groupResourceType = parsed.(*spec.ResourceType)
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
