package filters

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/pkg/core"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/protocol"
)

func NewValidationResourceFilter(persistence protocol.PersistenceProvider) protocol.ResourceFilter {
	return NewResourceFieldFilterOf(NewValidationFieldFilter(persistence))
}

// Create a validation filter that validates the property value according to the defined attributes. This filter
// currently validates required, canonicalValues, mutability and uniqueness.
func NewValidationFieldFilter(persistence protocol.PersistenceProvider) protocol.FieldFilter {
	return &validationFilter{
		persistence: persistence,
	}
}

type validationFilter struct {
	persistence protocol.PersistenceProvider
}

func (f *validationFilter) Supports(attribute *core.Attribute) bool {
	return true
}

func (f *validationFilter) Filter(ctx *protocol.FilterContext, resource *prop.Resource, property core.Property) error {
	if err := f.validateRequired(property); err != nil {
		return err
	}

	if err := f.validateCanonical(property); err != nil {
		return err
	}

	if err := f.validateUniqueness(ctx.RequestContext(), resource, property); err != nil {
		return err
	}

	return nil
}

func (f *validationFilter) FieldRef(ctx *protocol.FilterContext, resource *prop.Resource, property core.Property, refResource *prop.Resource, refProperty core.Property) error {
	if err := f.validateRequired(property); err != nil {
		return err
	}

	if err := f.validateCanonical(property); err != nil {
		return err
	}

	if err := f.validateMutability(property, refProperty); err != nil {
		return err
	}

	if err := f.validateUniqueness(ctx.RequestContext(), resource, property); err != nil {
		return err
	}

	return nil
}

func (f *validationFilter) validateRequired(property core.Property) error {
	if property.Attribute().Required() && property.IsUnassigned() {
		return errors.InvalidValue("'%s' is required, but is unassigned", property.Attribute().Path())
	}
	return nil
}

func (f *validationFilter) validateCanonical(property core.Property) error {
	if property.IsUnassigned() {
		return nil
	}

	var attr = property.Attribute()

	if attr.CountCanonicalValues() == 0 || attr.HasAnnotation("@relaxCanonical") {
		return nil
	}

	match := false
	attr.ForEachCanonicalValue(func(canonicalValue string) {
		if !match && canonicalValue == property.Raw().(string) {
			match = true
		}
	})
	if !match {
		return errors.InvalidValue("'%s' does not conform to its canonical values", attr.Path())
	}

	return nil
}

func (f *validationFilter) validateMutability(property core.Property, refProp core.Property) error {
	if refProp == nil {
		return nil
	}

	switch property.Attribute().Mutability() {
	case core.MutabilityReadOnly:
	// read only mutability is not asserted here because it is taken care
	// of by previous filters (cleared and copied), and server has right to
	// modify it as it sees fit
	case core.MutabilityImmutable:
		if !refProp.IsUnassigned() && !property.Matches(refProp) {
			return errors.Mutability("'%s' is immutable, but value has changed", property.Attribute().Path())
		}
	}

	return nil
}

func (f *validationFilter) validateUniqueness(ctx context.Context, resource *prop.Resource, property core.Property) error {
	if property.IsUnassigned() {
		return nil
	}

	attr := property.Attribute()
	if attr.Uniqueness() == core.UniquenessNone || attr.Uniqueness() == core.UniquenessGlobal {
		return nil
	}

	id := resource.ID()
	if len(id) == 0 {
		return errors.Internal("resource has no id")
	}

	// We may run into problem where the uniqueness=server attribute is 'id' itself. However, as of
	// now, 'id' is defined as uniqueness=global by assigning a UUID to it.
	filter := fmt.Sprintf("(id ne \"%s\") and (%s eq \"%s\")", id, attr.Path(), property.Raw())
	n, err := f.persistence.Count(ctx, filter)
	if err != nil {
		return err
	} else if n > 0 {
		return errors.Uniqueness("'%s' violated uniqueness constraint", attr.Path())
	}

	return nil
}
