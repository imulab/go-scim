package stage

import (
	"context"
	"github.com/imulab/go-scim/pkg/core"
	"golang.org/x/crypto/bcrypt"
)

const (
	annotationBcrypt = "@bcrypt"
	orderBcrypt = 350
)

// Create a bcrypt filter. This filter is responsible for transforming singleValued string properties whose attribute
// is marked with annotation '@bcrypt'. The standard behaviour is hashing the non-unassigned value of the property and
// replaces the original value. However, when FilterWithUpdate is called, the standard behaviour will be skipped if
// the reference matches the property, indicating the value has not changed.
func NewBCryptFilter(cost int) PropertyFilter {
	return &bcryptFilter{
		cost: cost,
	}
}

var (
	_ PropertyFilter = (*bcryptFilter)(nil)
)

type bcryptFilter struct {
	cost int
}

func (f *bcryptFilter) Supports(attribute *core.Attribute) bool {
	return !attribute.MultiValued &&
		attribute.Type == core.TypeString &&
		containsAnnotation(attribute, annotationBcrypt)
}

func (f *bcryptFilter) Order() int {
	return orderBcrypt
}

func (f *bcryptFilter) FilterOnCreate(ctx context.Context, resource *core.Resource, property core.Property) error {
	return f.bcrypt(property)
}

func (f *bcryptFilter) FilterOnUpdate(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	if refProp != nil && property.(core.EqualAware).Matches(refProp) {
		return nil
	}
	return f.bcrypt(property)
}

func (f *bcryptFilter) bcrypt(property core.Property) error {
	if property.IsUnassigned() {
		return nil
	}
	original := property.Raw().(string)
	hashed, err := bcrypt.GenerateFromPassword([]byte(original), f.cost)
	if err != nil {
		return core.Errors.Internal("failed to process field with annotation @bcrypt: '%s'", property.Attribute().DisplayName())
	}
	return property.(core.Crud).Replace(nil, string(hashed))
}