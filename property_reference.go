package scim

// TODO complete the rest of properties
// TODO remove simpleProperty
// TODO evaluator
// TODO traverseQualifiedElements
// TODO crud methods on resources

type referenceProperty struct {
	*stringProperty
}

func (p *referenceProperty) greaterThan(_ any) bool {
	return false
}

func (p *referenceProperty) lessThan(_ any) bool {
	return false
}
