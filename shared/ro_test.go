package shared

import (
	"context"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

func TestIdAssignment_AssignValue(t *testing.T) {
	ro := NewIdAssignment()
	r := &Resource{Complex{}}
	ctx := context.Background()
	ro.AssignValue(r, ctx)
	assert.NotEmpty(t, r.GetId())
}

func TestMetaAssignment_AssignValue(t *testing.T) {
	properties := &mapPropertySource{
		data: map[string]interface{}{
			"scim.resources.user.locationBase":  "http://scim.com/Users",
			"scim.resources.group.locationBase": "http://scim.com/Groups",
		},
	}
	ro := NewMetaAssignment(properties, UserResourceType)

	for _, test := range []struct {
		r         *Resource
		assertion func(meta map[string]interface{}, err error)
	}{
		{
			&Resource{Complex{"id": "foo"}},
			func(meta map[string]interface{}, err error) {
				assert.Equal(t, UserResourceType, meta["resourceType"])
				assert.Equal(t, "http://scim.com/Users/foo", meta["location"])
				assert.NotEmpty(t, meta["created"])
				assert.NotEmpty(t, meta["lastModified"])
				assert.NotEmpty(t, meta["version"])
			},
		},
		{
			&Resource{Complex{
				"id": "foo",
				"meta": map[string]interface{}{
					"resourceType": UserResourceType,
					"location":     "http://scim.com/Users/foo",
					"created":      "2017-04-13T01:50:13Z",
					"lastModified": "2017-04-13T01:50:13Z",
					"version":      "W/\"o1tXvt0DW3gT6IekR3h+h0oBqQo=\"",
				},
			}},
			func(meta map[string]interface{}, err error) {
				assert.Equal(t, UserResourceType, meta["resourceType"])
				assert.Equal(t, "http://scim.com/Users/foo", meta["location"])
				assert.Equal(t, "2017-04-13T01:50:13Z", meta["created"])
				assert.NotEqual(t, "2017-04-13T01:50:13Z", meta["lastModified"])
				assert.NotEqual(t, "W/\"o1tXvt0DW3gT6IekR3h+h0oBqQo=\"", meta["version"])
			},
		},
	} {
		ctx := context.Background()
		err := ro.AssignValue(test.r, ctx)
		test.assertion(test.r.Complex["meta"].(map[string]interface{}), err)
	}
}

func TestGroupAssignment_AssignValue(t *testing.T) {
	repo := &roTestMockDB{}
	repo.init()
	ro := NewGroupAssignment(repo)

	for _, test := range []struct {
		userId    string
		assertion func(group []interface{})
	}{
		{
			"u_001",
			func(group []interface{}) {
				assert.True(t, reflect.DeepEqual(group, []interface{}{
					map[string]interface{}{
						"value":   "g_001",
						"$ref":    "http://scim.com/Groups/g_001",
						"display": "Group001",
						"type":    "direct",
					},
					map[string]interface{}{
						"value":   "g_002",
						"$ref":    "http://scim.com/Groups/g_002",
						"display": "Group002",
						"type":    "direct",
					},
				}))
			},
		},
		{
			"u_002",
			func(group []interface{}) {
				assert.True(t, reflect.DeepEqual(group, []interface{}{
					map[string]interface{}{
						"value":   "g_001",
						"$ref":    "http://scim.com/Groups/g_001",
						"display": "Group001",
						"type":    "direct",
					},
					map[string]interface{}{
						"value":   "g_003",
						"$ref":    "http://scim.com/Groups/g_003",
						"display": "Group003",
						"type":    "direct",
					},
					map[string]interface{}{
						"value":   "g_002",
						"$ref":    "http://scim.com/Groups/g_002",
						"display": "Group002",
						"type":    "indirect",
					},
				}))
			},
		},
		{
			"u_003",
			func(group []interface{}) {
				assert.True(t, reflect.DeepEqual(group, []interface{}{
					map[string]interface{}{
						"value":   "g_002",
						"$ref":    "http://scim.com/Groups/g_002",
						"display": "Group002",
						"type":    "direct",
					},
				}))
			},
		},
		{
			"u_004",
			func(group []interface{}) {
				assert.Equal(t, 0, len(group))
			},
		},
	} {
		r := &Resource{Complex{"id": test.userId}}
		ctx := context.Background()
		err := ro.AssignValue(r, ctx)
		assert.Nil(t, err)
		test.assertion(r.Complex["groups"].([]interface{}))
	}
}

// Mock property source
type mapPropertySource struct{ data map[string]interface{} }

func (s *mapPropertySource) Get(key string) interface{}  { return s.data[key] }
func (s *mapPropertySource) GetString(key string) string { return s.data[key].(string) }
func (s *mapPropertySource) GetInt(key string) int       { return s.data[key].(int) }
func (s *mapPropertySource) GetBool(key string) bool     { return s.data[key].(bool) }

// Mock Database that was created to support TestGroupAssignment_AssignValue
type roTestMockDB struct{ data map[string]DataProvider }

func (r *roTestMockDB) Create(provider DataProvider, ctx context.Context) error { return Error.Text("not implemented") }
func (r *roTestMockDB) Get(id, version string, ctx context.Context) (DataProvider, error) {
	return nil, Error.Text("not implemented")
}
func (r *roTestMockDB) GetAll(context.Context) ([]Complex, error) { return nil, Error.Text("not implemented") }
func (r *roTestMockDB) Count(query string, ctx context.Context) (int, error) { return 0, Error.Text("not implemented") }
func (r *roTestMockDB) Update(id, version string, provider DataProvider, ctx context.Context) error {
	return Error.Text("not implemented")
}
func (r *roTestMockDB) Delete(id, version string, ctx context.Context) error { return Error.Text("not implemented") }
func (r *roTestMockDB) Search(payload SearchRequest, ctx context.Context) (*ListResponse, error) {
	// supports test user u_001, u_002, u_003, u_004
	// supports test group g_001, g_002, g_003
	// membership layout: g_001(u_001, u_002), g_002(u_001, u_003, g_003), g_003(u_002)
	// hence:
	// u_001 is a member of g_001(direct), g_002(direct)
	// u_002 is a member of g_001(direct), g_002(indirect), g_003(direct)
	// u_003 is a member of g_002(direct)
	// u_004 is not a member of any group
	switch {
	case strings.Contains(payload.Filter, "u_001"):
		return r.createResponse(r.data["g_001"], r.data["g_002"]), nil

	case strings.Contains(payload.Filter, "u_002"):
		return r.createResponse(r.data["g_001"], r.data["g_003"]), nil

	case strings.Contains(payload.Filter, "u_003"):
		return r.createResponse(r.data["g_002"]), nil

	case strings.Contains(payload.Filter, "u_004"):
		return r.createResponse(), nil

	case strings.Contains(payload.Filter, "g_001"):
		return r.createResponse(), nil

	case strings.Contains(payload.Filter, "g_002"):
		return r.createResponse(), nil

	case strings.Contains(payload.Filter, "g_003"):
		return r.createResponse(r.data["g_002"]), nil
	}
	return nil, nil
}
func (r *roTestMockDB) init() {
	u001 := map[string]interface{}{
		"id":          "u_001",
		"displayName": "User001",
		"meta": map[string]interface{}{
			"location": "http://scim.com/Users/u_001",
		},
	}
	u002 := map[string]interface{}{
		"id":          "u_002",
		"displayName": "User002",
		"meta": map[string]interface{}{
			"location": "http://scim.com/Users/u_002",
		},
	}
	u003 := map[string]interface{}{
		"id":          "u_003",
		"displayName": "User003",
		"meta": map[string]interface{}{
			"location": "http://scim.com/Users/u_003",
		},
	}
	g003 := map[string]interface{}{
		"id":          "g_003",
		"displayName": "Group003",
		"meta": map[string]interface{}{
			"location": "http://scim.com/Groups/g_003",
		},
	}

	r.data = map[string]DataProvider{
		"g_001": &Resource{
			Complex{
				"id":          "g_001",
				"displayName": "Group001",
				"members":     []interface{}{u001, u002},
				"meta": map[string]interface{}{
					"location": "http://scim.com/Groups/g_001",
				},
			},
		},
		"g_002": &Resource{
			Complex{
				"id":          "g_002",
				"displayName": "Group002",
				"members":     []interface{}{u001, u003, g003},
				"meta": map[string]interface{}{
					"location": "http://scim.com/Groups/g_002",
				},
			},
		},
		"g_003": &Resource{
			Complex{
				"id":          "g_003",
				"displayName": "Group003",
				"members":     []interface{}{u002},
				"meta": map[string]interface{}{
					"location": "http://scim.com/Groups/g_003",
				},
			},
		},
	}
}
func (r *roTestMockDB) createResponse(d ...DataProvider) *ListResponse {
	return &ListResponse{Resources: d}
}
