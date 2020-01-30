package prop

import (
	"encoding/base64"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"hash/fnv"
)

// NewBinary creates a new binary property associated with attribute.
func NewBinary(attr *spec.Attribute) Property {
	ensureSingularBinaryType(attr)
	p := binaryProperty{attr: attr, subscribers: []Subscriber{}}
	attr.ForEachAnnotation(func(annotation string, params map[string]interface{}) {
		if subscriber, ok := SubscriberFactory().Create(annotation, &p, params); ok {
			p.subscribers = append(p.subscribers, subscriber)
		}
	})
	return &p
}

// NewBinaryOf creates a new binary property of given value associated with attribute.
func NewBinaryOf(attr *spec.Attribute, value string) Property {
	p := NewBinary(attr)
	_, err := p.Replace(value)
	if err != nil {
		panic(err)
	}
	return p
}

func ensureSingularBinaryType(attr *spec.Attribute) {
	if attr.MultiValued() || attr.Type() != spec.TypeBinary {
		panic("invalid attribute for binary property")
	}
}

type binaryProperty struct {
	attr        *spec.Attribute
	value       []byte
	hash        uint64
	dirty       bool
	subscribers []Subscriber
}

func (p *binaryProperty) Attribute() *spec.Attribute {
	return p.attr
}

func (p *binaryProperty) Raw() interface{} {
	if p.value == nil {
		return nil
	}
	return base64.StdEncoding.EncodeToString(p.value)
}

func (p *binaryProperty) IsUnassigned() bool {
	return len(p.value) == 0
}

func (p *binaryProperty) Dirty() bool {
	return p.dirty
}

func (p *binaryProperty) Hash() uint64 {
	return p.hash
}

func (p *binaryProperty) computeHash() {
	h := fnv.New64a()
	_, err := h.Write(p.value)
	if err != nil {
		panic("error computing hash")
	}
	p.hash = h.Sum64()
}

func (p *binaryProperty) Matches(another Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}
	if p.IsUnassigned() {
		return another.IsUnassigned()
	}
	return p.Hash() == another.Hash()
}

func (p *binaryProperty) Clone() Property {
	c := binaryProperty{
		attr:        p.attr,
		value:       make([]byte, len(p.value), len(p.value)),
		hash:        p.hash,
		dirty:       p.dirty,
		subscribers: p.subscribers,
	}
	copy(c.value, p.value)
	return &c
}

func (p *binaryProperty) Add(value interface{}) (*Event, error) {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *binaryProperty) Replace(value interface{}) (*Event, error) {
	if value == nil {
		return p.Delete()
	}

	s, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("%w: value is incompatible with '%s'", spec.ErrInvalidValue, p.attr.Path())
	}

	b64, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("%w: value for '%s' is not base64 encoded", spec.ErrInvalidValue, p.attr.Path())
	}

	p.dirty = true
	if p.byteArrayEquals(p.value, b64) {
		return nil, nil
	}

	ev := Event{typ: EventAssigned, source: p, pre: p.Raw()}
	p.value = b64
	p.computeHash()
	return &ev, nil
}

func (p *binaryProperty) Delete() (*Event, error) {
	p.dirty = true
	if len(p.value) == 0 {
		return nil, nil
	}

	ev := Event{typ: EventUnassigned, source: p, pre: p.Raw()}
	p.value = nil
	p.computeHash()
	return &ev, nil
}

func (p *binaryProperty) Notify(events *Events) error {
	for _, sub := range p.subscribers {
		if err := sub.Notify(p, events); err != nil {
			return err
		}
	}
	return nil
}

func (p *binaryProperty) CountChildren() int {
	return 0
}

func (p *binaryProperty) ForEachChild(_ func(index int, child Property) error) error {
	return nil
}

func (p *binaryProperty) FindChild(_ func(child Property) bool) Property {
	return nil
}

func (p *binaryProperty) ChildAtIndex(_ interface{}) (Property, error) {
	return nil, nil
}

func (p *binaryProperty) byteArrayEquals(b1 []byte, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	} else {
		for i := range b1 {
			if b1[i] != b2[i] {
				return false
			}
		}
		return true
	}
}

func (p *binaryProperty) EqualsTo(value interface{}) bool {
	if p.value == nil || value == nil {
		return false
	}

	s, ok := value.(string)
	if !ok {
		return false
	}

	b64, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return false
	}

	return p.byteArrayEquals(p.value, b64)
}

func (p *binaryProperty) Present() bool {
	return len(p.value) > 0
}

var (
	_ EqCapable = (*binaryProperty)(nil)
	_ PrCapable = (*binaryProperty)(nil)
)
