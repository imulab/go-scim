package shared

import (
	"context"
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

		ctx := context.Background()
		test.assertion(ValidateUniqueness(r, sch, repo, ctx))
	}
}

// A mock repository that mocks the Count(query string) method
// If the query contains "foo", returns 1, else
type mockRepository struct{}

func (r *mockRepository) Create(provider DataProvider, ctx context.Context) error { return nil }
func (r *mockRepository) Get(id, version string, ctx context.Context) (DataProvider, error) { return nil, nil }
func (r *mockRepository) GetAll(context.Context) ([]Complex, error)               { return nil, nil }
func (r *mockRepository) Update(id, version string, provider DataProvider, ctx context.Context) error { return nil }
func (r *mockRepository) Delete(id, version string, ctx context.Context) error    { return nil }
func (r *mockRepository) Search(payload SearchRequest, ctx context.Context) (*ListResponse, error) { return nil, nil }
func (r *mockRepository) Count(query string, ctx context.Context) (int, error) {
	if strings.Contains(query, "foo") {
		return 1, nil
	} else {
		return 0, nil
	}
}
