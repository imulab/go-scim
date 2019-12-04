package prop

import (
	"fmt"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
)

// Create a new unassigned boolean property. The method will panic if
// given attribute is not singular boolean type.
func NewBoolean(attr *core.Attribute, parent core.Container) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeBoolean {
		panic("invalid attribute for boolean property")
	}
	p := &booleanProperty{
		parent:      parent,
		attr:        attr,
		value:       nil,
		subscribers: []core.Subscriber{},
	}
	subscribeWithAnnotation(p)
	return p
}

// Create a new boolean property with given value. The method will panic if
// given attribute is not singular boolean type. The property will be
// marked dirty at the start.
func NewBooleanOf(attr *core.Attribute, parent core.Container, value interface{}) core.Property {
	p := NewBoolean(attr, parent)
	if err := p.Replace(value); err != nil {
		panic(err)
	}
	return p
}

var (
	_ core.Property = (*booleanProperty)(nil)
)

type booleanProperty struct {
	parent      core.Container
	attr        *core.Attribute
	value       *bool
	touched     bool
	subscribers []core.Subscriber
}

func (p *booleanProperty) Parent() core.Container {
	return p.parent
}

func (p *booleanProperty) Subscribe(subscriber core.Subscriber) {
	p.subscribers = append(p.subscribers, subscriber)
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

func (p *booleanProperty) Add(value interface{}) error {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *booleanProperty) Replace(value interface{}) error {
	if value == nil {
		return p.Delete()
	}

	if b, ok := value.(bool); !ok {
		return p.errIncompatibleValue(value)
	} else {
		p.touched = true
		if eq, _ := p.EqualsTo(b); !eq {
			p.value = &b
			if err := p.publish(core.EventAssigned); err != nil {
				return err
			}
		}
		return nil
	}
}

func (p *booleanProperty) Delete() error {
	p.touched = true
	if p.value != nil {
		p.value = nil
		if err := p.publish(core.EventUnassigned); err != nil {
			return err
		}
	}
	return nil
}

func (p *booleanProperty) publish(t core.EventType) error {
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

func (p *booleanProperty) Touched() bool {
	return p.touched
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
