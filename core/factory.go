package core

var (
	// Entry point to create a new resource
	Resources = resourceFactory{}
	// Entry point to create a new property
	Properties = propertyFactory{}
)

type (
	resourceFactory struct{}
	propertyFactory struct{}
)

// Create a new empty resource. All properties within this resource will be unassigned.
func (f resourceFactory) New(rt *ResourceType) *Resource {
	return &Resource{
		rt: rt,
		base: Properties.NewComplex(&Attribute{
			Type:          TypeComplex,
			SubAttributes: rt.DerivedAttributes(),
			Mutability:    MutabilityReadWrite,
			Returned:      ReturnedDefault,
			Uniqueness:    UniquenessNone,
			Metadata:      &Metadata{},
		}),
	}
}

// Create a new unassigned property of any type.
func (f propertyFactory) New(attr *Attribute) Property {
	if attr.MultiValued {
		return &multiValuedProperty{attr: attr, props: make([]Property, 0)}
	} else {
		switch attr.Type {
		case TypeString:
			return &stringProperty{attr: attr, v: nil}
		case TypeInteger:
			return &integerProperty{attr: attr, v: nil}
		case TypeDecimal:
			return &decimalProperty{attr: attr, v: nil}
		case TypeBoolean:
			return &booleanProperty{attr: attr, v: nil}
		case TypeDateTime:
			return &dateTimeProperty{attr: attr, v: nil}
		case TypeReference:
			return &referenceProperty{attr: attr, v: nil}
		case TypeBinary:
			return &binaryProperty{attr: attr, v: nil}
		case TypeComplex:
			return f.NewComplex(attr)
		default:
			panic("invalid attribute type")
		}
	}
}

// Create a new unassigned complex property.
func (f propertyFactory) NewComplex(attr *Attribute) *complexProperty {
	container := &complexProperty{
		attr:  attr,
		props: make([]Property, 0, len(attr.SubAttributes)),
	}

	for _, subAttr := range attr.SubAttributes {
		container.props = append(container.props, f.New(subAttr))
	}
	container.syncIndex()

	return container
}
