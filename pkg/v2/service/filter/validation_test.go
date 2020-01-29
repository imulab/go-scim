package filter

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/db"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func TestValidationFilter(t *testing.T) {
	getResourceType := func() *spec.ResourceType {
		var resourceType *spec.ResourceType
		{
			for _, each := range []struct {
				filepath  string
				structure interface{}
				post      func(parsed interface{})
			}{
				{
					filepath:  "../../../../public/schemas/core_schema.json",
					structure: new(spec.Schema),
					post: func(parsed interface{}) {
						spec.Schemas().Register(parsed.(*spec.Schema))
					},
				},
				{
					filepath:  "../../../../public/schemas/user_schema.json",
					structure: new(spec.Schema),
					post: func(parsed interface{}) {
						spec.Schemas().Register(parsed.(*spec.Schema))
					},
				},
				{
					filepath:  "../../../../public/resource_types/user_resource_type.json",
					structure: new(spec.ResourceType),
					post: func(parsed interface{}) {
						resourceType = parsed.(*spec.ResourceType)
					},
				},
			} {
				f, err := os.Open(each.filepath)
				require.Nil(t, err)
				raw, err := ioutil.ReadAll(f)
				require.Nil(t, err)
				err = json.Unmarshal(raw, each.structure)
				require.Nil(t, err)
				if each.post != nil {
					each.post(each.structure)
				}
			}
		}
		return resourceType
	}

	tests := []struct {
		name         string
		attrJson     string
		getProperty  func(t *testing.T, attr *spec.Attribute) prop.Navigator
		getReference func(t *testing.T, attr *spec.Attribute) prop.Navigator
		getDB        func() db.DB
		expect       func(t *testing.T, err error)
	}{
		{
			name: "unassigned property fails required check",
			attrJson: `
{
  "id": "userName",
  "name": "userName",
  "_path": "userName",
  "type": "string",
  "required": true
}
`,
			getProperty: func(t *testing.T, attr *spec.Attribute) prop.Navigator {
				return prop.Navigate(prop.NewProperty(attr))
			},
			getReference: func(t *testing.T, attr *spec.Attribute) prop.Navigator {
				return nil
			},
			getDB: func() db.DB {
				return nil
			},
			expect: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrInvalidValue, errors.Unwrap(err))
			},
		},
		{
			name: "out of scope value fails when canonical values enforced as Enum",
			attrJson: `
{
  "id": "type",
  "name": "type",
  "_path": "type",
  "type": "string",
  "canonicalValues": ["A", "B"],
  "_annotations": {
    "@Enum": {}
  }
}
`,
			getProperty: func(t *testing.T, attr *spec.Attribute) prop.Navigator {
				p := prop.NewProperty(attr)
				_, err := p.Replace("C")
				assert.Nil(t, err)
				return prop.Navigate(p)
			},
			getReference: func(t *testing.T, attr *spec.Attribute) prop.Navigator {
				return nil
			},
			getDB: func() db.DB { return nil },
			expect: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrInvalidValue, errors.Unwrap(err))
			},
		},
		{
			name: "in scope value passes when canonical values enforced as Enum",
			attrJson: `
{
  "id": "type",
  "name": "type",
  "_path": "type",
  "type": "string",
  "canonicalValues": ["A", "B"],
  "_annotations": {
    "@Enum": {}
  }
}
`,
			getProperty: func(t *testing.T, attr *spec.Attribute) prop.Navigator {
				p := prop.NewProperty(attr)
				_, err := p.Replace("A")
				assert.Nil(t, err)
				return prop.Navigate(p)
			},
			getReference: func(t *testing.T, attr *spec.Attribute) prop.Navigator {
				return nil
			},
			getDB: func() db.DB { return nil },
			expect: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
		{
			name: "immutable property fails check when value different with reference",
			attrJson: `
{
  "id": "field",
  "name": "field",
  "_path": "field",
  "type": "string",
  "mutability": "immutable"
}
`,
			getProperty: func(t *testing.T, attr *spec.Attribute) prop.Navigator {
				p := prop.NewProperty(attr)
				_, err := p.Replace("changed!!!")
				assert.Nil(t, err)
				return prop.Navigate(p)
			},
			getReference: func(t *testing.T, attr *spec.Attribute) prop.Navigator {
				p := prop.NewProperty(attr)
				_, err := p.Replace("foobar")
				assert.Nil(t, err)
				return prop.Navigate(p)
			},
			getDB: func() db.DB { return nil },
			expect: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrMutability, errors.Unwrap(err))
			},
		},
		{
			name:     "non-unique value fails check",
			attrJson: `{}`,
			getProperty: func(t *testing.T, _ *spec.Attribute) prop.Navigator {
				resourceType := getResourceType()
				nav := prop.NewResource(resourceType).Navigator()
				assert.False(t, nav.Replace(map[string]interface{}{
					"id":       "return_1_please",
					"userName": "foobar",
				}).HasError())

				return nav.Dot("userName")
			},
			getReference: func(t *testing.T, attr *spec.Attribute) prop.Navigator {
				return nil
			},
			getDB: func() db.DB {
				return &uniquenessTestMockDatabase{}
			},
			expect: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Equal(t, spec.ErrInvalidValue, errors.Unwrap(err))
			},
		},
		{
			name:     "unique value passes check",
			attrJson: `{}`,
			getProperty: func(t *testing.T, _ *spec.Attribute) prop.Navigator {
				resourceType := getResourceType()
				nav := prop.NewResource(resourceType).Navigator()
				assert.False(t, nav.Replace(map[string]interface{}{
					"id":       "return_0_please",
					"userName": "foobar",
				}).HasError())

				return nav.Dot("userName")
			},
			getReference: func(t *testing.T, attr *spec.Attribute) prop.Navigator {
				return nil
			},
			getDB: func() db.DB {
				return &uniquenessTestMockDatabase{}
			},
			expect: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attr := new(spec.Attribute)
			assert.Nil(t, json.Unmarshal([]byte(test.attrJson), attr))

			filter := ValidationFilter(test.getDB())
			property := test.getProperty(t, attr)
			reference := test.getReference(t, attr)

			var err error
			if reference == nil {
				err = filter.Filter(context.Background(), nil, property)
			} else {
				err = filter.FilterRef(context.Background(), nil, property, reference)
			}

			test.expect(t, err)
		})
	}
}

type uniquenessTestMockDatabase struct {
	mock.Mock
}

func (d *uniquenessTestMockDatabase) Count(_ context.Context, filter string) (int, error) {
	f, err := expr.CompileFilter(filter)
	if err != nil {
		return 0, err
	}

	// mock logic: check on the id
	switch f.Left().Right().Token() {
	case `"return_0_please"`:
		return 0, nil
	case `"return_1_please"`:
		return 1, nil
	default:
		return 1, nil
	}
}

func (d *uniquenessTestMockDatabase) Insert(_ context.Context, _ *prop.Resource) error {
	return nil
}

func (d *uniquenessTestMockDatabase) Get(_ context.Context, _ string, _ *crud.Projection) (*prop.Resource, error) {
	return nil, nil
}

func (d *uniquenessTestMockDatabase) Replace(_ context.Context, _ *prop.Resource, _ *prop.Resource) error {
	return nil
}

func (d *uniquenessTestMockDatabase) Delete(_ context.Context, _ *prop.Resource) error {
	return nil
}

func (d *uniquenessTestMockDatabase) Query(_ context.Context, _ string, _ *crud.Sort, _ *crud.Pagination, _ *crud.Projection) ([]*prop.Resource, error) {
	return []*prop.Resource{}, nil
}
