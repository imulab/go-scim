package filter

import (
	"context"
	"github.com/imulab/go-scim/core"
	"github.com/satori/go.uuid"
	"strings"
)

// A property filter that generates and assigns a new uuid V4 as value to the property, if the property
// is marked with annotation '@uuid'. The generation only happens when there's no reference resource.
type uuidFilter struct {}

func (f *uuidFilter) Supports(attribute *core.Attribute) bool {
	return ContainsAnnotation(attribute, "@uuid")
}

func (f *uuidFilter) Order(attribute *core.Attribute) int {
	// Expected to happen after validation phase.
	return 200
}

func (f *uuidFilter) Filter(ctx context.Context, _ *core.Resource, property core.Property, ref *core.Resource) error {
	if ref != nil {
		return nil
	}
	return property.(core.Crud).Replace(nil, strings.ToLower(uuid.NewV4().String()))
}

// Create a new UUID filter
func NewUUIDFilter() PropertyFilter {
	return &uuidFilter{}
}