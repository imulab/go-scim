package filter

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/core/annotations"
	"github.com/imulab/go-scim/core/errors"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/core/spec"
	"github.com/imulab/go-scim/protocol/db"
	"strings"
)

// Create a ForResource filter that performs resource validation. Attributes checks such as required, canonicalValues,
// uniqueness and mutability is performed, where applicable.
func Validation(database db.DB) ForResource {
	return FromForProperty(&validationFilter{
		database: database,
	})
}

type validationFilter struct {
	database db.DB
}

func (f *validationFilter) Supports(attribute *spec.Attribute) bool {
	return true
}

func (f *validationFilter) Filter(ctx context.Context, resource *prop.Resource, property prop.Property) error {
	if err := f.validateRequired(property); err != nil {
		return err
	}
	if err := f.validateCanonical(property); err != nil {
		return err
	}
	if err := f.validateUniqueness(ctx, resource, property); err != nil {
		return err
	}
	return nil
}

func (f *validationFilter) FieldRef(ctx context.Context, resource *prop.Resource, property prop.Property,
	refResource *prop.Resource, refProperty prop.Property) error {
	if err := f.validateRequired(property); err != nil {
		return err
	}
	if err := f.validateCanonical(property); err != nil {
		return err
	}
	if err := f.validateMutability(property, refProperty); err != nil {
		return err
	}
	if err := f.validateUniqueness(ctx, resource, property); err != nil {
		return err
	}
	return nil
}

func (f *validationFilter) validateRequired(property prop.Property) error {
	if property.Attribute().Required() && property.IsUnassigned() {
		return errors.InvalidValue("'%s' is required, but is unassigned", property.Attribute().Path())
	}
	return nil
}

func (f *validationFilter) validateCanonical(property prop.Property) error {
	if property.IsUnassigned() {
		return nil
	}

	var attr = property.Attribute()
	if attr.CountCanonicalValues() == 0 || attr.HasAnnotation(annotations.RelaxCanonical) {
		return nil
	}

	if ok := attr.HasCanonicalValue(func(value string) bool {
		pv := property.Raw().(string)
		if attr.CaseExact() {
			return value == pv
		} else {
			return strings.ToLower(value) == strings.ToLower(pv)
		}
	}); !ok {
		return errors.InvalidValue("'%s' does not conform to its canonical values", attr.Path())
	}

	return nil
}

func (f *validationFilter) validateMutability(property prop.Property, refProp prop.Property) error {
	if refProp == nil {
		return nil
	}

	switch property.Attribute().Mutability() {
	case spec.MutabilityReadOnly:
	// read only mutability is not asserted here because it is taken care
	// of by previous filters (cleared and copied), and server has right to
	// modify it as it sees fit
	case spec.MutabilityImmutable:
		if !refProp.IsUnassigned() && !property.Matches(refProp) {
			return errors.Mutability("'%s' is immutable, but value has changed", property.Attribute().Path())
		}
	}

	return nil
}

func (f *validationFilter) validateUniqueness(ctx context.Context, resource *prop.Resource, property prop.Property) error {
	if property.IsUnassigned() {
		return nil
	}

	attr := property.Attribute()
	if attr.Uniqueness() == spec.UniquenessNone || attr.Uniqueness() == spec.UniquenessGlobal {
		return nil
	}

	id := resource.ID()
	if len(id) == 0 {
		return errors.Internal("resource has no id")
	}

	// We may run into problem where the uniqueness=server attribute is 'id' itself. However, as of
	// now, 'id' is defined as uniqueness=global by assigning a UUID to it.
	filter := fmt.Sprintf("(id ne \"%s\") and (%s eq \"%s\")", id, attr.Path(), property.Raw())
	n, err := f.database.Count(ctx, filter)
	if err != nil {
		return err
	} else if n > 0 {
		return errors.Uniqueness("'%s' violated uniqueness constraint", attr.Path())
	}

	return nil
}
