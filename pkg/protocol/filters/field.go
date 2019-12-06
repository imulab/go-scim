package filters

import (
	"github.com/imulab/go-scim/pkg/core"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/protocol"
)

func NewResourceFieldFilterOf(filters ...protocol.FieldFilter) protocol.ResourceFilter {
	return &resourceFieldFilter{
		filters: filters,
	}
}

type (
	resourceFieldFilter struct {
		filters []protocol.FieldFilter
	}
	// Visitor implementation to traverse the resource structure. If reference is present,
	// best effort to keep the reference property in sync with the traversing property so
	// they can be compared.
	resourceFieldVisitor struct {
		ctx      *protocol.FilterContext
		resource *prop.Resource
		ref      *prop.Resource
		refNav   *prop.Navigator
		filters  []protocol.FieldFilter
		stack    []*frame
	}
	// Stack frame to keep track during the resource traversal.
	frame struct {
		// Container properties for all properties being visited in the current frame
		container core.Container
		// true, if the currently focused property on refNav corresponds to the above container
		inSync bool
		// function to invoke when the current frame is about to be retired
		exitFunc func()
	}
)

func (f *resourceFieldFilter) Filter(ctx *protocol.FilterContext, resource *prop.Resource) error {
	v := &resourceFieldVisitor{
		ctx:      ctx,
		resource: resource,
		filters:  f.filters,
		stack:    make([]*frame, 0),
	}
	return resource.Visit(v)
}

func (f *resourceFieldFilter) FilterRef(ctx *protocol.FilterContext, resource *prop.Resource, ref *prop.Resource) error {
	if ref == nil {
		return f.Filter(ctx, resource)
	}

	v := &resourceFieldVisitor{
		ctx:      ctx,
		resource: resource,
		ref:      ref,
		refNav:   ref.NewNavigator(),
		filters:  f.filters,
		stack:    make([]*frame, 0),
	}
	return resource.Visit(v)
}

func (v *resourceFieldVisitor) ShouldVisit(property core.Property) bool {
	return true
}

func (v *resourceFieldVisitor) Visit(property core.Property) error {
	if v.ref == nil {
		return v.visit(property)
	}

	var refProp core.Property = nil
	{
		if v.currentFrame().inSync {
			container := v.refNav.Current().(core.Container)
			switch {
			case container.Attribute().MultiValued():
				if r, err := v.refNav.FocusCriteria(func(child core.Property) bool {
					return child.Matches(property)
				}); err == nil {
					v.refNav.Retract() // we don't want to actually focus, just get the child property
					refProp = r
				}
			case container.Attribute().Type() == core.TypeComplex:
				refProp = container.ChildAtIndex(property.Attribute().Name())
			default:
				panic("invalid state")
			}
		}
	}
	return v.visitWithRef(property, refProp)
}

func (v *resourceFieldVisitor) visitWithRef(property core.Property, refProp core.Property) error {
	for _, filter := range v.filters {
		if filter.Supports(property.Attribute()) {
			if err := filter.FieldRef(v.ctx, v.resource, property, v.ref, refProp); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *resourceFieldVisitor) visit(property core.Property) error {
	for _, filter := range v.filters {
		if filter.Supports(property.Attribute()) {
			if err := filter.Filter(v.ctx, v.resource, property); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *resourceFieldVisitor) BeginChildren(container core.Container) {
	switch {
	case container.Attribute().MultiValued():
		v.beginMulti(container)
	case container.Attribute().Type() == core.TypeComplex:
		v.beginComplex(container)
	}
}

func (v *resourceFieldVisitor) beginComplex(newContainer core.Container) {
	// Short circuit: no reference to worry about
	if v.refNav == nil {
		v.push(newContainer)
		return
	}

	// Short circuit: We are on the top level.
	// At the very start, BeginContainer is invoked on the base complex property by the visitor.
	// And reference navigator, if exists, is already focused on the base property. Hence, we
	// immediately assumes they are in sync.
	if len(v.stack) == 0 {
		v.push(newContainer)
		v.currentFrame().inSync = true
		v.currentFrame().exitFunc = func() {
			v.refNav.Retract()
		}
		return
	}

	// Another short circuit: no need to sync up if the current frame is not in sync
	if !v.currentFrame().inSync {
		v.push(newContainer)
		return
	}

	// We are in sync:
	// v.currentFrame().container matches v.refNav.Current()
	var focusErr error
	{
		oldContainerAttr := v.currentFrame().container.Attribute()
		switch {
		case oldContainerAttr.MultiValued():
			_, focusErr = v.refNav.FocusCriteria(func(child core.Property) bool {
				return newContainer.Matches(child)
			})
		case oldContainerAttr.Type() == core.TypeComplex:
			_, focusErr = v.refNav.FocusName(newContainer.Attribute().Name())
		default:
			panic("invalid frame")
		}
	}
	v.push(newContainer)
	if focusErr == nil {
		v.currentFrame().inSync = true
		v.currentFrame().exitFunc = func() {
			v.refNav.Retract()
		}
	}
}

func (v *resourceFieldVisitor) beginMulti(newContainer core.Container) {
	// Short circuit: no reference to worry about
	if v.refNav == nil {
		v.push(newContainer)
		return
	}

	// Short circuit: the current frame is not in sync
	if !v.currentFrame().inSync {
		v.push(newContainer)
		return
	}

	// After this point, we are sure we have a reference and it is in sync.
	// Because multiValued property can only appear as a sub property of a complex property,
	// we select from the refNav directly by name
	v.push(newContainer)
	if _, err := v.refNav.FocusName(newContainer.Attribute().Name()); err == nil {
		v.currentFrame().inSync = true
		v.currentFrame().exitFunc = func() {
			v.refNav.Retract()
		}
	}
}

func (v *resourceFieldVisitor) EndChildren(container core.Container) {
	if v.currentFrame().exitFunc != nil {
		v.currentFrame().exitFunc()
	}
	v.pop()
}

func (v *resourceFieldVisitor) push(container core.Container) {
	v.stack = append(v.stack, &frame{
		container: container,
	})
}

func (v *resourceFieldVisitor) pop() {
	if len(v.stack) > 0 {
		v.stack = v.stack[:len(v.stack)-1]
	}
}

func (v *resourceFieldVisitor) currentFrame() *frame {
	if len(v.stack) == 0 {
		panic("stack is empty")
	}
	return v.stack[len(v.stack)-1]
}
