package prop

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/spec"
)

type integerProperty struct {
	parent      Container
	attr        *spec.Attribute
	value       *int64
	touched     bool
	subscribers []Subscriber
}

func (p *integerProperty) Attribute() *spec.Attribute {
	return p.attr
}

func (p *integerProperty) Parent() Container {
	return p.parent
}

func (p *integerProperty) Raw() interface{} {
	if p.value == nil {
		return nil
	}
	return *(p.value)
}

func (p *integerProperty) IsUnassigned() bool {
	return p.value == nil
}

func (p *integerProperty) Matches(another Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}

	if p.IsUnassigned() {
		return another.IsUnassigned()
	}

	return p.Hash() == another.Hash()
}

func (p *integerProperty) Hash() uint64 {
	if p.value == nil {
		// This will be hash collision, but we are fine since
		// we will check the unassigned case first before comparing
		// value hashes
		return uint64(int64(0))
	} else {
		return uint64(*(p.value))
	}
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

func (p *integerProperty) Add(value interface{}) error {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *integerProperty) Replace(value interface{}) error {
	if value == nil {
		return p.Delete()
	}

	if i64, err := p.tryInt64(value); err != nil {
		return err
	} else {
		p.touched = true
		if eq, _ := p.EqualsTo(i64); !eq {
			p.value = &i64
			if err := p.publish(EventAssigned); err != nil {
				return err
			}
		}
		return nil
	}
}

func (p *integerProperty) Delete() error {
	p.touched = true
	if p.value != nil {
		p.value = nil
		if err := p.publish(EventUnassigned); err != nil {
			return err
		}
	}
	return nil
}

func (p *integerProperty) publish(t EventType) error {
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

func (p *integerProperty) Dirty() bool {
	return p.touched
}

func (p *integerProperty) Subscribe(subscriber Subscriber) {
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *integerProperty) Clone(parent Container) Property {
	c := &integerProperty{
		parent:      parent,
		attr:        p.attr,
		value:       nil,
		touched:     p.touched,
		subscribers: p.subscribers,
	}
	if p.value != nil {
		v := *(p.value)
		p.value = &v
	}
	return c
}

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

var (
	_ Property = (*integerProperty)(nil)
)
