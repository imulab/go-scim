package prop

import (
	"encoding/json"
	"github.com/imulab/go-scim/src/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

func TestExclusivePrimaryPropertySubscriber(t *testing.T) {
	s := new(ExclusivePrimaryTestSuite)
	s.resourceBase = "../../tests/exclusive_primary_test_suite"
	suite.Run(t, s, )
}

type ExclusivePrimaryTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *ExclusivePrimaryTestSuite) TestSubscriber() {
	_ = s.mustSchema("/user_schema.json")
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

				p, err := resource.NewNavigator().FocusName("emails")
				require.Nil(t, err)
				p.Subscribe(NewExclusivePrimaryPropertySubscriber())

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

				p, err := resource.NewNavigator().FocusName("emails")
				require.Nil(t, err)
				p.Subscribe(NewExclusivePrimaryPropertySubscriber())

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

				p, err := resource.NewNavigator().FocusName("emails")
				require.Nil(t, err)
				p.Subscribe(NewExclusivePrimaryPropertySubscriber())

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

func (s *ExclusivePrimaryTestSuite) mustResourceType(filePath string) *core.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(core.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *ExclusivePrimaryTestSuite) mustSchema(filePath string) *core.Schema {
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
