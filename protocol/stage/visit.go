package stage

import (
	"context"
	"github.com/imulab/go-scim/core"
)

type (
	// Function signature that runs the given resource through all necessary filters, with the help of a reference
	// resource, and return any error
	FilterStage func(ctx context.Context, resource *core.Resource, ref *core.Resource) error

	// Implementation of core.Visitor that visits all properties within a resource, trying to invoke all property
	// filters assigned to it. This visitor can also be equipped with a reference navigator, which navigates through
	// a reference resource, to keep a reference property in async with the currently visited property for reference
	// and comparison purposes.
	filterVisitor struct {
		// The context to use for all filters. If nil, a context.Background() will be used.
		context context.Context
		// The resource currently being visited. This field must be set.
		resource *core.Resource
		// The reference currently serving as a reference to the resource being visited. This
		// field is optional. If nil, it means the resource is being visited without reference.
		// This will impact the behaviour of the filter.
		ref *core.Resource
		// Navigator of the reference resource. The currently focused property on the navigator serves as the reference
		// container to search for the corresponding reference property to the visited property, if the current stack
		// frame indicates the two containers are in sync.
		refNav core.Navigator
		// Map of attribute ids to all property filters handling the attribute. The filters for an attribute
		// will be invoked in order when the property with that attribute is visited, along with the reference
		// property, if necessary. Any error from the filter will results in the visitor's premature exit.
		filters map[string][]PropertyFilter
		// stack frames
		stack []*frame
	}

	// Stack frame to keep track during the resource traversal.
	frame struct {
		// Container properties for all properties being visited in the current frame
		container core.Property
		// true, if the currently focused property on refNav corresponds to the above container
		inSync bool
		// function to invoke when the current frame is about to be retired
		exitFunc func()
	}
)

// Construct and return a filter executor with reference to run during the filter stage.
func NewFilterStage(resourceTypes []*core.ResourceType, filters []PropertyFilter) FilterStage {
	index := buildIndex(resourceTypes, filters)
	return func(ctx context.Context, resource *core.Resource, ref *core.Resource) error {
		if ref == nil {
			return resource.Visit(&filterVisitor{
				context:  ctx,
				resource: resource,
				filters:  index,
				stack:    make([]*frame, 0),
			})
		} else {
			return resource.Visit(&filterVisitor{
				context:  ctx,
				resource: resource,
				ref:      ref,
				refNav:   core.NewNavigator(ref),
				filters:  index,
				stack:    make([]*frame, 0),
			})
		}
	}
}

func (v *filterVisitor) ShouldVisit(property core.Property) bool {
	return true
}

func (v *filterVisitor) Visit(property core.Property) error {
	filters, ok := v.filters[property.Attribute().Id]
	if !ok || len(filters) == 0 {
		return nil
	}

	var ctx context.Context
	{
		if v.context != nil {
			ctx = v.context
		} else {
			ctx = context.Background()
		}
	}

	// Short circuit: no reference case, just invoke all filters in sequence.
	if v.ref == nil {
		for _, filter := range filters {
			err := filter.FilterOnCreate(ctx, v.resource, property)
			if err != nil {
				return err
			}
		}
		return nil
	}

	var ref core.Property
	{
		if !v.currentFrame().inSync {
			ref = nil
		} else {
			ref = v.getReferenceFromNav(property)
		}
	}
	for _, filter := range filters {
		err := filter.FilterOnUpdate(ctx, v.resource, property, v.ref, ref)
		if err != nil {
			return err
		}
	}

	return nil
}

func (v *filterVisitor) getReferenceFromNav(property core.Property) core.Property {
	base := v.refNav.Current()
	switch {
	case base.Attribute().MultiValued:
		if ref, err := v.refNav.Focus(func(elem core.Property) bool {
			return property.(core.EqualAware).Matches(elem)
		}); err == nil {
			v.refNav.Release()
			return ref
		} else {
			return nil
		}
	case base.Attribute().Type == core.TypeComplex:
		if ref, err := v.refNav.Focus(property.Attribute().Name); err == nil {
			v.refNav.Release()
			return ref
		} else {
			return nil
		}
	default:
		panic("invalid state")
	}
}

func (v *filterVisitor) BeginComplex(complex core.Property) {
	v.push(complex)

	// Short circuit: no reference to worry about
	if v.refNav == nil {
		return
	}

	// Short circuit: the only frame is the one we just pushed => meaning we are on the top level.
	// At the very start, BeginComplex is invoked on the base complex property by the visitor.
	// And reference navigator, if exists, is already focused on the base property. Hence, we
	// immediately assumes they are in sync.
	if len(v.stack) == 1 {
		v.currentFrame().inSync = true
		v.currentFrame().exitFunc = func() {
			v.refNav.Release()
		}
		return
	}

	// Another short circuit: no need to sync up if the current frame is not in sync
	if !v.currentFrame().inSync {
		return
	}

	// After this point, we are sure there is at least one frame, meaning we are not on top level.
	var (
		selector interface{}
		attr = v.currentFrame().container.Attribute()
	)
	switch {
	case attr.MultiValued:
		// complex as an element of a multiValued container => find the match
		selector = func(property core.Property) bool {
			return complex.(core.EqualAware).Matches(property)
		}
	case attr.Type == core.TypeComplex:
		// complex as a sub property of a complex container => find by name
		selector = complex.Attribute().Name
	default:
		panic("invalid frame")
	}

	if _, err := v.refNav.Focus(selector); err == nil {
		v.currentFrame().inSync = true
		v.currentFrame().exitFunc = func() {
			v.refNav.Release()
		}
	}
}

func (v *filterVisitor) EndComplex(complex core.Property) {
	if v.currentFrame().exitFunc != nil {
		v.currentFrame().exitFunc()
	}
	v.pop()
}

func (v *filterVisitor) BeginMulti(multi core.Property) {
	// Short circuit: no reference to worry about
	if v.refNav == nil {
		v.push(multi)
		return
	}

	// Short circuit: the current frame is not in sync
	if !v.currentFrame().inSync {
		v.push(multi)
		return
	}

	// After this point, we are sure we have a reference and it is in sync.
	// Because multiValued property can appear as a sub property of a complex property, we select from the refNav
	// directly by name
	v.push(multi)
	if _, err := v.refNav.Focus(multi.Attribute().Name); err == nil {
		v.currentFrame().inSync = true
		v.currentFrame().exitFunc = func() {
			v.refNav.Release()
		}
	}
}

func (v *filterVisitor) EndMulti(multi core.Property) {
	if v.currentFrame().exitFunc != nil {
		v.currentFrame().exitFunc()
	}
	v.pop()
}

func (v *filterVisitor) push(container core.Property) {
	v.stack = append(v.stack, &frame{
		container: container,
	})
}

func (v *filterVisitor) pop() {
	if len(v.stack) > 0 {
		v.stack = v.stack[:len(v.stack)-1]
	}
}

func (v *filterVisitor) currentFrame() *frame {
	if len(v.stack) == 0 {
		panic("stack is empty")
	}
	return v.stack[len(v.stack)-1]
}
