package core

// Visitor can be implemented by parties that wish to traverse the Property tree.
type Visitor interface {
	// ShouldVisit returns boolean value to control whether the Property will be visited.
	ShouldVisit(prop Property) bool

	// Visit is called when shouldVisit returns true. The walkProperty process can be aborted
	// by returning an error here.
	Visit(prop Property) error

	// BeginChildren is called when walkProperty traversal is about to visit the sub properties
	// of a container property (i.e. multiValued, complex).
	BeginChildren(container Property)

	// EndChildren is called when walkProperty has ended traversal of sub properties of a container property.
	EndChildren(container Property)
}

// Visit does a DFS traversal on the given Property and invokes related methods on the visitor.
func Visit(prop Property, visitor Visitor) error {
	if !visitor.ShouldVisit(prop) {
		return nil
	}

	if err := visitor.Visit(prop); err != nil {
		return err
	}

	if prop.Attr().MultiValued || prop.Attr().Type == TypeComplex {
		visitor.BeginChildren(prop)

		if err := prop.ForEach(func(_ int, child Property) error {
			return Visit(child, visitor)
		}); err != nil {
			return err
		}

		visitor.EndChildren(prop)
	}

	return nil
}
