package scim

type ResourceType[T any] struct {
	id          string
	name        string
	description string
	endpoint    string
	location    string
	schema      *Schema
	extensions  []*Schema
	newModel    func() *T
	mappings    []*Mapping[T]
}
