package prop

import (
	"encoding/binary"
	"github.com/imulab/go-scim/core/errors"
	"github.com/imulab/go-scim/core/spec"
	"hash/fnv"
	"strings"
)

type complexProperty struct {
	parent      Container
	attr        *spec.Attribute
	subProps    []Property     // array of sub properties to maintain determinate iteration order
	nameIndex   map[string]int // attribute's name (to lower case) to index in subProps to allow fast access
	subscribers []Subscriber
}

func (p *complexProperty) Attribute() *spec.Attribute {
	return p.attr
}

func (p *complexProperty) Parent() Container {
	return p.parent
}

// Caution: slow operation
func (p *complexProperty) Raw() interface{} {
	values := make(map[string]interface{})
	_ = p.ForEachChild(func(_ int, child Property) error {
		values[child.Attribute().Name()] = child.Raw()
		return nil
	})
	return values
}

func (p *complexProperty) IsUnassigned() bool {
	for _, prop := range p.subProps {
		if !prop.IsUnassigned() {
			return false
		}
	}
	return true
}

func (p *complexProperty) Matches(another Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}

	// Usually this won't happen, but still check it to be sure.
	if p.CountChildren() != another.(Container).CountChildren() {
		return false
	}

	return p.Hash() == another.Hash()
}

func (p *complexProperty) Hash() uint64 {
	if p.IsUnassigned() {
		return 0
	}

	h := fnv.New64a()

	hasIdentity := p.attr.HasIdentitySubAttributes()
	err := p.ForEachChild(func(_ int, child Property) error {
		// Include fields in the computation if
		// - no sub attributes are marked as identity
		// - this sub attribute is marked identity
		if hasIdentity && !child.Attribute().IsIdentity() {
			return nil
		}

		_, err := h.Write([]byte(child.Attribute().Name()))
		if err != nil {
			return err
		}

		// Skip the value hash if it is unassigned
		if !child.IsUnassigned() {
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, child.Hash())
			_, err := h.Write(b)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		panic("error computing hash")
	}

	return h.Sum64()
}

func (p *complexProperty) EqualsTo(value interface{}) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *complexProperty) StartsWith(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *complexProperty) EndsWith(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *complexProperty) Contains(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *complexProperty) GreaterThan(value interface{}) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *complexProperty) LessThan(value interface{}) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *complexProperty) Present() bool {
	for _, prop := range p.subProps {
		if prop.Present() {
			return true
		}
	}
	return false
}

func (p *complexProperty) Add(value interface{}) error {
	if value == nil {
		return nil
	}

	if m, ok := value.(map[string]interface{}); !ok {
		return p.errIncompatibleValue(value)
	} else {
		for k, v := range m {
			i, ok := p.nameIndex[strings.ToLower(k)]
			if !ok {
				continue
			}
			if err := p.subProps[i].Add(v); err != nil {
				return err
			}
		}
		return nil
	}
}

func (p *complexProperty) Replace(value interface{}) (err error) {
	if value == nil {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			err = p.errIncompatibleValue(value)
		}
	}()

	err = p.Delete()
	if err != nil {
		return
	}

	err = p.Add(value)
	if err != nil {
		return
	}

	return
}

func (p *complexProperty) Delete() error {
	for _, subProp := range p.subProps {
		if err := subProp.Delete(); err != nil {
			return err
		}
	}
	return nil
}

func (p *complexProperty) Dirty() bool {
	for _, subProp := range p.subProps {
		if subProp.Dirty() {
			return true
		}
	}
	return false
}

func (p *complexProperty) Clone(parent Container) Property {
	c := &complexProperty{
		parent:      parent,
		attr:        p.attr,
		subProps:    make([]Property, 0, len(p.subProps)),
		nameIndex:   make(map[string]int),
		subscribers: p.subscribers,
	}
	for i, sp := range p.subProps {
		c.subProps = append(c.subProps, sp.Clone(c))
		c.nameIndex[strings.ToLower(sp.Attribute().Name())] = i
	}
	return c
}

func (p *complexProperty) Subscribe(subscriber Subscriber) {
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *complexProperty) CountChildren() int {
	return len(p.subProps)
}

func (p *complexProperty) ForEachChild(callback func(index int, child Property) error) error {
	for i, prop := range p.subProps {
		if err := callback(i, prop); err != nil {
			return err
		}
	}
	return nil
}

func (p *complexProperty) NewChild() int {
	return childNotCreated
}

func (p *complexProperty) ChildAtIndex(index interface{}) Property {
	name, ok := index.(string)
	if !ok {
		return nil
	}

	i, ok := p.nameIndex[strings.ToLower(name)]
	if !ok {
		return nil
	}

	return p.subProps[i]
}

func (p *complexProperty) Compact() {}

func (p *complexProperty) Propagate(e *Event) error {
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

func (p *complexProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("value of type %T is incompatible with attribute '%s'", value, p.attr.Path())
}

func (p *complexProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}

var (
	_ Property  = (*complexProperty)(nil)
	_ Container = (*complexProperty)(nil)
)
