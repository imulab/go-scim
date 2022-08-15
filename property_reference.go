package scim

type referenceProperty struct {
	*stringProperty
}

func (p *referenceProperty) greaterThan(_ any) bool {
	return false
}

func (p *referenceProperty) lessThan(_ any) bool {
	return false
}
