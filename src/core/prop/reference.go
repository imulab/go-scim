package prop

import (
	"fmt"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
	"strings"
)

// Create a new unassigned reference property. The method will panic if
// given attribute is not singular reference type.
func NewReference(attr *core.Attribute) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeReference {
		panic("invalid attribute for reference property")
	}
	return &referenceProperty{
		attr:  attr,
		value: nil,
		dirty: false,
	}
}

// Create a new reference property with given value. The method will panic if
// given attribute is not singular reference type. The property will be
// marked dirty at the start.
func NewReferenceOf(attr *core.Attribute, value string) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeReference {
		panic("invalid attribute for reference property")
	}
	return &referenceProperty{
		attr:  attr,
		value: &value,
		dirty: true,
	}
}

type referenceProperty struct {
	attr  *core.Attribute
	value *string
	dirty bool
}

func (p *referenceProperty) Attribute() *core.Attribute {
	return p.attr
}

func (p *referenceProperty) Raw() interface{} {
	if p.value == nil {
		return nil
	}
	return *(p.value)
}

func (p *referenceProperty) IsUnassigned() (unassigned bool, dirty bool) {
	return p.value == nil, p.dirty
}

func (p *referenceProperty) CountChildren() int {
	return 0
}

func (p *referenceProperty) ForEachChild(callback func(index int, child core.Property)) {}

func (p *referenceProperty) Matches(another core.Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}

	if unassigned, _ := p.IsUnassigned(); unassigned {
		alsoUnassigned, _ := another.IsUnassigned()
		return alsoUnassigned
	}

	ok, err := p.EqualsTo(another.Raw())
	return ok && err == nil
}

func (p *referenceProperty) EqualsTo(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	if s, ok := value.(string); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		return *(p.value) == s, nil
	}
}

func (p *referenceProperty) StartsWith(value string) (bool, error) {
	if p.value == nil {
		return false, nil
	}
	return strings.HasPrefix(*(p.value), value), nil
}

func (p *referenceProperty) EndsWith(value string) (bool, error) {
	if p.value == nil {
		return false, nil
	}
	return strings.HasSuffix(*(p.value), value), nil
}

func (p *referenceProperty) Contains(value string) (bool, error) {
	if p.value == nil {
		return false, nil
	}
	return strings.Contains(*(p.value), value), nil
}

func (p *referenceProperty) GreaterThan(value interface{}) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *referenceProperty) LessThan(value interface{}) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *referenceProperty) Present() bool {
	return p.value != nil && len(*(p.value)) > 0
}

func (p *referenceProperty) DFS(callback func(property core.Property)) {
	callback(p)
}

func (p *referenceProperty) Add(value interface{}) (bool, error) {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *referenceProperty) Replace(value interface{}) (bool, error) {
	if value == nil {
		return p.Delete()
	}

	if s, ok := value.(string); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		equal, _ := p.EqualsTo(s)
		if !equal {
			p.value = &s
			p.dirty = true
		}
		return !equal, nil
	}
}

func (p *referenceProperty) Delete() (bool, error) {
	present := p.Present()
	if present {
		p.value = nil
		p.dirty = true
	}
	return present, nil
}

func (p *referenceProperty) Compact() {}

func (p *referenceProperty) String() string {
	return fmt.Sprintf("[%s] %v", p.attr.String(), p.Raw())
}

func (p *referenceProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("%v is incompatible with attribute '%s'", value, p.attr.Path())
}

func (p *referenceProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}
