package prop

import (
	"encoding/base64"
	"fmt"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/spec"
	"hash/fnv"
)

type binaryProperty struct {
	parent      Container
	attr        *spec.Attribute
	value       []byte
	hash        uint64
	dirty       bool
	subscribers []Subscriber
}

func (p *binaryProperty) Clone(parent Container) Property {
	c := &binaryProperty{
		parent:      parent,
		attr:        p.attr,
		value:       make([]byte, len(p.value), len(p.value)),
		hash:        p.hash,
		dirty:       p.dirty,
		subscribers: p.subscribers,
	}
	copy(c.value, p.value)
	return c
}

func (p *binaryProperty) Attribute() *spec.Attribute {
	return p.attr
}

func (p *binaryProperty) Parent() Container {
	return p.parent
}

func (p *binaryProperty) Raw() interface{} {
	if p.value == nil {
		return nil
	}
	return base64.RawStdEncoding.EncodeToString(p.value)
}

func (p *binaryProperty) IsUnassigned() bool {
	return len(p.value) == 0
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

func (p *binaryProperty) EqualsTo(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	if s, ok := value.(string); !ok {
		return false, p.errIncompatibleValue(value)
	} else if b64, err := base64.RawStdEncoding.DecodeString(s); err != nil {
		return false, p.errIncompatibleValue(value)
	} else {
		return p.compareByteArray(p.value, b64), nil
	}
}

func (p *binaryProperty) StartsWith(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *binaryProperty) EndsWith(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *binaryProperty) Contains(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *binaryProperty) GreaterThan(value interface{}) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *binaryProperty) LessThan(value interface{}) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *binaryProperty) Present() bool {
	return len(p.value) > 0
}

func (p *binaryProperty) Add(value interface{}) error {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *binaryProperty) Replace(value interface{}) error {
	if value == nil {
		return p.Delete()
	}

	if s, ok := value.(string); !ok {
		return p.errIncompatibleValue(value)
	} else if b64, err := base64.RawStdEncoding.DecodeString(s); err != nil {
		return p.errIncompatibleValue(value)
	} else {
		p.dirty = true
		if !p.compareByteArray(p.value, b64) {
			p.value = b64
			p.computeHash()
			if err := p.publish(EventAssigned); err != nil {
				return err
			}
		}
		return nil
	}
}

func (p *binaryProperty) Delete() error {
	p.dirty = true
	if len(p.value) > 0 {
		p.value = nil
		p.computeHash()
		if err := p.publish(EventUnassigned); err != nil {
			return err
		}
	}
	return nil
}

func (p *binaryProperty) publish(t EventType) error {
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

func (p *binaryProperty) Dirty() bool {
	return p.dirty
}

func (p *binaryProperty) Subscribe(subscriber Subscriber) {
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *binaryProperty) String() string {
	return fmt.Sprintf("[%s] len=%d", p.attr.String(), len(p.value))
}

func (p *binaryProperty) compareByteArray(b1 []byte, b2 []byte) bool {
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

func (p *binaryProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("%v is incompatible with attribute '%s'", value, p.attr.Path())
}

func (p *binaryProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}

var (
	_ Property = (*binaryProperty)(nil)
)
