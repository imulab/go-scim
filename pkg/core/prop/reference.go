package prop

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/spec"
	"hash/fnv"
	"strings"
)

type referenceProperty struct {
	parent      Container
	attr        *spec.Attribute
	value       *string
	hash        uint64
	dirty       bool
	subscribers []Subscriber
}

func (p *referenceProperty) Attribute() *spec.Attribute {
	return p.attr
}

func (p *referenceProperty) Parent() Container {
	return p.parent
}

func (p *referenceProperty) Raw() interface{} {
	if p.value == nil {
		return nil
	}
	return *(p.value)
}

func (p *referenceProperty) IsUnassigned() bool {
	return p.value == nil
}

func (p *referenceProperty) Matches(another Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}

	if p.IsUnassigned() {
		return another.IsUnassigned()
	}

	return p.Hash() == another.Hash()
}

func (p *referenceProperty) Hash() uint64 {
	return p.hash
}

func (p *referenceProperty) computeHash() {
	if p.value == nil {
		p.hash = 0
	} else {
		h := fnv.New64a()
		_, err := h.Write([]byte(*(p.value)))
		if err != nil {
			panic("error computing hash")
		}
		p.hash = h.Sum64()
	}
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

func (p *referenceProperty) Add(value interface{}) error {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *referenceProperty) Replace(value interface{}) error {
	if value == nil {
		return p.Delete()
	}

	if s, ok := value.(string); !ok {
		return p.errIncompatibleValue(value)
	} else {
		p.dirty = true
		if eq, _ := p.EqualsTo(s); !eq {
			p.value = &s
			p.computeHash()
			if err := p.publish(EventAssigned); err != nil {
				return err
			}
		}
		return nil
	}
}

func (p *referenceProperty) Delete() error {
	p.dirty = true
	if p.value != nil {
		p.value = nil
		p.computeHash()
		if err := p.publish(EventUnassigned); err != nil {
			return err
		}
	}
	return nil
}

func (p *referenceProperty) publish(t EventType) error {
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

func (p *referenceProperty) Dirty() bool {
	return p.dirty
}

func (p *referenceProperty) Clone(parent Container) Property {
	c := &referenceProperty{
		parent:      parent,
		attr:        p.attr,
		value:       nil,
		hash:        p.hash,
		dirty:       p.dirty,
		subscribers: p.subscribers,
	}
	if p.value != nil {
		v := *(p.value)
		c.value = &v
	}
	return c
}

func (p *referenceProperty) Subscribe(subscriber Subscriber) {
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *referenceProperty) String() string {
	return fmt.Sprintf("[%s] %v", p.attr.String(), p.Raw())
}

func (p *referenceProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("%v is incompatible with attribute '%s'", value, p.attr.Path())
}

func (p *referenceProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}

var (
	_ Property = (*referenceProperty)(nil)
)