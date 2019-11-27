package prop

import "github.com/imulab/go-scim/src/core"

// Create a new resource of the given resource type. The method panics if something went wrong.
func NewResource(resourceType *core.ResourceType) *Resource {
	return &Resource{
		resourceType: resourceType,
		data:         NewComplex(resourceType.SuperAttribute(true)).(*complexProperty),
	}
}

// Create a new resource of the given resource type and value. The method panics if something went wrong.
func NewResourceOf(resourceType *core.ResourceType, value interface{}) *Resource {
	resource := NewResource(resourceType)
	err := resource.replace(value)
	if err != nil {
		panic(err)
	}
	return resource
}

// Resource represents a SCIM resource. This is the main object of interaction in the SCIM spec. It is implemented
// as a wrapper around the top level complex property and the resource type.
type Resource struct {
	resourceType *core.ResourceType
	data         *complexProperty
}

// Adapting constructor method to return a new navigator for the top level property of the resource.
func (r *Resource) NewNavigator() *Navigator {
	return NewNavigator(r.data)
}

// Adapting method to start a DFS visit on the top level property of the resource.
func (r *Resource) Visit(visitor core.Visitor) error {
	return core.Visit(r.data, visitor)
}

// Internal adapter to the Replace method of the data Property. It is used exclusively by package
// methods. Other modifications to the resource should go through Navigator.
func (r *Resource) replace(value interface{}) error {
	_, err := r.data.Replace(value)
	return err
}
