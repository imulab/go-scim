package prop

import (
	"fmt"
	"github.com/imulab/go-scim/core/errors"
	"github.com/imulab/go-scim/core/spec"
	"hash/fnv"
	"strings"
)

type stringProperty struct {
	parent      Container
	attr        *spec.Attribute
	value       *string
	hash        uint64
	dirty       bool
	subscribers []Subscriber
}

func (p *stringProperty) Attribute() *spec.Attribute {
	return p.attr
}

func (p *stringProperty) Parent() Container {
	return p.parent
}

func (p *stringProperty) Raw() interface{} {
	if p.value == nil {
		return nil
	}
	return *(p.value)
}

func (p *stringProperty) IsUnassigned() bool {
	return p.value == nil
}

func (p *stringProperty) Matches(another Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}
	if p.IsUnassigned() {
		return another.IsUnassigned()
	}
	return p.Hash() == another.Hash()
}

func (p *stringProperty) Hash() uint64 {
	return p.hash
}

func (p *stringProperty) computeHash() {
	if p.value == nil {
		p.hash = 0
	} else {
		h := fnv.New64a()
		_, err := h.Write([]byte(p.formatCase(*(p.value))))
		if err != nil {
			panic("error computing hash")
		}
		p.hash = h.Sum64()
	}
}

func (p *stringProperty) EqualsTo(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	if s, ok := value.(string); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		v1, v2 := p.formatCase(*(p.value)), p.formatCase(s)
		return strings.Compare(v1, v2) == 0, nil
	}
}

func (p *stringProperty) StartsWith(value string) (bool, error) {
	if p.value == nil {
		return false, nil
	}
	v1, v2 := p.formatCase(*(p.value)), p.formatCase(value)
	return strings.HasPrefix(v1, v2), nil
}

func (p *stringProperty) EndsWith(value string) (bool, error) {
	if p.value == nil {
		return false, nil
	}
	v1, v2 := p.formatCase(*(p.value)), p.formatCase(value)
	return strings.HasSuffix(v1, v2), nil
}

func (p *stringProperty) Contains(value string) (bool, error) {
	if p.value == nil {
		return false, nil
	}
	v1, v2 := p.formatCase(*(p.value)), p.formatCase(value)
	return strings.Contains(v1, v2), nil
}

func (p *stringProperty) GreaterThan(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	if s, ok := value.(string); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		v1, v2 := p.formatCase(*(p.value)), p.formatCase(s)
		return strings.Compare(v1, v2) > 0, nil
	}
}

func (p *stringProperty) LessThan(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	if s, ok := value.(string); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		v1, v2 := p.formatCase(*(p.value)), p.formatCase(s)
		return strings.Compare(v1, v2) < 0, nil
	}
}

func (p *stringProperty) Present() bool {
	return p.value != nil && len(*(p.value)) > 0
}

func (p *stringProperty) Add(value interface{}) error {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *stringProperty) Replace(value interface{}) error {
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

func (p *stringProperty) Delete() error {
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

func (p *stringProperty) publish(t EventType) error {
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

func (p *stringProperty) Dirty() bool {
	return p.dirty
}

func (p *stringProperty) Subscribe(subscriber Subscriber) {
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *stringProperty) Clone(parent Container) Property {
	c := &stringProperty{
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

func (p *stringProperty) String() string {
	return fmt.Sprintf("[%s] %v", p.attr.String(), p.Raw())
}

// Return a case appropriate version of the given value, based on attribute's caseExact setting.
func (p *stringProperty) formatCase(value string) string {
	if p.attr.CaseExact() {
		return value
	} else {
		return strings.ToLower(value)
	}
}

func (p *stringProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("%v is incompatible with attribute '%s'", value, p.attr.Path())
}

var (
	_ Property = (*stringProperty)(nil)
)
