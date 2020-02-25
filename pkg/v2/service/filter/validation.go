package filter

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/annotation"
	"github.com/imulab/go-scim/pkg/v2/db"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"strconv"
	"strings"
)

// ValidationFilter returns a ByProperty that performs validation on each property. The validation carried out are
// required check, canonical check, mutability check and uniqueness check.
//
// The required check fails when attribute is required but property is unassigned.
//
// The canonical check fails when @Enum is annotated with the attribute, indicating that the canonicalValues
// defined should be treated as the only valid values of holding property, and the property value is not among
// the canonicalValues.
//
// The mutability check only fails when attribute is immutable, and the property value differs from the reference
// property value, if one exists. It does not check for readOnly attributes because the logic is largely handled
// by ReadOnlyFilter.
//
// The uniqueness check fails when the property value already exists in the database. It formulates the query
// (id ne <id>) and (<path> eq <value>), where <id> is the resource id, <path> is the unique attribute path, and
// <value> is the property value. The database returns the number of records matching this filter. If the count is
// greater than 0, the check fails. Note this check only handles the uniqueness=server case.
//
// Error is returned to caller if any of these check fails.
func ValidationFilter(database db.DB) ByProperty {
	return &validationPropertyFilter{database: database}
}

type validationPropertyFilter struct {
	database db.DB
}

func (f *validationPropertyFilter) Supports(_ *spec.Attribute) bool {
	return true
}

func (f *validationPropertyFilter) Filter(ctx context.Context, _ *spec.ResourceType, nav prop.Navigator) error {
	if nav.HasError() {
		return nav.Error()
	}

	property := nav.Current()
	if err := f.validateRequired(property); err != nil {
		return err
	}
	if err := f.validateCanonical(property); err != nil {
		return err
	}
	if err := f.validateUniqueness(ctx, nav); err != nil {
		return err
	}

	return nil
}

func (f *validationPropertyFilter) FilterRef(ctx context.Context, _ *spec.ResourceType, nav prop.Navigator, refNav prop.Navigator) error {
	if nav.HasError() {
		return nav.Error()
	}

	if err := f.validateRequired(nav.Current()); err != nil {
		return err
	}
	if err := f.validateCanonical(nav.Current()); err != nil {
		return err
	}
	if err := f.validateMutability(nav.Current(), refNav.Current()); err != nil {
		return err
	}
	if err := f.validateUniqueness(ctx, nav); err != nil {
		return err
	}

	return nil
}

func (f *validationPropertyFilter) validateRequired(property prop.Property) error {
	if !property.Attribute().Required() || !property.IsUnassigned() {
		return nil
	}
	return fmt.Errorf("%w: '%s' is required", spec.ErrInvalidValue, property.Attribute().Path())
}

func (f *validationPropertyFilter) validateCanonical(property prop.Property) error {
	if property.Attribute().CountCanonicalValues() == 0 {
		return nil
	}

	if property.IsUnassigned() {
		return nil
	}

	if _, ok := property.Attribute().Annotation(annotation.Enum); !ok {
		return nil
	}

	v, ok := property.Raw().(string)
	if !ok {
		return nil
	}

	if ok := property.Attribute().ExistsCanonicalValue(func(canonicalValue string) bool {
		if property.Attribute().CaseExact() {
			return v == canonicalValue
		} else {
			return strings.ToLower(v) == strings.ToLower(canonicalValue)
		}
	}); !ok {
		return fmt.Errorf("%w: value of '%s' does not conform to canonicalValues", spec.ErrInvalidValue, property.Attribute().Path())
	}

	return nil
}

func (f *validationPropertyFilter) validateMutability(property prop.Property, ref prop.Property) error {
	if ref == nil || IsOutOfSync(ref) {
		return nil
	}

	switch property.Attribute().Mutability() {
	case spec.MutabilityReadOnly:
	// read only mutability is not asserted here because it is taken care
	// of by previous filters (reset and copied), and server has right to
	// modify it as it sees fit
	case spec.MutabilityImmutable:
		if !ref.IsUnassigned() && !property.Matches(ref) {
			return fmt.Errorf("%w: '%s' is immutable", spec.ErrMutability, property.Attribute().Path())
		}
	}

	return nil
}

func (f *validationPropertyFilter) validateUniqueness(ctx context.Context, nav prop.Navigator) error {
	property := nav.Current()
	switch property.Attribute().Uniqueness() {
	case spec.UniquenessNone, spec.UniquenessGlobal:
		return nil
	}

	if property.IsUnassigned() {
		return nil
	}

	var id string
	{
		idProperty, err := nav.Source().ChildAtIndex("id")
		if err != nil || idProperty.Attribute().ID() != "id" || idProperty.IsUnassigned() {
			return fmt.Errorf("%w: no id", spec.ErrInternal)
		}
		id = idProperty.Raw().(string)
	}

	// We may run into problem where the uniqueness=server attribute is 'id' itself. However, as of
	// now, 'id' is defined as uniqueness=global by assigning a UUID to it.
	filter := fmt.Sprintf("(id ne %s) and (%s eq %s)",
		strconv.Quote(id),
		property.Attribute().Path(),
		strconv.Quote(fmt.Sprintf("%v", property.Raw())),
	)
	n, err := f.database.Count(ctx, filter)
	if err != nil {
		return err
	} else if n > 0 {
		return fmt.Errorf("%w: value of '%s' is not unique", spec.ErrInvalidValue, property.Attribute().Path())
	}

	return nil
}
