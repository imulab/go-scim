package prop

import (
	"fmt"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
)

// Create a new unassigned decimal property. The method will panic if
// given attribute is not singular decimal type.
func NewDecimal(attr *core.Attribute, parent core.Container) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeDecimal {
		panic("invalid attribute for integer property")
	}
	p := &decimalProperty{
		parent:      parent,
		attr:        attr,
		value:       nil,
		subscribers: []core.Subscriber{},
	}
	subscribeWithAnnotation(p)
	return p
}

// Create a new decimal property with given value. The method will panic if
// given attribute is not singular decimal type. The property will be
// marked dirty at the start.
func NewDecimalOf(attr *core.Attribute, parent core.Container, value interface{}) core.Property {
	p := NewDecimal(attr, parent)
	if err := p.Replace(value); err != nil {
		panic(err)
	}
	return p
}

var (
	_ core.Property = (*decimalProperty)(nil)
)

type decimalProperty struct {
	parent      core.Container
	attr        *core.Attribute
	value       *float64
	touched     bool
	subscribers []core.Subscriber
}

func (p *decimalProperty) Attribute() *core.Attribute {
	return p.attr
}

func (p *decimalProperty) Parent() core.Container {
	return p.parent
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
		p.touched = true
		if eq, _ := p.EqualsTo(f64); !eq {
			p.value = &f64
			if err := p.publish(core.EventAssigned); err != nil {
				return err
			}
		}
		return nil
	}
}

func (p *decimalProperty) Delete() error {
	p.touched = true
	if p.value != nil {
		p.value = nil
		if err := p.publish(core.EventUnassigned); err != nil {
			return err
		}
	}
	return nil
}

func (p *decimalProperty) publish(t core.EventType) error {
	e := t.NewFrom(p)
	if len(p.subscribers) > 0 {
		for _, subscriber := range p.subscribers {
			if err := subscriber.Notify(p, e); err != nil {
				return err
			}
		}
	}
	if p.parent != nil && e.WillPropagate() {
		if err := p.parent.Propagate(e); err != nil {
			return err
		}
	}
	return nil
}

func (p *decimalProperty) Touched() bool {
	return p.touched
}

func (p *decimalProperty) Subscribe(subscriber core.Subscriber) {
	p.subscribers = append(p.subscribers, subscriber)
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
