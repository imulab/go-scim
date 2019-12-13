package prop

import (
	"github.com/imulab/go-scim/pkg/core/spec"
	"strings"
)

// Create a new unassigned property based on the given attribute, and set its parent. The
// method will panic if anything went wrong.
func New(attr *spec.Attribute, parent Container) Property {
	if attr.MultiValued() {
		return NewMulti(attr, parent)
	} else {
		switch attr.Type() {
		case spec.TypeString:
			return NewString(attr, parent)
		case spec.TypeInteger:
			return NewInteger(attr, parent)
		case spec.TypeDecimal:
			return NewDecimal(attr, parent)
		case spec.TypeBoolean:
			return NewBoolean(attr, parent)
		case spec.TypeDateTime:
			return NewDateTime(attr, parent)
		case spec.TypeReference:
			return NewReference(attr, parent)
		case spec.TypeBinary:
			return NewBinary(attr, parent)
		case spec.TypeComplex:
			return NewComplex(attr, parent)
		default:
			panic("invalid type")
		}
	}
}

// Create a new unassigned string property. The method will panic if
// given attribute is not singular string type.
func NewString(attr *spec.Attribute, parent Container) Property {
	if !attr.SingleValued() || attr.Type() != spec.TypeString {
		panic("invalid attribute for string property")
	}
	p := &stringProperty{
		parent:      parent,
		attr:        attr,
		value:       nil,
		subscribers: []Subscriber{},
	}
	subscribeWithAnnotation(p)
	return p
}

// Create a new string property with given value. The method will panic if
// given attribute is not singular string type. The property will be
// marked dirty at the start.
func NewStringOf(attr *spec.Attribute, parent Container, value interface{}) Property {
	p := NewString(attr, parent)
	if err := p.Replace(value); err != nil {
		panic(err)
	}
	return p
}

// Create a new unassigned integer property. The method will panic if
// given attribute is not singular integer type.
func NewInteger(attr *spec.Attribute, parent Container) Property {
	if !attr.SingleValued() || attr.Type() != spec.TypeInteger {
		panic("invalid attribute for integer property")
	}
	p := &integerProperty{
		parent:      parent,
		attr:        attr,
		value:       nil,
		subscribers: []Subscriber{},
	}
	subscribeWithAnnotation(p)
	return p
}

// Create a new integer property with given value. The method will panic if
// given attribute is not singular integer type. The property will be
// marked dirty at the start.
func NewIntegerOf(attr *spec.Attribute, parent Container, value interface{}) Property {
	p := NewInteger(attr, parent)
	if err := p.Replace(value); err != nil {
		panic(err)
	}
	return p
}

// Create a new unassigned decimal property. The method will panic if
// given attribute is not singular decimal type.
func NewDecimal(attr *spec.Attribute, parent Container) Property {
	if !attr.SingleValued() || attr.Type() != spec.TypeDecimal {
		panic("invalid attribute for integer property")
	}
	p := &decimalProperty{
		parent:      parent,
		attr:        attr,
		value:       nil,
		subscribers: []Subscriber{},
	}
	subscribeWithAnnotation(p)
	return p
}

// Create a new decimal property with given value. The method will panic if
// given attribute is not singular decimal type. The property will be
// marked dirty at the start.
func NewDecimalOf(attr *spec.Attribute, parent Container, value interface{}) Property {
	p := NewDecimal(attr, parent)
	if err := p.Replace(value); err != nil {
		panic(err)
	}
	return p
}

// Create a new unassigned boolean property. The method will panic if
// given attribute is not singular boolean type.
func NewBoolean(attr *spec.Attribute, parent Container) Property {
	if !attr.SingleValued() || attr.Type() != spec.TypeBoolean {
		panic("invalid attribute for boolean property")
	}
	p := &booleanProperty{
		parent:      parent,
		attr:        attr,
		value:       nil,
		subscribers: []Subscriber{},
	}
	subscribeWithAnnotation(p)
	return p
}

// Create a new boolean property with given value. The method will panic if
// given attribute is not singular boolean type. The property will be
// marked dirty at the start.
func NewBooleanOf(attr *spec.Attribute, parent Container, value interface{}) Property {
	p := NewBoolean(attr, parent)
	if err := p.Replace(value); err != nil {
		panic(err)
	}
	return p
}

