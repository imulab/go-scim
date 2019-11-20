package filter

import (
	"context"
	"github.com/imulab/go-scim/core"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewUUIDFilter(t *testing.T) {
	filter := NewUUIDFilter()

	tests := []struct {
		name        string
		getProperty func() core.Property
		assert      func(t *testing.T, prop core.Property, err error)
	}{
		{
			name: "default",
			getProperty: func() core.Property {
				return core.Properties.NewStringOf(&core.Attribute{
					Name: "id",
					Type: core.TypeString,
				}, "")
			},
			assert: func(t *testing.T, prop core.Property, err error) {
				assert.Nil(t, err)
				assert.NotEmpty(t, prop.Raw())
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			prop := test.getProperty()
			err := filter.Filter(context.Background(), nil, prop, nil)
			test.assert(t, prop, err)
		})
	}
}
