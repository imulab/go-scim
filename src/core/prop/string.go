package prop

import (
	"fmt"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
	"hash/fnv"
	"strings"
)

// Create a new unassigned string property. The method will panic if
// given attribute is not singular string type.
func NewString(attr *core.Attribute) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeString {
		panic("invalid attribute for string property")
	}
	return &stringProperty{
		attr:  attr,
		value: nil,
		dirty: false,
	}
}

// Create a new string property with given value. The method will panic if
// given attribute is not singular string type. The property will be
// marked dirty at the start.
func NewStringOf(attr *core.Attribute, value interface{}) core.Property {
	p := NewString(attr)
	_, err := p.Replace(value)
	if err != nil {
		panic(err)
	}
	return p
}

type stringProperty struct {
	attr  *core.Attribute
	value *string
	dirty bool
	hash  uint64
}

func (p *stringProperty) Attribute() *core.Attribute {
	return p.attr
}

func (p *stringProperty) Raw() interface{} {
	if p.value == nil {
		return nil
	}
	return *(p.value)
}

func (p *stringProperty) IsUnassigned() (unassigned bool, dirty bool) {
	return p.value == nil, p.dirty
}

func (p *stringProperty) CountChildren() int {
	return 0
}

func (p *stringProperty) ForEachChild(callback func(index int, child core.Property)) {}

func (p *stringProperty) Matches(another core.Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}

	if unassigned, _ := p.IsUnassigned(); unassigned {
		alsoUnassigned, _ := another.IsUnassigned()
		return alsoUnassigned
	}

	return p.Hash() == another.Hash()
}

func (p *stringProperty) Hash() uint64 {
	return p.hash
}

func (p *stringProperty) EqualsTo(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	if s, ok := value.(string); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		v1, v2 := p.formatCase(*(p.value)), p.formatCase(s)
		return strings.Compare(v1, v2) == 0, nil
	}
}

func (p *stringProperty) StartsWith(value string) (bool, error) {
	if p.value == nil {
		return false, nil
	}
	v1, v2 := p.formatCase(*(p.value)), p.formatCase(value)
	return strings.HasPrefix(v1, v2), nil
}

func (p *stringProperty) EndsWith(value string) (bool, error) {
	if p.value == nil {
		return false, nil
	}
	v1, v2 := p.formatCase(*(p.value)), p.formatCase(value)
	return strings.HasSuffix(v1, v2), nil
}

func (p *stringProperty) Contains(value string) (bool, error) {
	if p.value == nil {
		return false, nil
	}
	v1, v2 := p.formatCase(*(p.value)), p.formatCase(value)
	return strings.Contains(v1, v2), nil
}

func (p *stringProperty) GreaterThan(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	if s, ok := value.(string); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		v1, v2 := p.formatCase(*(p.value)), p.formatCase(s)
		return strings.Compare(v1, v2) > 0, nil
	}
}

func (p *stringProperty) LessThan(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	if s, ok := value.(string); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		v1, v2 := p.formatCase(*(p.value)), p.formatCase(s)
		return strings.Compare(v1, v2) < 0, nil
	}
}

func (p *stringProperty) Present() bool {
	return p.value != nil && len(*(p.value)) > 0
}

func (p *stringProperty) DFS(callback func(property core.Property)) {
	callback(p)
}

func (p *stringProperty) Add(value interface{}) (bool, error) {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *stringProperty) Replace(value interface{}) (bool, error) {
	if value == nil {
		return p.Delete()
	}

	if s, ok := value.(string); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		equal, _ := p.EqualsTo(s)
		if !equal {
			p.value = &s
			p.dirty = true
			p.computeHash()
		}
		return !equal, nil
	}
}

func (p *stringProperty) Delete() (bool, error) {
	present := p.Present()
	p.value = nil
	p.dirty = true
	p.computeHash()
	return present, nil
}

func (p *stringProperty) Compact() {}

func (p *stringProperty) String() string {
	return fmt.Sprintf("[%s] %v", p.attr.String(), p.Raw())
}

// Calculate the hash value of the property value
func (p *stringProperty) computeHash() {
	if p == nil {
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

// Return a case appropriate version of the given value, based on attribute's caseExact setting.
func (p *stringProperty) formatCase(value string) string {
	if p.attr.CaseExact() {
		return value
	} else {
		return strings.ToLower(value)
	}
}

func (p *stringProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("%v is incompatible with attribute '%s'", value, p.attr.Path())
}
