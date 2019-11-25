package prop

import (
	"fmt"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
)

// Create a new unassigned integer property. The method will panic if
// given attribute is not singular integer type.
func NewInteger(attr *core.Attribute) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeInteger {
		panic("invalid attribute for integer property")
	}
	return &integerProperty{
		attr:  attr,
		value: nil,
		dirty: false,
	}
}

// Create a new integer property with given value. The method will panic if
// given attribute is not singular integer type. The property will be
// marked dirty at the start.
func NewIntegerOf(attr *core.Attribute, value int64) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeInteger {
		panic("invalid attribute for integer property")
	}
	return &integerProperty{
		attr:  attr,
		value: &value,
		dirty: true,
	}
}

type integerProperty struct {
	attr  *core.Attribute
	value *int64
	dirty bool
}

func (p *integerProperty) Attribute() *core.Attribute {
	return p.attr
}

func (p *integerProperty) Raw() interface{} {
	if p.value == nil {
		return nil
	}
	return *(p.value)
}

func (p *integerProperty) IsUnassigned() (unassigned bool, dirty bool) {
	return p.value == nil, p.dirty
}

func (p *integerProperty) CountChildren() int {
	return 0
}

func (p *integerProperty) ForEachChild(callback func(child core.Property)) {}

func (p *integerProperty) Matches(another core.Property) bool {
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

func (p *integerProperty) EqualsTo(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	i64, err := p.tryInt64(value)
	if err != nil {
		return false, err
	}

	return *(p.value) == i64, nil
}

func (p *integerProperty) StartsWith(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *integerProperty) EndsWith(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *integerProperty) Contains(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *integerProperty) GreaterThan(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	i64, err := p.tryInt64(value)
	if err != nil {
		return false, err
	}

	return *(p.value) > i64, nil
}

func (p *integerProperty) LessThan(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	i64, err := p.tryInt64(value)
	if err != nil {
		return false, err
	}

	return *(p.value) < i64, nil
}

func (p *integerProperty) Present() bool {
	return p.value != nil
}

func (p *integerProperty) DFS(callback func(property core.Property)) {
	callback(p)
}

func (p *integerProperty) Add(value interface{}) (bool, error) {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *integerProperty) Replace(value interface{}) (bool, error) {
	if value == nil {
		return p.Delete()
	}

	if i64, err := p.tryInt64(value); err != nil {
		return false, err
	} else {
		equal, _ := p.EqualsTo(i64)
		if !equal {
			p.value = &i64
			p.dirty = true
		}
		return !equal, nil
	}
}

func (p *integerProperty) Delete() (bool, error) {
	present := p.Present()
	if present {
		p.value = nil
		p.dirty = true
	}
	return present, nil
}

func (p *integerProperty) Compact() {}

func (p *integerProperty) String() string {
	return fmt.Sprintf("[%s] %v", p.attr.String(), p.Raw())
}

func (p *integerProperty) tryInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint:
		return int64(v), nil
	default:
		return 0, errors.InvalidValue("'%v' is incompatible with integer property '%s'", value, p.attr.Path())
	}
}

func (p *integerProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}
