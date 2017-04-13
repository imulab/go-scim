package shared

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestValidateUniqueness(t *testing.T) {
	repo := &mockRepository{}

	sch, _, err := ParseSchema("../resources/tests/user_schema.json")
	require.Nil(t, err)
	require.NotNil(t, sch)

	for _, test := range []struct {
		getResource func(r *Resource) *Resource
		assertion   func(err error)
	}{
		{
			func(r *Resource) *Resource {
				return r
			},
			func(err error) {
				assert.Nil(t, err)
			},
		},
		{
			func(r *Resource) *Resource {
				r.Complex["id"] = "foo"
				return r
			},
			func(err error) {
				assert.NotNil(t, err)
				assert.IsType(t, &DuplicateError{}, err)
				assert.Equal(t, "id", err.(*DuplicateError).Path)
				assert.Equal(t, "foo", err.(*DuplicateError).Value)
			},
		},
	} {
		r, _, err := ParseResource("../resources/tests/user_1.json")
		require.Nil(t, err)
		require.NotNil(t, r)
		r = test.getResource(r)

		test.assertion(ValidateUniqueness(r, sch, repo))
	}
}

// A mock repository that mocks the Count(query string) method
// If the query contains "foo", returns 1, else
type mockRepository struct{}

func (r *mockRepository) Create(provider DataProvider) error                  { return nil }
func (r *mockRepository) Get(id string) (DataProvider, error)                 { return nil, nil }
func (r *mockRepository) GetAll() ([]Complex, error)                          { return nil, nil }
func (r *mockRepository) Update(provider DataProvider) error                  { return nil }
func (r *mockRepository) Delete(id string) error                              { return nil }
func (r *mockRepository) Search(payload SearchRequest) (*ListResponse, error) { return nil, nil }
func (r *mockRepository) Count(query string) (int, error) {
	if strings.Contains(query, "foo") {
		return 1, nil
	} else {
		return 0, nil
	}
}
