package stage

import (
	"context"
	"github.com/imulab/go-scim/core"
	uuid "github.com/satori/go.uuid"
	"strings"
)

const (
	orderId = 300
)

// Create a new ID filter. The filter is responsible of processing field id. It is intended to generate a new UUID
// for the resource to serve as its id, hence the annotation is usually marked on the id field. The filter replaces
// the id field value with a new UUID when Filter is called; it does nothing when FilterWithRef is called.
func NewIDFilter() PropertyFilter {
	return &idFilter{}
}

type idFilter struct{}

func (f *idFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Id == "id"
}

func (f *idFilter) Order() int {
	return orderId
}

func (f *idFilter) FilterOnCreate(ctx context.Context, resource *core.Resource, property core.Property) error {
	return property.(core.Crud).Replace(nil, strings.ToLower(uuid.NewV4().String()))
}

func (f *idFilter) FilterOnUpdate(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	return nil
}

var (
	_ PropertyFilter = (*idFilter)(nil)
)