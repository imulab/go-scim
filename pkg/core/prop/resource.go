package prop

import (
	"github.com/imulab/go-scim/pkg/core/spec"
)

// Create a new resource of the given resource type. The method panics if something went wrong.
func NewResource(resourceType *spec.ResourceType) *Resource {
	return &Resource{
		resourceType: resourceType,
		data:         NewComplex(resourceType.SuperAttribute(true), nil).(*complexProperty),
	}
}

// Create a new resource of the given resource type and value. The method panics if something went wrong.
func NewResourceOf(resourceType *spec.ResourceType, value interface{}) *Resource {
	resource := NewResource(resourceType)
	if err := resource.data.Replace(value); err != nil {
		panic(err)
	}
	return resource
}

// Resource represents a SCIM resource. This is the main object of interaction in the SCIM spec. It is implemented
// as a wrapper around the top level complex property and the resource type.
type Resource struct {
	resourceType *spec.ResourceType
	data         *complexProperty
}

// Return a clone of this resource. The clone will contain properties that share the same instance of attribute and
// subscribers with the original property before the clone, but retain separate instance of values.
func (r *Resource) Clone() *Resource {
	return &Resource{
		resourceType: r.resourceType,
		data:         r.data.Clone(nil).(*complexProperty),
	}
}

// Return the resource type of this resource
func (r *Resource) ResourceType() *spec.ResourceType {
	return r.resourceType
}

// Return the super attribute that describes this resource.
func (r *Resource) SuperAttribute() *spec.Attribute {
	return r.data.Attribute()
}

// Adapting constructor method to return a new navigator for the top level property of the resource.
func (r *Resource) NewNavigator() *Navigator {
	return NewNavigator(r.data)
}

// Adapting constructor method to return a new fluent navigator for the top level property of the resource.
func (r *Resource) NewFluentNavigator() *FluentNavigator {
	return NewFluentNavigator(r.data)
}

// Convenience method to return the ID of the resource. If id does not exist, return empty string.
func (r *Resource) ID() string {
	p := r.data.ChildAtIndex("id")
	if p == nil || p.IsUnassigned() {
		return ""
	}
	return p.Raw().(string)
}

// Convenience method to return the meta.location field of the resource, or empty string if it does not exist.
func (r *Resource) Location() string {
	if nav := r.NewFluentNavigator().FocusName("meta").FocusName("location"); nav.Error() != nil {
		return ""
	} else {
		return nav.Current().Raw().(string)
	}
}

// Convenience method to return the meta.version field of the resource, or empty string if it does not exist
func (r *Resource) Version() string {
	if nav := r.NewFluentNavigator().FocusName("meta").FocusName("version"); nav.Error() != nil {
		return ""
	} else {
		return nav.Current().Raw().(string)
	}
}

// Adapting method to start a DFS visit on the top level property of the resource.
func (r *Resource) Visit(visitor Visitor) error {
	visitor.BeginChildren(r.data)
	for _, prop := range r.data.subProps {
		if err := Visit(prop, visitor); err != nil {
			return err
		}
	}
	visitor.EndChildren(r.data)
	return nil
}

// Return the hash of the resource at its current state.
func (r *Resource) Hash() uint64 {
	return r.data.Hash()
}

// Add value to the top level of the resource.
func (r *Resource) Add(value interface{}) error {
	return r.data.Add(value)
}

// Replace value on the top level of the resource.
func (r *Resource) Replace(value interface{}) error {
	return r.data.Replace(value)
}

const (
	id = "id"
	meta = "meta"
	location = "location"
	version = "version"
)