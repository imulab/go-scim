package prop

import (
	"fmt"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
	"hash/fnv"
	"strings"
)

// Create a new unassigned reference property. The method will panic if
// given attribute is not singular reference type.
func NewReference(attr *core.Attribute, parent core.Container) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeReference {
		panic("invalid attribute for reference property")
	}
	return &referenceProperty{
		parent:      parent,
		attr:        attr,
		value:       nil,
		subscribers: []core.Subscriber{},
	}
}

// Create a new reference property with given value. The method will panic if
// given attribute is not singular reference type. The property will be
// marked dirty at the start.
func NewReferenceOf(attr *core.Attribute, parent core.Container, value interface{}) core.Property {
	p := NewReference(attr, parent)
	if err := p.Replace(value); err != nil {
		panic(err)
	}
	return p
}

var (
	_ core.Property = (*referenceProperty)(nil)
)

type referenceProperty struct {
	parent      core.Container
	attr        *core.Attribute
	value       *string
	hash        uint64
	touched     bool
	subscribers []core.Subscriber
}

func (p *referenceProperty) Parent() core.Container {
	return p.parent
}

func (p *referenceProperty) Subscribe(subscriber core.Subscriber) {
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *referenceProperty) Attribute() *core.Attribute {
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

func (p *referenceProperty) CountChildren() int {
	return 0
}

func (p *referenceProperty) ForEachChild(callback func(index int, child core.Property)) {}

func (p *referenceProperty) Matches(another core.Property) bool {
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
		p.value = &s
		p.touched = true
		return nil
	}
}

func (p *referenceProperty) Delete() error {
	p.value = nil
	p.touched = true
	p.computeHash()
	return nil
}

func (p *referenceProperty) Touched() bool {
	return p.touched
}

func (p *referenceProperty) Compact() {}

func (p *referenceProperty) String() string {
	return fmt.Sprintf("[%s] %v", p.attr.String(), p.Raw())
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

func (p *referenceProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("%v is incompatible with attribute '%s'", value, p.attr.Path())
}

func (p *referenceProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}
