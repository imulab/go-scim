package prop

import (
	"fmt"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
)

// Create a new unassigned boolean property. The method will panic if
// given attribute is not singular boolean type.
func NewBoolean(attr *core.Attribute) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeBoolean {
		panic("invalid attribute for boolean property")
	}
	return &booleanProperty{
		attr:  attr,
		value: nil,
	}
}

// Create a new boolean property with given value. The method will panic if
// given attribute is not singular boolean type. The property will be
// marked dirty at the start.
func NewBooleanOf(attr *core.Attribute, value interface{}) core.Property {
	p := NewBoolean(attr)
	_, err := p.Replace(value)
	if err != nil {
		panic(err)
	}
	return p
}

type booleanProperty struct {
	attr  *core.Attribute
	value *bool
	mod   int
}

func (p *booleanProperty) Attribute() *core.Attribute {
	return p.attr
}

func (p *booleanProperty) Raw() interface{} {
	if p.value == nil {
		return nil
	}
	return *(p.value)
}

func (p *booleanProperty) IsUnassigned() bool {
	return p.value == nil
}

func (p *booleanProperty) ModCount() int {
	return p.mod
}

func (p *booleanProperty) CountChildren() int {
	return 0
}

func (p *booleanProperty) ForEachChild(callback func(index int, child core.Property)) {}

func (p *booleanProperty) Matches(another core.Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}

	if p.IsUnassigned() {
		return another.IsUnassigned()
	}

	return p.Hash() == another.Hash()
}

func (p *booleanProperty) Hash() uint64 {
	if p.value == nil {
		return uint64(0)
	} else {
		if *(p.value) {
			return uint64(1)
		} else {
			return uint64(2)
		}
	}
}

func (p *booleanProperty) EqualsTo(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	if b, ok := value.(bool); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		return *(p.value) == b, nil
	}
}

func (p *booleanProperty) StartsWith(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *booleanProperty) EndsWith(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *booleanProperty) Contains(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *booleanProperty) GreaterThan(value interface{}) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *booleanProperty) LessThan(value interface{}) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *booleanProperty) Present() bool {
	return p.value != nil
}

func (p *booleanProperty) Add(value interface{}) (bool, error) {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *booleanProperty) Replace(value interface{}) (bool, error) {
	if value == nil {
		return p.Delete()
	}

	if b, ok := value.(bool); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		equal, _ := p.EqualsTo(b)
		if !equal {
			p.value = &b
			p.mod++
		}
		return !equal, nil
	}
}

func (p *booleanProperty) Delete() (bool, error) {
	present := p.Present()
	p.value = nil
	if p.mod == 0 || present {
		p.mod++
	}
	return present, nil
}

func (p *booleanProperty) Compact() {}

func (p *booleanProperty) String() string {
	return fmt.Sprintf("[%s] %v", p.attr.String(), p.Raw())
}

func (p *booleanProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("%v is incompatible with attribute '%s'", value, p.attr.Path())
}

func (p *booleanProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}
