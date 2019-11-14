package core

// Interface to implement if the caller wish to visit the property.
type Visitor interface {
	// Returns true if property should be visited; if false, the property will not be visited.
	ShouldVisit(property Property) bool
	// Visit the property, only when ShouldVisit returns true. If this method returns non-nil error,
	// the rest of the traversal will be aborted.
	Visit(property Property) error
	// Invoked when the sub attributes of a complex attribute is about to be visited.
	BeginComplex()
	// Invoked when all sub attributes of a complex attribute has been visited.
	EndComplex()
	// Invoked when the elements of a multiValued attribute is about to be visited.
	BeginMulti()
	// Invoked when all elements of a multiValued attribute has been visited.
	EndMulti()
}

// Universal interface for all properties to implement to reverse the control of the visitor.
type Visitable interface {
	Property
	// Invoke the visitor on this property.
	VisitedBy(visitor Visitor) error
}

func (s *stringProperty) VisitedBy(visitor Visitor) error {
	return visitor.Visit(s)
}

func (i *integerProperty) VisitedBy(visitor Visitor) error {
	return visitor.Visit(i)
}

func (d *decimalProperty) VisitedBy(visitor Visitor) error {
	return visitor.Visit(d)
}

func (b *booleanProperty) VisitedBy(visitor Visitor) error {
	return visitor.Visit(b)
}

func (d *dateTimeProperty) VisitedBy(visitor Visitor) error {
	return visitor.Visit(d)
}

func (r *referenceProperty) VisitedBy(visitor Visitor) error {
	return visitor.Visit(r)
}

func (b *binaryProperty) VisitedBy(visitor Visitor) error {
	return visitor.Visit(b)
}

func (c *complexProperty) VisitedBy(visitor Visitor) error {
	err := visitor.Visit(c)
	if err != nil {
		return err
	}

	visitor.BeginComplex()
	for _, subProp := range c.subProps {
		if !visitor.ShouldVisit(subProp) {
			continue
		}

		err := subProp.(Visitable).VisitedBy(visitor)
		if err != nil {
			return err
		}
	}
	visitor.EndComplex()

	return nil
}

func (m *multiValuedProperty) VisitedBy(visitor Visitor) error {
	err := visitor.Visit(m)
	if err != nil {
		return err
	}

	visitor.BeginMulti()
	for _, elemProp := range m.props {
		if !visitor.ShouldVisit(elemProp) {
			continue
		}

		err := elemProp.(Visitable).VisitedBy(visitor)
		if err != nil {
			return err
		}
	}
	visitor.EndMulti()

	return nil
}

// implementation checks
var (
	_ Visitable = (*stringProperty)(nil)
	_ Visitable = (*integerProperty)(nil)
	_ Visitable = (*decimalProperty)(nil)
	_ Visitable = (*booleanProperty)(nil)
	_ Visitable = (*dateTimeProperty)(nil)
	_ Visitable = (*binaryProperty)(nil)
	_ Visitable = (*referenceProperty)(nil)
	_ Visitable = (*complexProperty)(nil)
	_ Visitable = (*multiValuedProperty)(nil)
)