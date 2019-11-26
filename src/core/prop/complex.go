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
func NewComplex(attr *core.Attribute) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeComplex {
		panic("invalid attribute for complex property")
	}

	var (
		subProps  = make([]core.Property, 0, attr.CountSubAttributes())
		nameIndex = make(map[string]int)
	)
	{
		attr.ForEachSubAttribute(func(subAttribute *core.Attribute) {
			if subAttribute.MultiValued() {
				subProps = append(subProps, NewMulti(subAttribute))
			} else {
				switch subAttribute.Type() {
				case core.TypeString:
					subProps = append(subProps, NewString(subAttribute))
				case core.TypeInteger:
					subProps = append(subProps, NewInteger(subAttribute))
				case core.TypeDecimal:
					subProps = append(subProps, NewDecimal(subAttribute))
				case core.TypeBoolean:
					subProps = append(subProps, NewBoolean(subAttribute))
				case core.TypeDateTime:
					subProps = append(subProps, NewDateTime(subAttribute))
				case core.TypeReference:
					subProps = append(subProps, NewReference(subAttribute))
				case core.TypeBinary:
					subProps = append(subProps, NewBinary(subAttribute))
				case core.TypeComplex:
					subProps = append(subProps, NewComplex(subAttribute))
				default:
					panic("invalid type")
				}
			}
			nameIndex[strings.ToLower(subAttribute.Name())] = len(subProps) - 1
		})
	}

	var p core.Property
	{
		p = &complexProperty{
			attr:      attr,
			subProps:  subProps,
			nameIndex: nameIndex,
		}
		if !p.IsUnassigned() || p.ModCount() != 0 {
			panic("new complex property is supposed to be unassigned and un-modded")
		}
	}

	return p
}

// Create a new complex property with given value. The method will panic if
// given attribute is not singular complex type. The property will be
// marked dirty at the start unless value is empty
func NewComplexOf(attr *core.Attribute, value interface{}) core.Property {
	p := NewComplex(attr)
	_, err := p.Add(value)
	if err != nil {
		panic(err)
	}
	return p
}

type complexProperty struct {
	attr      *core.Attribute
	subProps  []core.Property // array of sub properties to maintain determinate iteration order
	nameIndex map[string]int  // attribute's name (to lower case) to index in subProps to allow fast access
	hash      uint64
}

func (p *complexProperty) Attribute() *core.Attribute {
	return p.attr
}

// Caution: slow operation
func (p *complexProperty) Raw() interface{} {
	values := make(map[string]interface{})
	p.ForEachChild(func(_ int, child core.Property) {
		values[child.Attribute().Name()] = child.Raw()
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

func (p *complexProperty) ModCount() int {
	// Watch out for double counting
	n := 0
	p.ForEachChild(func(_ int, child core.Property) {
		n += child.ModCount()
	})
	return n
}

func (p *complexProperty) CountChildren() int {
	return len(p.subProps)
}

func (p *complexProperty) ForEachChild(callback func(index int, child core.Property)) {
	for i, prop := range p.subProps {
		callback(i, prop)
	}
}

func (p *complexProperty) Matches(another core.Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}

	// Usually this won't happen, but still check it to be sure.
	if p.CountChildren() != another.CountChildren() {
		return false
	}

	return p.Hash() == another.Hash()
}

func (p *complexProperty) Hash() uint64 {
	return p.hash
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

func (p *complexProperty) DFS(callback func(property core.Property)) {
	callback(p)
	p.ForEachChild(func(_ int, child core.Property) {
		callback(child)
	})
}

func (p *complexProperty) Add(value interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}

	if m, ok := value.(map[string]interface{}); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		changed := false
		for k, v := range m {
			i, ok := p.nameIndex[strings.ToLower(k)]
			if !ok {
				continue
			}
			ok, err := p.subProps[i].Add(v)
			if err != nil {
				return false, err
			}
			changed = changed || ok
		}
		p.computeHash()
		return changed, nil
	}
}

func (p *complexProperty) Replace(value interface{}) (changed bool, err error) {
	if value == nil {
		return false, nil
	}

	defer func() {
		if r := recover(); r != nil {
			err = p.errIncompatibleValue(value)
		}
	}()

	wip := NewComplexOf(p.attr, value)
	if p.Matches(wip) {
		changed = false
		return
	}

	changed = true
	p.subProps = wip.(*complexProperty).subProps
	p.nameIndex = wip.(*complexProperty).nameIndex
	p.hash = wip.(*complexProperty).hash
	wip = nil

	return
}

func (p *complexProperty) Delete() (bool, error) {
	present := p.Present()
	for i := range p.subProps {
		_, err := p.subProps[i].Delete()
		if err != nil {
			return false, err
		}
	}
	p.computeHash()
	return present, nil
}

func (p *complexProperty) Compact() {}

func (p *complexProperty) computeHash() {
	h := fnv.New64a()

	hasIdentity := p.attr.HasIdentitySubAttributes()
	p.ForEachChild(func(_ int, child core.Property) {
		// Include fields in the computation if
		// - no sub attributes are marked as identity
		// - this sub attribute is marked identity
		if hasIdentity && !child.Attribute().IsIdentity() {
			return
		}

		_, err := h.Write([]byte(child.Attribute().Name()))
		if err != nil {
			panic("error computing hash")
		}

		// Skip the value hash if it is unassigned
		if !child.IsUnassigned() {
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, child.Hash())
			_, err := h.Write(b)
			if err != nil {
				panic("error computing hash")
			}
		}
	})
	p.hash = h.Sum64()
}

func (p *complexProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("value of type %T is incompatible with attribute '%s'", value, p.attr.Path())
}

func (p *complexProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}
