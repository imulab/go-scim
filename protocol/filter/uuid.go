package filter

import (
	"context"
	"github.com/imulab/go-scim/core"
	"github.com/satori/go.uuid"
	"strings"
)

// A property filter that generates and assigns a new uuid V4 as value to the property, if the property
// is marked with annotation '@uuid'. The generation only happens when Filter is called. When FilterWithRef
// is called, the filter does nothing.
type uuidFilter struct {}

func (f *uuidFilter) Supports(attribute *core.Attribute) bool {
	return ContainsAnnotation(attribute, "@uuid")
}

func (f *uuidFilter) Order(attribute *core.Attribute) int {
	// Expected to happen after validation phase.
	return 200
}

func (f *uuidFilter) Filter(ctx context.Context, _ *core.Resource, property core.Property) error {
	return property.(core.Crud).Replace(nil, strings.ToLower(uuid.NewV4().String()))
}

func (f *uuidFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, _ *core.Resource, _ core.Property) error {
	return nil
}

// Create a new UUID filter
func NewUUIDFilter() PropertyFilter {
	return &uuidFilter{}
}