package prop

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"hash/fnv"
	"strings"
)

// NewString creates a new string property associated with attribute.
func NewString(attr *spec.Attribute) Property {
	ensureSingularStringType(attr)
	p := stringProperty{attr: attr, subscribers: []Subscriber{}}
	attr.ForEachAnnotation(func(annotation string, params map[string]interface{}) {
		if subscriber, ok := SubscriberFactory().Create(annotation, &p, params); ok {
			p.subscribers = append(p.subscribers, subscriber)
		}
	})
	return &p
}

// NewStringOf creates a new string property of given value associated with attribute.
func NewStringOf(attr *spec.Attribute, value string) Property {
	p := NewString(attr)
	_, err := p.Replace(value)
	if err != nil {
		panic(err)
	}
	return p
}

func ensureSingularStringType(attr *spec.Attribute) {
	if attr.MultiValued() || attr.Type() != spec.TypeString {
		panic("invalid attribute for string property")
	}
}

type stringProperty struct {
	attr        *spec.Attribute
	value       *string
	hash        uint64
	dirty       bool
	subscribers []Subscriber
}

func (p *stringProperty) Attribute() *spec.Attribute {
	return p.attr
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

func (p *stringProperty) Dirty() bool {
	return p.dirty
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

func (p *stringProperty) Matches(another Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}
	if p.IsUnassigned() {
		return another.IsUnassigned()
	}
	return p.Hash() == another.Hash()
}

func (p *stringProperty) Clone() Property {
	c := &stringProperty{
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

func (p *stringProperty) Add(value interface{}) (*Event, error) {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *stringProperty) Replace(value interface{}) (*Event, error) {
	if value == nil {
		return p.Delete()
	}

	if s, ok := value.(string); !ok {
		return nil, p.errIncompatibleValue()
	} else {
		p.dirty = true
		if !p.EqualsTo(s) {
			ev := Event{typ: EventAssigned, source: p, pre: p.Raw()}
			p.value = &s
			p.computeHash()
			return &ev, nil
		}
		return nil, nil
	}
}

func (p *stringProperty) Delete() (*Event, error) {
	p.dirty = true
	if p.value != nil {
		ev := Event{typ: EventUnassigned, source: p, pre: p.Raw()}
		p.value = nil
		p.computeHash()
		return &ev, nil
	}
	return nil, nil
}

func (p *stringProperty) Notify(events *Events) error {
	for _, sub := range p.subscribers {
		if err := sub.Notify(p, events); err != nil {
			return err
		}
	}
	return nil
}

func (p *stringProperty) CountChildren() int {
	return 0
}

func (p *stringProperty) ForEachChild(_ func(index int, child Property) error) error {
	return nil
}

func (p *stringProperty) FindChild(_ func(child Property) bool) Property {
	return nil
}

func (p *stringProperty) ChildAtIndex(_ interface{}) (Property, error) {
	return nil, nil
}

func (p *stringProperty) formatCase(value string) string {
	if p.attr.CaseExact() {
		return value
	} else {
		return strings.ToLower(value)
	}
}

func (p *stringProperty) errIncompatibleValue() error {
	return fmt.Errorf("%w: value incompatible with '%s'", spec.ErrInvalidValue, p.attr.Path())
}

func (p *stringProperty) EqualsTo(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	s, ok := value.(string)
	if !ok {
		return false
	}

	v1, v2 := p.formatCase(*(p.value)), p.formatCase(s)
	return strings.Compare(v1, v2) == 0
}

func (p *stringProperty) StartsWith(value string) bool {
	return p.value != nil && strings.HasPrefix(p.formatCase(*(p.value)), p.formatCase(value))
}

func (p *stringProperty) EndsWith(value string) bool {
	return p.value != nil && strings.HasSuffix(p.formatCase(*(p.value)), p.formatCase(value))
}

func (p *stringProperty) Contains(value string) bool {
	return p.value != nil && strings.Contains(p.formatCase(*(p.value)), p.formatCase(value))
}

func (p *stringProperty) GreaterThan(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	if s, ok := value.(string); !ok {
		return false
	} else {
		v1, v2 := p.formatCase(*(p.value)), p.formatCase(s)
		return strings.Compare(v1, v2) > 0
	}
}

func (p *stringProperty) GreaterThanOrEqualTo(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	if s, ok := value.(string); !ok {
		return false
	} else {
		v1, v2 := p.formatCase(*(p.value)), p.formatCase(s)
		return strings.Compare(v1, v2) >= 0
	}
}

func (p *stringProperty) LessThan(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	if s, ok := value.(string); !ok {
		return false
	} else {
		v1, v2 := p.formatCase(*(p.value)), p.formatCase(s)
		return strings.Compare(v1, v2) < 0
	}
}

func (p *stringProperty) LessThanOrEqualTo(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	if s, ok := value.(string); !ok {
		return false
	} else {
		v1, v2 := p.formatCase(*(p.value)), p.formatCase(s)
		return strings.Compare(v1, v2) <= 0
	}
}

func (p *stringProperty) Present() bool {
	return p.value != nil && len(*(p.value)) > 0
}

var (
	_ EqCapable = (*stringProperty)(nil)
	_ SwCapable = (*stringProperty)(nil)
	_ EwCapable = (*stringProperty)(nil)
	_ CoCapable = (*stringProperty)(nil)
	_ GtCapable = (*stringProperty)(nil)
	_ GeCapable = (*stringProperty)(nil)
	_ LtCapable = (*stringProperty)(nil)
	_ LeCapable = (*stringProperty)(nil)
	_ PrCapable = (*stringProperty)(nil)
)
