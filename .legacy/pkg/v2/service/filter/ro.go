package filter

import (
	"context"
	"github.com/imulab/go-scim/pkg/v2/annotation"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// ReadOnlyFilter returns a ByProperty filter that supports resetting and copying values for readOnly properties that
// was annotated with @ReadOnly. The annotation can specify two boolean parameters named "reset" and "copy". When "reset"
// is set to true, the filter will delete the property value; When "copy" is set to true and a reference is available,
// the filter will replace the property value with that of the reference. If any of these two parameters are not set, they
// are treated as false. The value changes in this filter generates additional event propagation.
func ReadOnlyFilter() ByProperty {
	return readOnlyPropertyFilter{}
}

type readOnlyPropertyFilter struct{}

func (f readOnlyPropertyFilter) Supports(attribute *spec.Attribute) bool {
	if _, ok := attribute.Annotation(annotation.ReadOnly); !ok {
		return false
	}
	return attribute.Mutability() == spec.MutabilityReadOnly
}

func (f readOnlyPropertyFilter) Filter(_ context.Context, _ *spec.ResourceType, nav prop.Navigator) error {
	if nav.HasError() {
		return nav.Error()
	}

	if err := f.tryReset(nav); err != nil {
		return err
	}

	return nil
}

func (f readOnlyPropertyFilter) FilterRef(_ context.Context, _ *spec.ResourceType, nav prop.Navigator, refNav prop.Navigator) error {
	if nav.HasError() {
		return nav.Error()
	}

	if err := f.tryReset(nav); err != nil {
		return err
	}

	if err := f.tryCopy(nav, refNav); err != nil {
		return err
	}

	return nil
}

func (f readOnlyPropertyFilter) tryReset(nav prop.Navigator) error {
	attr := nav.Current().Attribute()
	params, _ := attr.Annotation(annotation.ReadOnly)
	if wantReset, ok := params["reset"].(bool); !ok || !wantReset {
		return nil
	}

	return nav.Delete().Error()
}

func (f readOnlyPropertyFilter) tryCopy(nav prop.Navigator, refNav prop.Navigator) error {
	attr := nav.Current().Attribute()
	params, _ := attr.Annotation(annotation.ReadOnly)
	if wantCopy, ok := params["copy"].(bool); !ok || !wantCopy {
		return nil
	}

	if refNav == nil || IsOutOfSync(refNav.Current()) {
		return nil
	}

	if refNav.Current().IsUnassigned() {
		return nav.Delete().Error()
	} else {
		return nav.Replace(refNav.Current().Raw()).Error()
	}
}
