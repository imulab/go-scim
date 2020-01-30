package prop

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// NewDecimal creates a new decimal property associated with attribute.
func NewDecimal(attr *spec.Attribute) Property {
	ensureSingularDecimalType(attr)
	p := decimalProperty{attr: attr, subscribers: []Subscriber{}}
	attr.ForEachAnnotation(func(annotation string, params map[string]interface{}) {
		if subscriber, ok := SubscriberFactory().Create(annotation, &p, params); ok {
			p.subscribers = append(p.subscribers, subscriber)
		}
	})
	return &p
}

// NewDecimalOf creates a new decimal property of given value associated with attribute.
func NewDecimalOf(attr *spec.Attribute, value float64) Property {
	p := NewDecimal(attr)
	_, err := p.Replace(value)
	if err != nil {
		panic(err)
	}
	return p
}

func ensureSingularDecimalType(attr *spec.Attribute) {
	if attr.MultiValued() || attr.Type() != spec.TypeDecimal {
		panic("invalid attribute for decimal property")
	}
}

type decimalProperty struct {
	attr        *spec.Attribute
	value       *float64
	dirty       bool
	subscribers []Subscriber
}

func (p *decimalProperty) Attribute() *spec.Attribute {
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

func (p *decimalProperty) Dirty() bool {
	return p.dirty
}

func (p *decimalProperty) Hash() uint64 {
	if p.value == nil {
		// This is a hash collision with the actual zero. But we circumvent this issue by checking unassigned
		// first before comparing hashes.
		return uint64(int64(0))
	} else {
		return uint64(*(p.value))
	}
}

func (p *decimalProperty) Matches(another Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}
	if p.IsUnassigned() {
		return another.IsUnassigned()
	}
	return p.Hash() == another.Hash()
}

func (p *decimalProperty) Clone() Property {
	c := decimalProperty{
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

func (p *decimalProperty) Add(value interface{}) (*Event, error) {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *decimalProperty) Replace(value interface{}) (*Event, error) {
	if value == nil {
		return p.Delete()
	}

	f64, err := p.tryFloat64(value)
	if err != nil {
		return nil, err
	}

	p.dirty = true
	if !p.EqualsTo(f64) {
		ev := Event{typ: EventAssigned, source: p, pre: p.Raw()}
		p.value = &f64
		return &ev, nil
	}

	return nil, nil
}

func (p *decimalProperty) Delete() (*Event, error) {
	p.dirty = true
	if p.value != nil {
		ev := Event{typ: EventUnassigned, source: p, pre: p.Raw()}
		p.value = nil
		return &ev, nil
	}
	return nil, nil
}

func (p *decimalProperty) Notify(events *Events) error {
	for _, sub := range p.subscribers {
		if err := sub.Notify(p, events); err != nil {
			return err
		}
	}
	return nil
}

func (p *decimalProperty) CountChildren() int {
	return 0
}

func (p *decimalProperty) ForEachChild(_ func(index int, child Property) error) error {
	return nil
}

func (p *decimalProperty) FindChild(_ func(child Property) bool) Property {
	return nil
}

func (p *decimalProperty) ChildAtIndex(_ interface{}) (Property, error) {
	return nil, nil
}

func (p *decimalProperty) tryFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("%w: value is incompatible with '%s'", spec.ErrInvalidValue, p.attr.Path())
	}
}

func (p *decimalProperty) EqualsTo(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	f64, err := p.tryFloat64(value)
	if err != nil {
		return false
	}

	return *(p.value) == f64
}

func (p *decimalProperty) GreaterThan(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	f64, err := p.tryFloat64(value)
	if err != nil {
		return false
	}

	return *(p.value) > f64
}

func (p *decimalProperty) GreaterThanOrEqualTo(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	f64, err := p.tryFloat64(value)
	if err != nil {
		return false
	}

	return *(p.value) >= f64
}

func (p *decimalProperty) LessThan(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	f64, err := p.tryFloat64(value)
	if err != nil {
		return false
	}

	return *(p.value) < f64
}

func (p *decimalProperty) LessThanOrEqualTo(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	f64, err := p.tryFloat64(value)
	if err != nil {
		return false
	}

	return *(p.value) <= f64
}

func (p *decimalProperty) Present() bool {
	return p.value != nil
}

var (
	_ EqCapable = (*decimalProperty)(nil)
	_ GtCapable = (*decimalProperty)(nil)
	_ GeCapable = (*decimalProperty)(nil)
	_ LtCapable = (*decimalProperty)(nil)
	_ LeCapable = (*decimalProperty)(nil)
	_ PrCapable = (*decimalProperty)(nil)
)
