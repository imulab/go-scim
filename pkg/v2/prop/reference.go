package prop

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"hash/fnv"
	"strings"
)

// NewReference creates a new reference property associated with attribute.
func NewReference(attr *spec.Attribute) Property {
	ensureSingularReferenceType(attr)
	p := referenceProperty{attr: attr, subscribers: []Subscriber{}}
	attr.ForEachAnnotation(func(annotation string, params map[string]interface{}) {
		if subscriber, ok := SubscriberFactory().Create(annotation, &p, params); ok {
			p.subscribers = append(p.subscribers, subscriber)
		}
	})
	return &p
}

// NewReferenceOf creates a new reference property of given value associated with attribute.
func NewReferenceOf(attr *spec.Attribute, value string) Property {
	p := NewReference(attr)
	_, err := p.Replace(value)
	if err != nil {
		panic(err)
	}
	return p
}

func ensureSingularReferenceType(attr *spec.Attribute) {
	if attr.MultiValued() || attr.Type() != spec.TypeReference {
		panic("invalid attribute for reference property")
	}
}

type referenceProperty struct {
	attr        *spec.Attribute
	value       *string
	hash        uint64
	dirty       bool
	subscribers []Subscriber
}

func (p *referenceProperty) Attribute() *spec.Attribute {
	return p.attr
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

func (p *referenceProperty) Dirty() bool {
	return p.dirty
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

func (p *referenceProperty) Matches(another Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}
	if p.IsUnassigned() {
		return another.IsUnassigned()
	}
	return p.Hash() == another.Hash()
}

func (p *referenceProperty) Clone() Property {
	c := referenceProperty{
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
	return &c
}

func (p *referenceProperty) Add(value interface{}) (*Event, error) {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *referenceProperty) Replace(value interface{}) (*Event, error) {
	if value == nil {
		return p.Delete()
	}

	s, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("%w: value is incompatible with '%s'", spec.ErrInvalidValue, p.attr.Path())
	}

	p.dirty = true
	if p.EqualsTo(s) {
		return nil, nil
	}

	ev := Event{typ: EventAssigned, source: p, pre: p.Raw()}
	p.value = &s
	p.computeHash()
	return &ev, nil
}

func (p *referenceProperty) Delete() (*Event, error) {
	p.dirty = true
	if p.value == nil {
		return nil, nil
	}

	ev := Event{typ: EventUnassigned, source: p, pre: p.Raw()}
	p.value = nil
	p.computeHash()
	return &ev, nil
}

func (p *referenceProperty) Notify(events *Events) error {
	for _, sub := range p.subscribers {
		if err := sub.Notify(p, events); err != nil {
			return err
		}
	}
	return nil
}

func (p *referenceProperty) CountChildren() int {
	return 0
}

func (p *referenceProperty) ForEachChild(_ func(index int, child Property) error) error {
	return nil
}

func (p *referenceProperty) FindChild(_ func(child Property) bool) Property {
	return nil
}

func (p *referenceProperty) ChildAtIndex(_ interface{}) (Property, error) {
	return nil, nil
}

func (p *referenceProperty) EqualsTo(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	s, ok := value.(string)
	if !ok {
		return false
	}

	return *(p.value) == s
}

func (p *referenceProperty) StartsWith(value string) bool {
	if p.value == nil {
		return false
	}
	return strings.HasPrefix(*(p.value), value)
}

func (p *referenceProperty) EndsWith(value string) bool {
	if p.value == nil {
		return false
	}
	return strings.HasSuffix(*(p.value), value)
}

func (p *referenceProperty) Contains(value string) bool {
	if p.value == nil {
		return false
	}
	return strings.Contains(*(p.value), value)
}

func (p *referenceProperty) Present() bool {
	return p.value != nil && len(*(p.value)) > 0
}

var (
	_ EqCapable = (*referenceProperty)(nil)
	_ SwCapable = (*referenceProperty)(nil)
	_ EwCapable = (*referenceProperty)(nil)
	_ CoCapable = (*referenceProperty)(nil)
	_ PrCapable = (*referenceProperty)(nil)
)
