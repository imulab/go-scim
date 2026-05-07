package prop

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// NewBoolean creates a new boolean property associated with attribute.
func NewBoolean(attr *spec.Attribute) Property {
	ensureSingularBooleanType(attr)
	p := booleanProperty{attr: attr, subscribers: []Subscriber{}}
	attr.ForEachAnnotation(func(annotation string, params map[string]interface{}) {
		if subscriber, ok := SubscriberFactory().Create(annotation, &p, params); ok {
			p.subscribers = append(p.subscribers, subscriber)
		}
	})
	return &p
}

// NewBooleanOf creates a new boolean property of given value associated with attribute.
func NewBooleanOf(attr *spec.Attribute, value bool) Property {
	p := NewBoolean(attr)
	_, err := p.Replace(value)
	if err != nil {
		panic(err)
	}
	return p
}

func ensureSingularBooleanType(attr *spec.Attribute) {
	if attr.MultiValued() || attr.Type() != spec.TypeBoolean {
		panic("invalid attribute for boolean property")
	}
}

type booleanProperty struct {
	attr        *spec.Attribute
	value       *bool
	dirty       bool
	subscribers []Subscriber
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

func (p *booleanProperty) Dirty() bool {
	return p.dirty
}

func (p *booleanProperty) Hash() uint64 {
	if p.value == nil {
		return uint64(0)
	}

	if *(p.value) {
		return uint64(1)
	} else {
		return uint64(2)
	}
}

func (p *booleanProperty) Matches(another Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}
	if p.IsUnassigned() {
		return another.IsUnassigned()
	}
	return p.Hash() == another.Hash()
}

func (p *booleanProperty) Clone() Property {
	c := booleanProperty{
		attr:        p.attr,
		value:       nil,
		dirty:       p.dirty,
		subscribers: p.subscribers,
	}
	if p.value != nil {
		v := *(p.value)
		c.value = &v
	}
	return &c
}

func (p *booleanProperty) Add(value interface{}) (*Event, error) {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *booleanProperty) Replace(value interface{}) (*Event, error) {
	if value == nil {
		return p.Delete()
	}

	b, ok := value.(bool)
	if !ok {
		return nil, fmt.Errorf("%w: value is incompatible with '%s'", spec.ErrInvalidValue, p.attr.Path())
	}

	p.dirty = true
	if !p.EqualsTo(b) {
		ev := Event{typ: EventAssigned, source: p, pre: p.Raw()}
		p.value = &b
		return &ev, nil
	}

	return nil, nil
}

func (p *booleanProperty) Delete() (*Event, error) {
	p.dirty = true
	if p.value != nil {
		ev := Event{typ: EventUnassigned, source: p, pre: p.Raw()}
		p.value = nil
		return &ev, nil
	}
	return nil, nil
}

func (p *booleanProperty) Notify(events *Events) error {
	for _, sub := range p.subscribers {
		if err := sub.Notify(p, events); err != nil {
			return err
		}
	}
	return nil
}

func (p *booleanProperty) CountChildren() int {
	return 0
}

func (p *booleanProperty) ForEachChild(_ func(index int, child Property) error) error {
	return nil
}

func (p *booleanProperty) FindChild(_ func(child Property) bool) Property {
	return nil
}

func (p *booleanProperty) ChildAtIndex(_ interface{}) (Property, error) {
	return nil, nil
}

func (p *booleanProperty) EqualsTo(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	if b, ok := value.(bool); !ok {
		return false
	} else {
		return *(p.value) == b
	}
}

func (p *booleanProperty) Present() bool {
	return p.value != nil
}

var (
	_ EqCapable = (*booleanProperty)(nil)
	_ PrCapable = (*booleanProperty)(nil)
)
