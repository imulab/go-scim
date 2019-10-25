package core

import (
	"fmt"
	"strings"
)

// Property represents a field in SCIM resource. It maintains the field metadata its attributes.
type Property interface {
	// Return the attribute for this property
	Attribute() *Attribute

	// Return the raw value for this property. Raw values must follow
	// correspondence rules to Go types. Check with implementations to
	// see the SCIM-GO type correspondence.
	Raw() interface{}

	// Return a slice of children properties. If this property does not have
	// children properties, return empty slice.
	Children() []Property

	// Return true if this property's value is unassigned. Unassigned, according to RFC7643,
	// is defined to be zero-length arrays for multiValued properties and null for others.
	IsUnassigned() bool

	// Return true if the property's value is present. Presence refers to non-null, non-empty.
	// This method corresponds directly to the 'pr' operator.
	IsPresent() bool
}

// A SCIM string property. The value is stored as a pointer to Go's builtin string type.
// The Raw() value call returns either nil or string type.
type stringProperty struct {
	attr *Attribute
	v    *string
}

func (s *stringProperty) Attribute() *Attribute {
	return s.attr
}

func (s *stringProperty) Raw() interface{} {
	if s.v == nil {
		return nil
	}
	return *(s.v)
}

func (s *stringProperty) Children() []Property {
	return []Property{}
}

func (s *stringProperty) IsUnassigned() bool {
	return s.v == nil
}

func (s *stringProperty) IsPresent() bool {
	return s.v != nil && len(*s.v) > 0
}

// A SCIM integer property. The value is stored as a pointer to Go's builtin int64 type.
// The Raw() value call returns either nil or int64 type.
type integerProperty struct {
	attr *Attribute
	v    *int64
}

func (i *integerProperty) Attribute() *Attribute {
	return i.attr
}

func (i *integerProperty) Raw() interface{} {
	if i.v == nil {
		return nil
	}
	return *(i.v)
}

func (i *integerProperty) Children() []Property {
	return []Property{}
}

func (i *integerProperty) IsUnassigned() bool {
	return i.v == nil
}

func (i *integerProperty) IsPresent() bool {
	return i.v != nil
}

func (i *integerProperty) Hash() int64 {
	if i.v == nil {
		return 0
	}
	return *(i.v)
}

// A SCIM decimal property. The value is stored as a pointer to Go's builtin float64 type.
// The Raw() value call returns either nil or float64 type.
type decimalProperty struct {
	attr *Attribute
	v    *float64
}

func (d *decimalProperty) Attribute() *Attribute {
	return d.attr
}

func (d *decimalProperty) Raw() interface{} {
	if d.v == nil {
		return nil
	}
	return *(d.v)
}

func (d *decimalProperty) Children() []Property {
	return []Property{}
}

func (d *decimalProperty) IsUnassigned() bool {
	return d.v == nil
}

func (d *decimalProperty) IsPresent() bool {
	return d.v != nil
}

// A SCIM boolean property. The value is stored as a pointer to Go's builtin bool type.
// The Raw() value call returns either nil or bool type. A nil value shall be regarded as false.
type booleanProperty struct {
	attr *Attribute
	v    *bool
}

func (b *booleanProperty) Attribute() *Attribute {
	return b.attr
}

func (b *booleanProperty) Raw() interface{} {
	if b.v == nil {
		return nil
	}
	return *(b.v)
}

func (b *booleanProperty) Children() []Property {
	return []Property{}
}

func (b *booleanProperty) IsUnassigned() bool {
	return b.v == nil
}

func (b *booleanProperty) IsPresent() bool {
	return b.v != nil
}

// A SCIM binary property. The value is stored as a pointer to Go's builtin string type, but the content
// must be the base64 representation of bytes. The Raw() value call returns either nil or string type.
type binaryProperty struct {
	attr *Attribute
	v    *string
}

func (b *binaryProperty) Attribute() *Attribute {
	return b.attr
}

func (b *binaryProperty) Raw() interface{} {
	if b.v == nil {
		return nil
	}
	return *(b.v)
}

func (b *binaryProperty) Children() []Property {
	return []Property{}
}

func (b *binaryProperty) IsUnassigned() bool {
	return b.v == nil
}

func (b *binaryProperty) IsPresent() bool {
	return b.v != nil && len(*(b.v)) > 0
}

// A SCIM dateTime property. The value is stored as a pointer to Go's builtin string type, but the content
// must be a valid SCIM dateTime. The Raw() value call returns either nil or string type.
type dateTimeProperty struct {
	attr *Attribute
	v    *string
}

func (d *dateTimeProperty) Attribute() *Attribute {
	return d.attr
}

func (d *dateTimeProperty) Raw() interface{} {
	if d.v == nil {
		return nil
	}
	return *(d.v)
}

