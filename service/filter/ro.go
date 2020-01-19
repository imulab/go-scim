package filter

import (
	"context"
	"github.com/elvsn/scim.go/annotation"
	"github.com/elvsn/scim.go/prop"
	"github.com/elvsn/scim.go/spec"
)

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

	if refNav == nil || refNav.Current() == outOfSync {
		return nil
	}

	if refNav.Current().IsUnassigned() {
		return nav.Delete().Error()
	} else {
		return nav.Replace(refNav.Current().Raw()).Error()
	}
}
