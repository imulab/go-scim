package prop

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"time"
)

// NewDateTime creates a new dateTime property associated with attribute.
func NewDateTime(attr *spec.Attribute) Property {
	ensureSingularDateTimeType(attr)
	p := dateTimeProperty{attr: attr, subscribers: []Subscriber{}}
	attr.ForEachAnnotation(func(annotation string, params map[string]interface{}) {
		if subscriber, ok := SubscriberFactory().Create(annotation, &p, params); ok {
			p.subscribers = append(p.subscribers, subscriber)
		}
	})
	return &p
}

// NewDateTimeOf creates a new dateTime property of given value associated with attribute.
func NewDateTimeOf(attr *spec.Attribute, value string) Property {
	p := NewDateTime(attr)
	_, err := p.Replace(value)
	if err != nil {
		panic(err)
	}
	return p
}

func ensureSingularDateTimeType(attr *spec.Attribute) {
	if attr.MultiValued() || attr.Type() != spec.TypeDateTime {
		panic("invalid attribute for dateTime property")
	}
}

type dateTimeProperty struct {
	attr        *spec.Attribute
	value       *time.Time
	dirty       bool
	subscribers []Subscriber
}

func (p *dateTimeProperty) Attribute() *spec.Attribute {
	return p.attr
}

func (p *dateTimeProperty) Raw() interface{} {
	if p.value == nil {
		return nil
	}
	return p.mustToISO8601()
}

func (p *dateTimeProperty) IsUnassigned() bool {
	return p.value == nil
}

func (p *dateTimeProperty) Dirty() bool {
	return p.dirty
}

func (p *dateTimeProperty) Hash() uint64 {
	if p.value == nil {
		return uint64(int64(0))
	} else {
		return uint64((*(p.value)).Unix())
	}
}

func (p *dateTimeProperty) Matches(another Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}
	if p.IsUnassigned() {
		return another.IsUnassigned()
	}
	return p.Hash() == another.Hash()
}

func (p *dateTimeProperty) Clone() Property {
	c := &dateTimeProperty{
		attr:        p.attr,
		value:       nil,
		dirty:       p.dirty,
		subscribers: p.subscribers,
	}
	if p.value != nil {
		v, _ := time.Parse(spec.ISO8601, p.Raw().(string))
		c.value = &v
	}
	return c
}

func (p *dateTimeProperty) Add(value interface{}) (*Event, error) {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *dateTimeProperty) Replace(value interface{}) (*Event, error) {
	if value == nil {
		return p.Delete()
	}

	s, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("%w, value is incompatible with '%s'", spec.ErrInvalidValue, p.attr.Path())
	}

	t, err := p.fromISO8601(s)
	if err != nil {
		return nil, err
	}

	p.dirty = true
	if p.value != nil && (*(p.value)).Equal(t) {
		return nil, nil
	}

	ev := Event{typ: EventAssigned, source: p, pre: p.Raw()}
	p.value = &t
	return &ev, nil
}

func (p *dateTimeProperty) Delete() (*Event, error) {
	p.dirty = true
	if p.value == nil {
		return nil, nil
	}

	ev := Event{typ: EventUnassigned, source: p, pre: p.Raw()}
	p.value = nil
	return &ev, nil
}

func (p *dateTimeProperty) Notify(events *Events) error {
	for _, sub := range p.subscribers {
		if err := sub.Notify(p, events); err != nil {
			return err
		}
	}
	return nil
}

func (p *dateTimeProperty) CountChildren() int {
	return 0
}

func (p *dateTimeProperty) ForEachChild(_ func(index int, child Property) error) error {
	return nil
}

func (p *dateTimeProperty) FindChild(_ func(child Property) bool) Property {
	return nil
}

func (p *dateTimeProperty) ChildAtIndex(_ interface{}) (Property, error) {
	return nil, nil
}

func (p *dateTimeProperty) mustToISO8601() string {
	if p.value == nil {
		panic("do not call this method when value is nil")
	}
	return (*(p.value)).Format(spec.ISO8601)
}

func (p *dateTimeProperty) fromISO8601(value string) (time.Time, error) {
	t, err := time.Parse(spec.ISO8601, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w, value for '%s' does not conform to ISO8601", spec.ErrInvalidValue, p.attr.Path())
	}
	return t, nil
}

func (p *dateTimeProperty) EqualsTo(value interface{}) bool {
	return p.compareThisAndValue(value, func(this time.Time, that time.Time) bool {
		return this.Equal(that)
	})
}

func (p *dateTimeProperty) GreaterThan(value interface{}) bool {
	return p.compareThisAndValue(value, func(this time.Time, that time.Time) bool {
		return this.After(that)
	})
}

func (p *dateTimeProperty) GreaterThanOrEqualTo(value interface{}) bool {
	return p.compareThisAndValue(value, func(this time.Time, that time.Time) bool {
		return this.After(that) || this.Equal(that)
	})
}

func (p *dateTimeProperty) LessThan(value interface{}) bool {
	return p.compareThisAndValue(value, func(this time.Time, that time.Time) bool {
		return this.Before(that)
	})
}

func (p *dateTimeProperty) LessThanOrEqualTo(value interface{}) bool {
	return p.compareThisAndValue(value, func(this time.Time, that time.Time) bool {
		return this.Before(that) || this.Equal(that)
	})
}

func (p *dateTimeProperty) compareThisAndValue(value interface{}, comparator func(this time.Time, that time.Time) bool) bool {
	if p.value == nil || value == nil {
		return false
	}

	s, ok := value.(string)
	if !ok {
		return false
	}

	t, err := p.fromISO8601(s)
	if err != nil {
		return false
	}

	return comparator(*(p.value), t)
}

func (p *dateTimeProperty) Present() bool {
	return p.value != nil
}

var (
	_ EqCapable = (*dateTimeProperty)(nil)
	_ GtCapable = (*dateTimeProperty)(nil)
	_ GeCapable = (*dateTimeProperty)(nil)
	_ LtCapable = (*dateTimeProperty)(nil)
	_ LeCapable = (*dateTimeProperty)(nil)
	_ PrCapable = (*dateTimeProperty)(nil)
)