func (d *dateTimeProperty) Children() []Property {
	return []Property{}
}

func (d *dateTimeProperty) IsUnassigned() bool {
	return d.v == nil
}

func (d *dateTimeProperty) IsPresent() bool {
	return d.v != nil && len(*(d.v)) > 0
}

// A SCIM reference property. The value is stored as a pointer to Go's builtin string type.
// The Raw() value call returns either nil or string type.
type referenceProperty struct {
	attr *Attribute
	v    *string
}

func (r *referenceProperty) Attribute() *Attribute {
	return r.attr
}

func (r *referenceProperty) Raw() interface{} {
	if r.v == nil {
		return nil
	}
	return *(r.v)
}

func (r *referenceProperty) Children() []Property {
	return []Property{}
}

func (r *referenceProperty) IsUnassigned() bool {
	return r.v == nil
}

func (r *referenceProperty) IsPresent() bool {
	return r.v != nil && len(*r.v) > 0
}

// A SCIM complex property. This property stores a collection of properties by their names. The
// Raw() value call returns a map of raw property values indexed to their attribute names. Since the call
// builds this map on the fly, such call may be expensive.
type complexProperty struct {
	attr     *Attribute
	subProps map[string]Property
}

func (c *complexProperty) Attribute() *Attribute {
	return c.attr
}

func (c *complexProperty) Raw() interface{} {
	values := make(map[string]interface{})
	for _, p := range c.subProps {
		values[p.Attribute().Name] = p.Raw()
	}
	return values
}

func (c *complexProperty) Children() []Property {
	subProps := make([]Property, 0)
	for _, p := range c.subProps {
		subProps = append(subProps, p)
	}
	return subProps
}

// If all sub properties are unassigned, this complex property is unassigned.
func (c *complexProperty) IsUnassigned() bool {
	for _, p := range c.subProps {
		if !p.IsUnassigned() {
			return false
		}
	}
	return true
}

// Complex property is always present.
func (c *complexProperty) IsPresent() bool {
	return true
}

// Return the boolean sub property whose metadata marked as 'exclusive' and has the value 'true'. If
// no such sub property, returns nil.
func (c *complexProperty) getExclusiveTrue() *booleanProperty {
	for _, p := range c.subProps {
		if p.Attribute().Type == TypeBoolean && p.Raw() == true {
			return p.(*booleanProperty)
		}
	}
	return nil
}

// Return the sub property by the name (case insensitive)
func (c *complexProperty) getSubProperty(name string) (Property, error) {
	p, ok := c.subProps[strings.ToLower(name)]
	if !ok {
		return nil, Errors.noTarget(fmt.Sprintf("%s.%s does not yield a target", c.attr.DisplayName(), name))
	}
	return p, nil
}

// Return a list of sub properties that is marked as 'identity' in metadata. If there is no such marked sub property,
// Return all sub properties.
func (c *complexProperty) getIdentityProps() []Property {
	identityProps := make([]Property, 0, len(c.subProps))
	for _, p := range c.subProps {
		if p.Attribute() != nil && p.Attribute().Metadata != nil && p.Attribute().Metadata.IsIdentity {
			identityProps = append(identityProps, p)
		}
	}

	if len(identityProps) == 0 {
		for _, p := range c.subProps {
			identityProps = append(identityProps, p)
		}
	}

	return identityProps
}

// A SCIM multiValued property. The Raw() value call returns a slice of the non-unassigned member property's
// Raw() value call result. Since this slice is computed on the fly, this call might be expensive.
type multiValuedProperty struct {
	attr  *Attribute
	props []Property

	// reference to the exclusively true boolean property
	// from the complex property elements, if any.
	excl *booleanProperty
}

func (m *multiValuedProperty) Attribute() *Attribute {
	return m.attr
}

func (m *multiValuedProperty) Raw() interface{} {
	values := make([]interface{}, 0, len(m.props))
	for _, prop := range m.props {
		if prop.IsUnassigned() {
			continue
		}
		values = append(values, prop.Raw())
	}
	return values
}

func (m *multiValuedProperty) Children() []Property {
	return m.props
}

func (m *multiValuedProperty) IsUnassigned() bool {
	return len(m.props) == 0
}

func (m *multiValuedProperty) IsPresent() bool {
	return m.props != nil
}

// implementation checks
var (
	_ Property = (*stringProperty)(nil)
	_ Property = (*integerProperty)(nil)
	_ Property = (*decimalProperty)(nil)
	_ Property = (*booleanProperty)(nil)
	_ Property = (*dateTimeProperty)(nil)
	_ Property = (*binaryProperty)(nil)
	_ Property = (*referenceProperty)(nil)
	_ Property = (*complexProperty)(nil)
	_ Property = (*multiValuedProperty)(nil)
)
