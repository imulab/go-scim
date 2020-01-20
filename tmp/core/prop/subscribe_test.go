package prop

import (
	"encoding/json"
	"github.com/imulab/go-scim/core/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestSubscribers(t *testing.T) {
	s := new(SubscriberTestSuite)
	s.resourceBase = "../internal/subscriber_test_suite"
	suite.Run(t, s)
}

type SubscriberTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *SubscriberTestSuite) TestSyncSchema() {
	_ = s.mustSchema("/user_schema.json")
	_ = s.mustSchema("/user_enterprise_extension_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name         string
		getResource  func(t *testing.T) *Resource
		modification func(t *testing.T, r *Resource) error
		expect       func(t *testing.T, r *Resource, err error)
	}{
		{
			name: "add to a previously unassigned schema extension property will add the schema",
			getResource: func(t *testing.T) *Resource {
				resource := NewResourceOf(resourceType, map[string]interface{}{
					"schemas": []interface{}{
						"urn:ietf:params:scim:schemas:core:2.0:User",
					},
					"userName": "imulab",
				})
				return resource
			},
			modification: func(t *testing.T, r *Resource) error {
				nav := r.NewNavigator()
				_, err := nav.FocusName("urn:ietf:params:scim:schemas:extension:enterprise:2.0:User")
				require.Nil(t, err)
				p, err := nav.FocusName("employeeNumber")
				require.Nil(t, err)
				return p.Add("imulab")
			},
			expect: func(t *testing.T, r *Resource, err error) {
				assert.Nil(t, err)
				schemas, err := r.NewNavigator().FocusName("schemas")
				assert.Nil(t, err)
				assert.Equal(t, 2, schemas.(Container).CountChildren())
				// use special eq operator to assert
				ok, _ := schemas.EqualsTo("urn:ietf:params:scim:schemas:extension:enterprise:2.0:User")
				assert.True(t, ok)
			},
		},
		{
			name: "add to a previously assigned schema extension property will not repeatedly add the schema",
			getResource: func(t *testing.T) *Resource {
				resource := NewResourceOf(resourceType, map[string]interface{}{
					"schemas": []interface{}{
						"urn:ietf:params:scim:schemas:core:2.0:User",
						"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
					},
					"userName": "imulab",
					"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]interface{}{
						"employeeNumber": "imulab",
					},
				})
				return resource
			},
			modification: func(t *testing.T, r *Resource) error {
				nav := r.NewNavigator()
				_, err := nav.FocusName("urn:ietf:params:scim:schemas:extension:enterprise:2.0:User")
				require.Nil(t, err)
				p, err := nav.FocusName("organization")
				require.Nil(t, err)
				return p.Add("imulab")
			},
			expect: func(t *testing.T, r *Resource, err error) {
				assert.Nil(t, err)
				schemas, err := r.NewNavigator().FocusName("schemas")
				assert.Nil(t, err)
				assert.Equal(t, 2, schemas.(Container).CountChildren())
			},
		},
		{
			name: "schema extension urn is removed when schema becomes completely unassigned",
			getResource: func(t *testing.T) *Resource {
				resource := NewResourceOf(resourceType, map[string]interface{}{
					"schemas": []interface{}{
						"urn:ietf:params:scim:schemas:core:2.0:User",
						"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
					},
					"userName": "imulab",
					"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]interface{}{
						"employeeNumber": "imulab",
					},
				})
				return resource
			},
			modification: func(t *testing.T, r *Resource) error {
				nav := r.NewNavigator()
				_, err := nav.FocusName("urn:ietf:params:scim:schemas:extension:enterprise:2.0:User")
				require.Nil(t, err)
				p, err := nav.FocusName("employeeNumber")
				require.Nil(t, err)
				return p.Delete()
			},
			expect: func(t *testing.T, r *Resource, err error) {
				assert.Nil(t, err)
				schemas, err := r.NewNavigator().FocusName("schemas")
				assert.Nil(t, err)
				assert.Equal(t, 1, schemas.(Container).CountChildren())
				// use special eq operator to assert
				ok, _ := schemas.EqualsTo("urn:ietf:params:scim:schemas:core:2.0:User")
				assert.True(t, ok)
			},
		},
		{
			name: "schema extension urn remains when schema is not completely unassigned",
			getResource: func(t *testing.T) *Resource {
				resource := NewResourceOf(resourceType, map[string]interface{}{
					"schemas": []interface{}{
						"urn:ietf:params:scim:schemas:core:2.0:User",
						"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
					},
					"userName": "imulab",
					"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]interface{}{
						"employeeNumber": "imulab",
						"organization":   "imulab",
					},
				})
				return resource
			},
			modification: func(t *testing.T, r *Resource) error {
				nav := r.NewNavigator()
				_, err := nav.FocusName("urn:ietf:params:scim:schemas:extension:enterprise:2.0:User")
				require.Nil(t, err)
				p, err := nav.FocusName("employeeNumber")
				require.Nil(t, err)
				return p.Delete()
			},
			expect: func(t *testing.T, r *Resource, err error) {
				assert.Nil(t, err)
				schemas, err := r.NewNavigator().FocusName("schemas")
				assert.Nil(t, err)
				assert.Equal(t, 2, schemas.(Container).CountChildren())
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			r := test.getResource(t)
			err := test.modification(t, r)
			test.expect(t, r, err)
		})
	}
}

