package core

import (
	"encoding/base64"
	"strings"
	"time"
)

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
		attr:     attr,
		subProps: make(map[string]Property, 0),
	}

	for _, subAttr := range attr.SubAttributes {
		container.subProps[strings.ToLower(subAttr.Name)] = f.New(subAttr)
	}

	return container
}

func (f propertyFactory) NewComplexOf(attr *Attribute, value map[string]interface{}) *complexProperty {
	prop := f.NewComplex(attr)
	for k, v := range value {
		_ = prop.Replace(Steps.NewPath(k), v)
	}
	return prop
}

func (f propertyFactory) NewStringOf(attr *Attribute, value string) *stringProperty {
	return &stringProperty{attr: attr, v: &value}
}

func (f propertyFactory) NewIntegerOf(attr *Attribute, value int64) *integerProperty {
	return &integerProperty{attr: attr, v: &value}
}

func (f propertyFactory) NewDecimalOf(attr *Attribute, value float64) *decimalProperty {
	return &decimalProperty{attr: attr, v: &value}
}

func (f propertyFactory) NewBoolean(attr *Attribute) *booleanProperty {
	return &booleanProperty{attr: attr}
}

func (f propertyFactory) NewBooleanOf(attr *Attribute, value bool) *booleanProperty {
	return &booleanProperty{attr: attr, v: &value}
}

func (f propertyFactory) NewDateTimeOf(attr *Attribute, value string) *dateTimeProperty {
	_, err := time.Parse(ISO8601, value)
	if err != nil {
		panic(Errors.invalidValue("invalid date time value"))
	}
	return &dateTimeProperty{attr: attr, v: &value}
}

func (f propertyFactory) NewBinaryOf(attr *Attribute, value string) *binaryProperty {
	_, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		panic(Errors.invalidValue("invalid binary value"))
	}
	return &binaryProperty{attr: attr, v: &value}
}

func (f propertyFactory) NewReferenceOf(attr *Attribute, value string) *referenceProperty {
	return &referenceProperty{attr: attr, v: &value}
}
