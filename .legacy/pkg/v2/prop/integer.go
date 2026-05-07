package prop

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// NewInteger creates a new integer property associated with attribute.
func NewInteger(attr *spec.Attribute) Property {
	ensureSingularIntegerType(attr)
	p := integerProperty{attr: attr, subscribers: []Subscriber{}}
	attr.ForEachAnnotation(func(annotation string, params map[string]interface{}) {
		if subscriber, ok := SubscriberFactory().Create(annotation, &p, params); ok {
			p.subscribers = append(p.subscribers, subscriber)
		}
	})
	return &p
}

// NewIntegerOf creates a new integer property of given value associated with attribute.
func NewIntegerOf(attr *spec.Attribute, value int64) Property {
	p := NewInteger(attr)
	_, err := p.Replace(value)
	if err != nil {
		panic(err)
	}
	return p
}

func ensureSingularIntegerType(attr *spec.Attribute) {
	if attr.MultiValued() || attr.Type() != spec.TypeInteger {
		panic("invalid attribute for integer property")
	}
}

type integerProperty struct {
	attr        *spec.Attribute
	value       *int64
	dirty       bool
	subscribers []Subscriber
}

func (p *integerProperty) Attribute() *spec.Attribute {
	return p.attr
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

func (p *integerProperty) Dirty() bool {
	return p.dirty
}

func (p *integerProperty) Hash() uint64 {
	if p.value == nil {
		// This will be hash collision, but we can circumvent this issue by checking the unassigned case first
		// before comparing value hashes
		return uint64(int64(0))
	} else {
		return uint64(*(p.value))
	}
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

func (p *integerProperty) Clone() Property {
	c := integerProperty{
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

func (p *integerProperty) Add(value interface{}) (*Event, error) {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *integerProperty) Replace(value interface{}) (*Event, error) {
	if value == nil {
		return p.Delete()
	}

	i64, err := p.tryInt64(value)
	if err != nil {
		return nil, err
	}

	p.dirty = true
	if !p.EqualsTo(i64) {
		ev := Event{typ: EventAssigned, source: p, pre: p.Raw()}
		p.value = &i64
		return &ev, nil
	}

	return nil, nil
}

func (p *integerProperty) Delete() (*Event, error) {
	p.dirty = true
	if p.value != nil {
		ev := Event{typ: EventUnassigned, source: p, pre: p.Raw()}
		p.value = nil
		return &ev, nil
	}
	return nil, nil
}

func (p *integerProperty) Notify(events *Events) error {
	for _, sub := range p.subscribers {
		if err := sub.Notify(p, events); err != nil {
			return err
		}
	}
	return nil
}

func (p *integerProperty) CountChildren() int {
	return 0
}

func (p *integerProperty) ForEachChild(_ func(index int, child Property) error) error {
	return nil
}

func (p *integerProperty) FindChild(_ func(child Property) bool) Property {
	return nil
}

func (p *integerProperty) ChildAtIndex(_ interface{}) (Property, error) {
	return nil, nil
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
		return 0, fmt.Errorf("%w: value is incompatible with '%s'", spec.ErrInvalidValue, p.attr.Path())
	}
}

func (p *integerProperty) EqualsTo(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	i64, err := p.tryInt64(value)
	if err != nil {
		return false
	}

	return *(p.value) == i64
}

func (p *integerProperty) GreaterThan(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	i64, err := p.tryInt64(value)
	if err != nil {
		return false
	}

	return *(p.value) > i64
}

func (p *integerProperty) GreaterThanOrEqualTo(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	i64, err := p.tryInt64(value)
	if err != nil {
		return false
	}

	return *(p.value) >= i64
}

func (p *integerProperty) LessThan(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	i64, err := p.tryInt64(value)
	if err != nil {
		return false
	}

	return *(p.value) < i64
}

func (p *integerProperty) LessThanOrEqualTo(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	i64, err := p.tryInt64(value)
	if err != nil {
		return false
	}

	return *(p.value) <= i64
}

func (p *integerProperty) Present() bool {
	return p.value != nil
}

var (
	_ EqCapable = (*integerProperty)(nil)
	_ GtCapable = (*integerProperty)(nil)
	_ GeCapable = (*integerProperty)(nil)
	_ LtCapable = (*integerProperty)(nil)
	_ LeCapable = (*integerProperty)(nil)
	_ PrCapable = (*integerProperty)(nil)
)