func (s *SubscriberTestSuite) TestAutoCompact() {
	_ = s.mustSchema("/user_schema.json")
	_ = s.mustSchema("/user_enterprise_extension_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name         string
		getResource  func(t *testing.T) *Resource
		modification func(t *testing.T, r *Resource) error
		expect       func(t *testing.T, r *Resource, err error)
	}{
		//{
		//	name: "compacting a simple element",
		//	getResource: func(t *testing.T) *Resource {
		//		resource := NewResourceOf(resourceType, map[string]interface{}{
		//			"schemas": []interface{}{
		//				"a", "b", "c",
		//			},
		//		})
		//		return resource
		//	},
		//	modification: func(t *testing.T, r *Resource) error {
		//		nav := r.NewNavigator()
		//		_, err := nav.FocusName("schemas")
		//		require.Nil(t, err)
		//		require.Equal(t, 3, nav.Current().(Container).CountChildren())
		//		b, err := nav.FocusIndex(1)
		//		require.Nil(t, err)
		//		return b.Delete()
		//	},
		//	expect: func(t *testing.T, r *Resource, err error) {
		//		assert.Nil(t, err)
		//		nav := r.NewNavigator()
		//		s, err := nav.FocusName("schemas")
		//		require.Nil(t, err)
		//		require.Equal(t, 2, s.(Container).CountChildren())
		//	},
		//},
		{
			name: "compacting a complex element",
			getResource: func(t *testing.T) *Resource {
				resource := NewResourceOf(resourceType, map[string]interface{}{
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
				})
				return resource
			},
			modification: func(t *testing.T, r *Resource) error {
				nav := r.NewNavigator()
				_, err := nav.FocusName("emails")
				require.Nil(t, err)
				require.Equal(t, 2, nav.Current().(Container).CountChildren())
				b, err := nav.FocusIndex(1)
				require.Nil(t, err)
				return b.(Container).ForEachChild(func(index int, child Property) error {
					return child.Delete()
				})
			},
			expect: func(t *testing.T, r *Resource, err error) {
				assert.Nil(t, err)
				nav := r.NewNavigator()
				s, err := nav.FocusName("emails")
				require.Nil(t, err)
				require.Equal(t, 1, s.(Container).CountChildren())
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			r := test.getResource(t)
			err := test.modification(t, r)
			test.expect(t, r, err)
		})
	}
}

