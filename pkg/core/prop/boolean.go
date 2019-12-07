package prop

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/spec"
)

type booleanProperty struct {
	parent      Container
	attr        *spec.Attribute
	value       *bool
	dirty       bool
	subscribers []Subscriber
}

func (p *booleanProperty) Parent() Container {
	return p.parent
}

func (p *booleanProperty) Attribute() *spec.Attribute {
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

func (p *booleanProperty) ForEachChild(callback func(index int, child Property)) {}

func (p *booleanProperty) Matches(another Property) bool {
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
		p.dirty = true
		if eq, _ := p.EqualsTo(b); !eq {
			p.value = &b
			if err := p.publish(EventAssigned); err != nil {
				return err
			}
		}
		return nil
	}
}

func (p *booleanProperty) Delete() error {
	p.dirty = true
	if p.value != nil {
		p.value = nil
		if err := p.publish(EventUnassigned); err != nil {
			return err
		}
	}
	return nil
}

func (p *booleanProperty) publish(t EventType) error {
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

func (p *booleanProperty) Subscribe(subscriber Subscriber) {
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *booleanProperty) Dirty() bool {
	return p.dirty
}

func (p *booleanProperty) Clone(parent Container) Property {
	c := &booleanProperty{
		parent:      parent,
		attr:        p.attr,
		value:       nil,
		dirty:       p.dirty,
		subscribers: p.subscribers,
	}
	if p.value != nil {
		v := *(p.value)
		c.value = &v
	}
	return c
}

func (p *booleanProperty) String() string {
	return fmt.Sprintf("[%s] %v", p.attr.String(), p.Raw())
}

func (p *booleanProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("%v is incompatible with attribute '%s'", value, p.attr.Path())
}

func (p *booleanProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}

var (
	_ Property = (*booleanProperty)(nil)
)
