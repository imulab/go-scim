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

func TestPrimaryMonitor(t *testing.T) {
	s := new(PrimaryMonitorTestSuite)
	s.resourceBase = "../../tests/primary_monitor_test_suite"
	suite.Run(t, s)
}

type PrimaryMonitorTestSuite struct {
	suite.Suite
	resourceBase string
}

func (s *PrimaryMonitorTestSuite) TestInitialScan() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name        string
		getResource func() *Resource
		expect      func(t *testing.T, err error)
	}{
		{
			name: "resource conforming to primary rule scans no error",
			getResource: func() *Resource {
				return NewResourceOf(resourceType, map[string]interface{}{
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
					"phoneNumbers": []interface{}{
						map[string]interface{}{
							"value":   "123-45678",
							"type":    "work",
							"primary": true,
							"display": "123-45678",
						},
						map[string]interface{}{
							"value":   "123-45679",
							"type":    "work",
							"display": "123-45679",
						},
					},
				})
			},
			expect: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "resource violating primary rule scans error",
			getResource: func() *Resource {
				return NewResourceOf(resourceType, map[string]interface{}{
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
							"primary": true,
							"display": "imulab@bar.com",
						},
					},
					"phoneNumbers": []interface{}{
						map[string]interface{}{
							"value":   "123-45678",
							"type":    "work",
							"primary": true,
							"display": "123-45678",
						},
						map[string]interface{}{
							"value":   "123-45679",
							"type":    "work",
							"display": "123-45679",
						},
					},
				})
			},
			expect: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			monitor := NewPrimaryMonitor(test.getResource())
			test.expect(t, monitor.Scan())
		})
	}
}

func (s *PrimaryMonitorTestSuite) TestScanAfterModification() {
	_ = s.mustSchema("/user_schema.json")
	resourceType := s.mustResourceType("/user_resource_type.json")

	tests := []struct {
		name         string
		getResource  func() *Resource
		modification func(t *testing.T, r *Resource)
		expect       func(t *testing.T, r *Resource, err error)
	}{
		{
			name: "turning off primary is fine",
			getResource: func() *Resource {
				return NewResourceOf(resourceType, map[string]interface{}{
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
			},
			modification: func(t *testing.T, r *Resource) {
				nav := r.NewNavigator()
				_, _ = nav.FocusName("emails")
				{
					_, _ = nav.FocusIndex(0)
					{
						_, _ = nav.FocusName("primary")
						_, _ = nav.Current().Delete()
					}
				}
			},
			expect: func(t *testing.T, r *Resource, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "setting new primary while turning off old primary is fine",
			getResource: func() *Resource {
				return NewResourceOf(resourceType, map[string]interface{}{
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
			},
			modification: func(t *testing.T, r *Resource) {
				nav := r.NewNavigator()
				_, _ = nav.FocusName("emails")
				{
					_, _ = nav.FocusIndex(0)
					{
						_, _ = nav.FocusName("primary")
						_, _ = nav.Current().Delete()
						nav.Retract()
					}
					nav.Retract()

					_, _ = nav.FocusIndex(1)
					{
						_, _ = nav.FocusName("primary")
						_, _ = nav.Current().Replace(true)
						nav.Retract()
					}
					nav.Retract()
				}
			},
			expect: func(t *testing.T, r *Resource, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "setting new primary while leaving old primary turns off the old primary",
			getResource: func() *Resource {
				return NewResourceOf(resourceType, map[string]interface{}{
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
			},
			modification: func(t *testing.T, r *Resource) {
				nav := r.NewNavigator()
				_, _ = nav.FocusName("emails")
				{
					_, _ = nav.FocusIndex(1)
					{
						_, _ = nav.FocusName("primary")
						_, _ = nav.Current().Replace(true)
						nav.Retract()
					}
					nav.Retract()
				}
			},
			expect: func(t *testing.T, r *Resource, err error) {
				assert.Nil(t, err)
				nav := r.NewNavigator()
				_, _ = nav.FocusName("emails")
				{
					_, _ = nav.FocusIndex(0)
					{
						_, _ = nav.FocusName("primary")
						assert.Nil(t, nav.Current().Raw())
						nav.Retract()
					}
					nav.Retract()

					_, _ = nav.FocusIndex(1)
					{
						_, _ = nav.FocusName("primary")
						assert.Equal(t, true, nav.Current().Raw())
						nav.Retract()
					}
					nav.Retract()
				}
			},
		},
		{
			name: "adding two new primary is invalid",
			getResource: func() *Resource {
				return NewResourceOf(resourceType, map[string]interface{}{
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
			},
			modification: func(t *testing.T, r *Resource) {
				nav := r.NewNavigator()
				_, _ = nav.FocusName("emails")
				{
					_, err := nav.Current().Add(map[string]interface{}{
						"value":   "one@foo.com",
						"type":    "work",
						"primary": true,
						"display": "one@foo.com",
					})
					require.Nil(t, err)

					_, err = nav.Current().Add(map[string]interface{}{
						"value":   "two@foo.com",
						"type":    "work",
						"primary": true,
						"display": "one@foo.com",
					})
				}
			},
			expect: func(t *testing.T, r *Resource, err error) {
				nav := r.NewNavigator()
				_, _ = nav.FocusName("emails")
				assert.Equal(t, 4, nav.Current().(core.Container).CountChildren())
				assert.NotNil(t, err)
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			resource := test.getResource()
			monitor := NewPrimaryMonitor(resource)
			require.Nil(t, monitor.Scan())
			test.modification(t, resource)
			test.expect(t, resource, monitor.Scan())
		})
	}
}

func (s *PrimaryMonitorTestSuite) mustResourceType(filePath string) *core.ResourceType {
	f, err := os.Open(s.resourceBase + filePath)
	s.Require().Nil(err)

	raw, err := ioutil.ReadAll(f)
	s.Require().Nil(err)

	rt := new(core.ResourceType)
	err = json.Unmarshal(raw, rt)
	s.Require().Nil(err)

	return rt
}

func (s *PrimaryMonitorTestSuite) mustSchema(filePath string) *core.Schema {
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
