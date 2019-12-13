package filter

import (
	"context"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
)

type (
	// Visitor implementation to traverse the resource structure. If reference is present,
	// best effort to keep the reference property in sync with the traversing property so
	// they can be compared.
	syncPropertyVisitor struct {
		ctx      context.Context
		resource *prop.Resource
		ref      *prop.Resource
		refNav   *prop.Navigator
		filters  []ForProperty
		stack    []*frame
	}
	// Stack frame to keep track during the resource traversal.
	frame struct {
		// Container properties for all properties being visited in the current frame
		container prop.Container
		// true, if the currently focused property on refNav corresponds to the above container
		inSync bool
		// function to invoke when the current frame is about to be retired
		exitFunc func()
	}
)

func (v *syncPropertyVisitor) ShouldVisit(property prop.Property) bool {
	return true
}

func (v *syncPropertyVisitor) Visit(property prop.Property) error {
	if v.ref == nil {
		return v.visit(property)
	}

	var refProp prop.Property
	{
		if v.currentFrame().inSync {
			container := v.refNav.Current().(prop.Container)
			switch {
			case container.Attribute().MultiValued():
				if r, err := v.refNav.FocusCriteria(func(child prop.Property) bool {
					return child.Matches(property)
				}); err == nil {
					v.refNav.Retract() // we don't want to actually focus, just get the child property
					refProp = r
				}
			case container.Attribute().Type() == spec.TypeComplex:
				refProp = container.ChildAtIndex(property.Attribute().Name())
			default:
				panic("invalid state")
			}
		}
	}
	return v.visitWithRef(property, refProp)
}

func (v *syncPropertyVisitor) visitWithRef(property prop.Property, refProp prop.Property) error {
	for _, filter := range v.filters {
		if filter.Supports(property.Attribute()) {
			if err := filter.FieldRef(v.ctx, v.resource, property, v.ref, refProp); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *syncPropertyVisitor) visit(property prop.Property) error {
	for _, filter := range v.filters {
		if filter.Supports(property.Attribute()) {
			if err := filter.Filter(v.ctx, v.resource, property); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *syncPropertyVisitor) BeginChildren(container prop.Container) {
	switch {
	case container.Attribute().MultiValued():
		v.beginMulti(container)
	case container.Attribute().Type() == spec.TypeComplex:
		v.beginComplex(container)
	}
}

func (v *syncPropertyVisitor) beginComplex(newContainer prop.Container) {
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
			_, focusErr = v.refNav.FocusCriteria(func(child prop.Property) bool {
				return newContainer.Matches(child)
			})
		case oldContainerAttr.Type() == spec.TypeComplex:
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

func (v *syncPropertyVisitor) beginMulti(newContainer prop.Container) {
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

func (v *syncPropertyVisitor) EndChildren(container prop.Container) {
	if v.currentFrame().exitFunc != nil {
		v.currentFrame().exitFunc()
	}
	v.pop()
}

func (v *syncPropertyVisitor) push(container prop.Container) {
	v.stack = append(v.stack, &frame{
		container: container,
	})
}

func (v *syncPropertyVisitor) pop() {
	if len(v.stack) > 0 {
		v.stack = v.stack[:len(v.stack)-1]
	}
}

func (v *syncPropertyVisitor) currentFrame() *frame {
	if len(v.stack) == 0 {
		panic("stack is empty")
	}
	return v.stack[len(v.stack)-1]
}

var (
	_ prop.Visitor = (*syncPropertyVisitor)(nil)
)