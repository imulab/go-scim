package prop

import (
	"encoding/binary"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
	"hash/fnv"
	"strings"
)

// Create a new unassigned complex property. The method will panic if
// given attribute is not singular complex type.
func NewComplex(attr *core.Attribute, parent core.Container) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeComplex {
		panic("invalid attribute for complex property")
	}

	var (
		p = &complexProperty{
			attr:      attr,
			subProps:  make([]core.Property, 0, attr.CountSubAttributes()),
			nameIndex: make(map[string]int),
		}
	)
	{
		attr.ForEachSubAttribute(func(subAttribute *core.Attribute) {
			if subAttribute.MultiValued() {
				p.subProps = append(p.subProps, NewMulti(subAttribute, p))
			} else {
				switch subAttribute.Type() {
				case core.TypeString:
					p.subProps = append(p.subProps, NewString(subAttribute, p))
				case core.TypeInteger:
					p.subProps = append(p.subProps, NewInteger(subAttribute, p))
				case core.TypeDecimal:
					p.subProps = append(p.subProps, NewDecimal(subAttribute, p))
				case core.TypeBoolean:
					p.subProps = append(p.subProps, NewBoolean(subAttribute, p))
				case core.TypeDateTime:
					p.subProps = append(p.subProps, NewDateTime(subAttribute, p))
				case core.TypeReference:
					p.subProps = append(p.subProps, NewReference(subAttribute, p))
				case core.TypeBinary:
					p.subProps = append(p.subProps, NewBinary(subAttribute, p))
				case core.TypeComplex:
					p.subProps = append(p.subProps, NewComplex(subAttribute, p))
				default:
					panic("invalid type")
				}
			}
			p.nameIndex[strings.ToLower(subAttribute.Name())] = len(p.subProps) - 1
		})
	}

	return p
}

// Create a new complex property with given value. The method will panic if
// given attribute is not singular complex type. The property will be
// marked dirty at the start unless value is empty
func NewComplexOf(attr *core.Attribute, parent core.Container, value interface{}) core.Property {
	p := NewComplex(attr, parent)
	if err := p.Add(value); err != nil {
		panic(err)
	}
	return p
}

var (
	_ core.Property  = (*complexProperty)(nil)
	_ core.Container = (*complexProperty)(nil)
)

type complexProperty struct {
	parent      core.Container
	attr        *core.Attribute
	subProps    []core.Property // array of sub properties to maintain determinate iteration order
	nameIndex   map[string]int  // attribute's name (to lower case) to index in subProps to allow fast access
	subscribers []core.Subscriber
}

func (p *complexProperty) Parent() core.Container {
	return p.parent
}

func (p *complexProperty) Subscribe(subscriber core.Subscriber) {
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *complexProperty) Propagate(e *core.Event) error {
	if len(p.subscribers) > 0 {
		for _, subscriber := range p.subscribers {
			if err := subscriber.Notify(e); err != nil {
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

func (p *complexProperty) Attribute() *core.Attribute {
	return p.attr
}

// Caution: slow operation
func (p *complexProperty) Raw() interface{} {
	values := make(map[string]interface{})
	_ = p.ForEachChild(func(_ int, child core.Property) error {
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

func (p *complexProperty) Matches(another core.Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}

	// Usually this won't happen, but still check it to be sure.
	if p.CountChildren() != another.(core.Container).CountChildren() {
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
	err := p.ForEachChild(func(_ int, child core.Property) error {
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

func (p *complexProperty) Touched() bool {
	for _, subProp := range p.subProps {
		if subProp.Touched() {
			return true
		}
	}
	return false
}

func (p *complexProperty) CountChildren() int {
	return len(p.subProps)
}

func (p *complexProperty) ForEachChild(callback func(index int, child core.Property) error) error {
	for i, prop := range p.subProps {
		if err := callback(i, prop); err != nil {
			return err
		}
	}
	return nil
}

func (p *complexProperty) NewChild() int {
	return -1
}

func (p *complexProperty) ChildAtIndex(index interface{}) core.Property {
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

func (p *complexProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("value of type %T is incompatible with attribute '%s'", value, p.attr.Path())
}

func (p *complexProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}
