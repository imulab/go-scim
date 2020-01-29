package prop

import "github.com/imulab/go-scim/pkg/v2/spec"

// Visitor defines behaviour for implementations to react to a passive Property structure traversal. It shall be used
// in cases where caller does not have knowledge of resource structure and has to rely on a spontaneous DFS traversal.
// By implementing this interface, caller will have some control over whether a property should be visited and be notified
// for entering and exiting container Property.
type Visitor interface {
	// Returns true if property should be visited; if false, the property will not be visited.
	ShouldVisit(property Property) bool
	// Visit the property, only when ShouldVisit returns true. If this method returns non-nil error,
	// the rest of the traversal will be aborted.
	Visit(property Property) error
	// Invoked when the children properties of a container property is about to be visited. The containing
	// property is supplied as an argument to provide context information. The container property itself,
	// however, has already been invoked on ShouldVisit and/or Visit.
	BeginChildren(container Property)
	// Invoked when the children properties of a container property has finished. The containing property
	// is supplied as a context argument.
	EndChildren(container Property)
}

// Visit is the entry point to visit a property in a depth-first-search fashion.
func Visit(property Property, visitor Visitor) error {
	if !visitor.ShouldVisit(property) {
		return nil
	}

	if err := visitor.Visit(property); err != nil {
		return err
	}

	if property.Attribute().MultiValued() || property.Attribute().Type() == spec.TypeComplex {
		visitor.BeginChildren(property)
		if err := property.ForEachChild(func(_ int, child Property) error {
			return Visit(child, visitor)
		}); err != nil {
			return err
		}
		visitor.EndChildren(property)
	}

	return nil
}
