package prop

import "github.com/elvsn/scim.go/spec"

// NewResource creates a resource prototype of the attributes defined in the resource type, along with the core SCIM attributes.
func NewResource(resourceType *spec.ResourceType) *Resource {
	r := Resource{
		resourceType: resourceType,
		data:         NewComplex(resourceType.SuperAttribute(true)).(*complexProperty),
	}
	return &r
}

// Resource represents a SCIM resource. It is a wrapper around the root property.
type Resource struct {
	resourceType *spec.ResourceType
	data         *complexProperty
}

// ResourceType returns the resource type of this resource
func (r *Resource) ResourceType() *spec.ResourceType {
	return r.resourceType
}

// RootAttribute returns the attribute of the root property
func (r *Resource) RootAttribute() *spec.Attribute {
	return r.data.Attribute()
}

// RootProperty returns the root property
func (r *Resource) RootProperty() Property {
	return r.data
}

// Hash returns the hash of this resource, which is same hash of the root property.
func (r *Resource) Hash() uint64 {
	return r.data.Hash()
}

// Navigator returns a navigator on the root property.
func (r *Resource) Navigator() *Navigator {
	return Navigate(r.data)
}

// Visit starts a DFS visit on the root property of the resource.
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

// IdOrEmpty returns the id of the resource, defined in the core schema. If in any case the id is not available, (i.e.
// unassigned, wrong type), empty string is returned.
func (r *Resource) IdOrEmpty() string {
	if p, err := r.data.ChildAtIndex("id"); err != nil || p.IsUnassigned() {
		return ""
	} else if s, ok := p.Raw().(string); !ok {
		return ""
	} else {
		return s
	}
}

// MetaLocationOrEmpty returns meta.location value of the resource, defined in the core schema. If in any case, the
// meta.location value is not available (i.e. unassigned, wrong type), empty string is returned.
func (r *Resource) MetaLocationOrEmpty() string {
	meta, err := r.data.ChildAtIndex("meta")
	if err != nil {
		return ""
	}

	loc, err := meta.ChildAtIndex("location")
	if err != nil {
		return ""
	}

	if loc.IsUnassigned() {
		return ""
	} else if s, ok := loc.Raw().(string); !ok {
		return ""
	} else {
		return s
	}
}

// MetaVersionOrEmpty returns meta.version value of the resource, defined in the core schema. If in any case, the
// meta.version value is not available (i.e. unassigned, wrong type), empty string is returned.
func (r *Resource) MetaVersionOrEmpty() string {
	meta, err := r.data.ChildAtIndex("meta")
	if err != nil {
		return ""
	}

	loc, err := meta.ChildAtIndex("version")
	if err != nil {
		return ""
	}

	if loc.IsUnassigned() {
		return ""
	} else if s, ok := loc.Raw().(string); !ok {
		return ""
	} else {
		return s
	}
}