func (s *SubscriberTestSuite) TestExclusivePrimary() {
	_ = s.mustSchema("/user_schema.json")
	_ = s.mustSchema("/user_enterprise_extension_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name         string
		getResource  func(t *testing.T) *Resource
		modification func(t *testing.T, r *Resource) error
		expect       func(t *testing.T, r *Resource, err error)
	}{
		{
			name: "turning off the primary has no effect",
			getResource: func(t *testing.T) *Resource {
				resource := NewResourceOf(resourceType, map[string]interface{}{
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
				})
				return resource
			},
			modification: func(t *testing.T, r *Resource) error {
				nav := r.NewNavigator()
				_, err := nav.FocusName("emails")
				require.Nil(t, err)
				_, err = nav.FocusIndex(0)
				require.Nil(t, err)
				_, err = nav.FocusName("primary")
				require.Nil(t, err)
				return nav.Current().Delete()
			},
			expect: func(t *testing.T, r *Resource, err error) {
				assert.Nil(t, err)
				nav := r.NewNavigator()
				{
					_, err = nav.FocusName("emails")
					assert.Nil(t, err)
					{
						for i, f := range []func(v interface{}){
							func(v interface{}) { assert.Nil(t, v) },
							func(v interface{}) { assert.Nil(t, v) },
						} {
							_, err = nav.FocusIndex(i)
							assert.Nil(t, err)
							{
								_, err = nav.FocusName("primary")
								assert.Nil(t, err)
								f(nav.Current().Raw())
								nav.Retract()
							}
							nav.Retract()
						}
					}
					nav.Retract()
				}
			},
		},
		{
			name: "setting the other primary turns off the original primary",
			getResource: func(t *testing.T) *Resource {
				resource := NewResourceOf(resourceType, map[string]interface{}{
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
				})
				return resource
			},
			modification: func(t *testing.T, r *Resource) error {
				nav := r.NewNavigator()
				_, err := nav.FocusName("emails")
				require.Nil(t, err)
				_, err = nav.FocusIndex(1)
				require.Nil(t, err)
				_, err = nav.FocusName("primary")
				require.Nil(t, err)
				return nav.Current().Replace(true)
			},
			expect: func(t *testing.T, r *Resource, err error) {
				assert.Nil(t, err)
				nav := r.NewNavigator()
				{
					_, err = nav.FocusName("emails")
					assert.Nil(t, err)
					{
						for i, f := range []func(v interface{}){
							func(v interface{}) { assert.Nil(t, v) },
							func(v interface{}) { assert.Equal(t, true, v) },
						} {
							_, err = nav.FocusIndex(i)
							assert.Nil(t, err)
							{
								_, err = nav.FocusName("primary")
								assert.Nil(t, err)
								f(nav.Current().Raw())
								nav.Retract()
							}
							nav.Retract()
						}
					}
					nav.Retract()
				}
			},
		},
		{
			name: "adding another member with primary turned on sets off the original primary",
			getResource: func(t *testing.T) *Resource {
				resource := NewResourceOf(resourceType, map[string]interface{}{
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
				})
				return resource
			},
			modification: func(t *testing.T, r *Resource) error {
				nav := r.NewNavigator()
				_, err := nav.FocusName("emails")
				require.Nil(t, err)
				return nav.Current().Add(map[string]interface{}{
					"value":   "random@bar.com",
					"type":    "work",
					"primary": true,
					"display": "random@bar.com",
				})
			},
			expect: func(t *testing.T, r *Resource, err error) {
				assert.Nil(t, err)

				nav := r.NewNavigator()
				{
					_, err = nav.FocusName("emails")
					assert.Nil(t, err)
					{
						for i, f := range []func(v interface{}){
							func(v interface{}) { assert.Nil(t, v) },
							func(v interface{}) { assert.Nil(t, v) },
							func(v interface{}) { assert.Equal(t, true, v) },
						} {
							_, err = nav.FocusIndex(i)
							assert.Nil(t, err)
							{
								_, err = nav.FocusName("primary")
								assert.Nil(t, err)
								f(nav.Current().Raw())
								nav.Retract()
							}
							nav.Retract()
						}
					}
					nav.Retract()
				}
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			r := test.getResource(t)
			err := test.modification(t, r)
			test.expect(t, r, err)
		})
	}
}

func (s *SubscriberTestSuite) mustResourceType(filePath string) *spec.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(spec.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *SubscriberTestSuite) mustSchema(filePath string) *spec.Schema {
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
