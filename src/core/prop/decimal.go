package prop

import (
	"fmt"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
)

// Create a new unassigned decimal property. The method will panic if
// given attribute is not singular decimal type.
func NewDecimal(attr *core.Attribute) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeDecimal {
		panic("invalid attribute for integer property")
	}
	return &decimalProperty{
		attr:  attr,
		value: nil,
	}
}

// Create a new decimal property with given value. The method will panic if
// given attribute is not singular decimal type. The property will be
// marked dirty at the start.
func NewDecimalOf(attr *core.Attribute, value interface{}) core.Property {
	p := NewDecimal(attr)
	if err := p.Replace(value); err != nil {
		panic(err)
	}
	return p
}

var (
	_ core.Property = (*decimalProperty)(nil)
)

type decimalProperty struct {
	attr    *core.Attribute
	value   *float64
	touched bool
}

func (p *decimalProperty) Attribute() *core.Attribute {
	return p.attr
}

func (p *decimalProperty) Raw() interface{} {
	if p.value == nil {
		return nil
	}
	return *(p.value)
}

func (p *decimalProperty) IsUnassigned() bool {
	return p.value == nil
}

func (p *decimalProperty) CountChildren() int {
	return 0
}

func (p *decimalProperty) ForEachChild(callback func(index int, child core.Property)) {}

func (p *decimalProperty) Matches(another core.Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}

	if p.IsUnassigned() {
		return another.IsUnassigned()
	}

	return p.Hash() == another.Hash()
}

func (p *decimalProperty) Hash() uint64 {
	if p == nil {
		// This is a hash collision with the actual zero. But we are fine
		// as we will check unassigned first before comparing hashes.
		return uint64(int64(0))
	} else {
		return uint64(*(p.value))
	}
}

func (p *decimalProperty) EqualsTo(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	f64, err := p.tryFloat64(value)
	if err != nil {
		return false, err
	}

	return *(p.value) == f64, nil
}

func (p *decimalProperty) StartsWith(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *decimalProperty) EndsWith(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *decimalProperty) Contains(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *decimalProperty) GreaterThan(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	f64, err := p.tryFloat64(value)
	if err != nil {
		return false, err
	}

	return *(p.value) > f64, nil
}

func (p *decimalProperty) LessThan(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	f64, err := p.tryFloat64(value)
	if err != nil {
		return false, err
	}

	return *(p.value) < f64, nil
}

func (p *decimalProperty) Present() bool {
	return p.value != nil
}

func (p *decimalProperty) Add(value interface{}) error {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *decimalProperty) Replace(value interface{}) error {
	if value == nil {
		return p.Delete()
	}

	if f64, err := p.tryFloat64(value); err != nil {
		return err
	} else {
		p.value = &f64
		p.touched = true
		return nil
	}
}

func (p *decimalProperty) Delete() error {
	p.value = nil
	p.touched = true
	return nil
}

func (p *decimalProperty) Touched() bool {
	return p.touched
}

func (p *decimalProperty) String() string {
	return fmt.Sprintf("[%s] %v", p.attr.String(), p.Raw())
}

func (p *decimalProperty) tryFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	default:
		return 0, errors.InvalidValue("'%v' is incompatible with decimal property '%s'", value, p.attr.Path())
	}
}

func (p *decimalProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}
