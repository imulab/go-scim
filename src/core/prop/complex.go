package prop

import (
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
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
				// todo
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
		unassigned, dirty := p.IsUnassigned()
		if !unassigned || dirty {
			panic("new complex property is supposed to be unassigned and non-dirty")
		}
	}

	return p
}

// Create a new complex property with given value. The method will panic if
// given attribute is not singular complex type. The property will be
// marked dirty at the start unless value is empty
func NewComplexOf(attr *core.Attribute, value map[string]interface{}) core.Property {
	p := NewComplex(attr)
	if len(value) > 0 {
		_, err := p.Replace(value)
		if err != nil {
			panic(err)
		}
	}
	return p
}

type complexProperty struct {
	attr      *core.Attribute
	subProps  []core.Property // array of sub properties to maintain determinate iteration order
	nameIndex map[string]int  // attribute's name (to lower case) to index in subProps to allow fast access
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

func (p *complexProperty) IsUnassigned() (unassigned bool, dirty bool) {
	for _, prop := range p.subProps {
		pUnassigned, pDirty := prop.IsUnassigned()
		unassigned = unassigned && pUnassigned
		dirty = dirty || pDirty
	}
	return
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

	q := another.(*complexProperty)

	// Attempt linear match before resorting to O(N^2) slow match
	for i := range p.subProps {
		if !p.subProps[i].Matches(q.subProps[i]) {
			goto SlowMatch
		}
	}
	return true

SlowMatch:
	for i := range p.subProps {
		match := false
		for j := range q.subProps {
			if p.subProps[i].Matches(q.subProps[j]) {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	return true
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

		return changed, nil
	}
}

func (p *complexProperty) Replace(value interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}

	if m, ok := value.(map[string]interface{}); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		changed := false

		for k, v := range m {
			i, ok := p.nameIndex[k]
			if !ok {
				continue
			}
			ok, err := p.subProps[i].Replace(v)
			if err != nil {
				return false, err
			}
			changed = changed || ok
		}

		return changed, nil
	}
}

func (p *complexProperty) Delete() (bool, error) {
	changed := false
	for i := range p.subProps {
		ok, err := p.subProps[i].Delete()
		if err != nil {
			return false, err
		}
		changed = changed || ok
	}
	return changed, nil
}

func (p *complexProperty) Compact() {}

func (p *complexProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("value of type %T is incompatible with attribute '%s'", value, p.attr.Path())
}

func (p *complexProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}
