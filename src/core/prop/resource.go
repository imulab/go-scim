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
	_, err := resource.data.Replace(value)
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
	lastModCount int
}

// Return the resource type of this resource
func (r *Resource) ResourceType() *core.ResourceType {
	return r.resourceType
}

// Adapting constructor method to return a new navigator for the top level property of the resource.
func (r *Resource) NewNavigator() *Navigator {
	return NewNavigator(r.data)
}

// Convenience method to return the ID of the resource.
func (r *Resource) ID() string {
	p, err := r.NewNavigator().FocusName("id")
	if err != nil {
		return ""
	}

	if p.IsUnassigned() {
		return ""
	}

	if id, ok := p.Raw().(string); !ok {
		return ""
	} else {
		return id
	}
}

// Convenience method to return the meta.location field of the resource.
func (r *Resource) Location() string {
	nav := r.NewNavigator()

	_, err := nav.FocusName("meta")
	if err != nil {
		return ""
	}
	p, err := nav.FocusName("location")
	if err != nil {
		return ""
	}

	if p.IsUnassigned() {
		return ""
	}

	if location, ok := p.Raw().(string); !ok {
		return ""
	} else {
		return location
	}
}

// Convenience method to return the meta.version field of the resource.
func (r *Resource) Version() string {
	nav := r.NewNavigator()

	_, err := nav.FocusName("meta")
	if err != nil {
		return ""
	}
	p, err := nav.FocusName("version")
	if err != nil {
		return ""
	}

	if p.IsUnassigned() {
		return ""
	}

	if version, ok := p.Raw().(string); !ok {
		return ""
	} else {
		return version
	}
}

// Adapting method to start a DFS visit on the top level property of the resource.
func (r *Resource) Visit(visitor core.Visitor) error {
	visitor.BeginChildren(r.data)
	for _, prop := range r.data.subProps {
		if err := core.Visit(prop, visitor); err != nil {
			return err
		}
	}
	visitor.EndChildren(r.data)
	return nil
}

// Return the total mod count of this resource at its current state.
func (r *Resource) ModCount() int {
	mc := &modCounter{}
	_ = r.Visit(mc)
	return mc.total
}

// Return the hash of the resource at its current state.
func (r *Resource) Hash() uint64 {
	return r.data.Hash()
}

// Visitor implementation to sum up all mod counts.
type modCounter struct {
	total int
}

func (mc *modCounter) ShouldVisit(property core.Property) bool {
	_, ok := property.(core.Container)
	return !ok // only visits non-container to avoid double counting
}

func (mc *modCounter) Visit(property core.Property) error {
	mc.total += property.ModCount()
}

func (mc *modCounter) BeginChildren(container core.Container) {}

func (mc *modCounter) EndChildren(container core.Container) {}