// Create a new unassigned string property. The method will panic if
// given attribute is not singular dateTime type.
func NewDateTime(attr *spec.Attribute, parent Container) Property {
	if !attr.SingleValued() || attr.Type() != spec.TypeDateTime {
		panic("invalid attribute for dateTime property")
	}
	p := &dateTimeProperty{
		parent:      parent,
		attr:        attr,
		value:       nil,
		subscribers: []Subscriber{},
	}
	subscribeWithAnnotation(p)
	return p
}

// Create a new string property with given value. The method will panic if
// given attribute is not singular dateTime type or the value is not of ISO8601 format.
// The property will be marked dirty at the start.
func NewDateTimeOf(attr *spec.Attribute, parent Container, value interface{}) Property {
	p := NewDateTime(attr, parent)
	if err := p.Replace(value); err != nil {
		panic(err)
	}
	return p
}

// Create a new unassigned reference property. The method will panic if
// given attribute is not singular reference type.
func NewReference(attr *spec.Attribute, parent Container) Property {
	if !attr.SingleValued() || attr.Type() != spec.TypeReference {
		panic("invalid attribute for reference property")
	}
	p := &referenceProperty{
		parent:      parent,
		attr:        attr,
		value:       nil,
		subscribers: []Subscriber{},
	}
	subscribeWithAnnotation(p)
	return p
}

// Create a new reference property with given value. The method will panic if
// given attribute is not singular reference type. The property will be
// marked dirty at the start.
func NewReferenceOf(attr *spec.Attribute, parent Container, value interface{}) Property {
	p := NewReference(attr, parent)
	if err := p.Replace(value); err != nil {
		panic(err)
	}
	return p
}

// Create a new unassigned binary property. The method will panic if
// given attribute is not singular binary type.
func NewBinary(attr *spec.Attribute, parent Container) Property {
	if !attr.SingleValued() || attr.Type() != spec.TypeBinary {
		panic("invalid attribute for binary property")
	}
	p := &binaryProperty{
		parent:      parent,
		attr:        attr,
		value:       nil,
		subscribers: []Subscriber{},
	}
	subscribeWithAnnotation(p)
	return p
}

// Create a new binary property with given base64 encoded value. The method will panic if
// given attribute is not singular binary type. The property will be
// marked dirty at the start.
func NewBinaryOf(attr *spec.Attribute, parent Container, value interface{}) Property {
	p := NewBinary(attr, parent)
	if err := p.Replace(value); err != nil {
		panic(err)
	}
	return p
}

// Create a new unassigned complex property. The method will panic if
// given attribute is not singular complex type.
func NewComplex(attr *spec.Attribute, parent Container) Property {
	if !attr.SingleValued() || attr.Type() != spec.TypeComplex {
		panic("invalid attribute for complex property")
	}
	var p = &complexProperty{
		parent:    parent,
		attr:      attr,
		subProps:  make([]Property, 0, attr.CountSubAttributes()),
		nameIndex: make(map[string]int),
	}
	subscribeWithAnnotation(p)
	attr.ForEachSubAttribute(func(subAttribute *spec.Attribute) {
		p.subProps = append(p.subProps, New(subAttribute, p))
		p.nameIndex[strings.ToLower(subAttribute.Name())] = len(p.subProps) - 1
	})
	return p
}

// Create a new complex property with given value. The method will panic if
// given attribute is not singular complex type. The property will be
// marked dirty at the start unless value is empty
func NewComplexOf(attr *spec.Attribute, parent Container, value interface{}) Property {
	p := NewComplex(attr, parent)
	if err := p.Add(value); err != nil {
		panic(err)
	}
	return p
}

// Create a new unassigned multiValued property. The method will panic if
// given attribute is not multiValued type.
func NewMulti(attr *spec.Attribute, parent Container) Property {
	if !attr.MultiValued() {
		panic("invalid attribute for multiValued property")
	}
	p := &multiValuedProperty{
		parent:      parent,
		attr:        attr,
		elements:    make([]Property, 0),
		subscribers: []Subscriber{},
	}
	subscribeWithAnnotation(p)
	return p
}

// Create a new multiValued property with given value. The method will panic if
// given attribute is not multiValued type. The property will be
// marked dirty at the start.
func NewMultiOf(attr *spec.Attribute, parent Container, value interface{}) Property {
	p := NewMulti(attr, parent)
	if err := p.Add(value); err != nil {
		panic(err)
	}
	return p
}
