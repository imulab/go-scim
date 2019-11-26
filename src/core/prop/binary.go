package prop

import (
	"encoding/base64"
	"fmt"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
	"hash/fnv"
)

// Create a new unassigned binary property. The method will panic if
// given attribute is not singular binary type.
func NewBinary(attr *core.Attribute) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeBinary {
		panic("invalid attribute for binary property")
	}
	return &binaryProperty{
		attr:  attr,
		value: nil,
	}
}

// Create a new binary property with given base64 encoded value. The method will panic if
// given attribute is not singular binary type. The property will be
// marked dirty at the start.
func NewBinaryOf(attr *core.Attribute, value interface{}) core.Property {
	p := NewBinary(attr)
	_, err := p.Replace(value)
	if err != nil {
		panic(err)
	}
	return p
}

type binaryProperty struct {
	attr  *core.Attribute
	value []byte
	mod   int
	hash  uint64
}

func (p *binaryProperty) Attribute() *core.Attribute {
	return p.attr
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

func (p *binaryProperty) ModCount() int {
	return p.mod
}

func (p *binaryProperty) CountChildren() int {
	return 0
}

func (p *binaryProperty) ForEachChild(callback func(index int, child core.Property)) {}

func (p *binaryProperty) Matches(another core.Property) bool {
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

func (p *binaryProperty) DFS(callback func(property core.Property)) {
	callback(p)
}

func (p *binaryProperty) Add(value interface{}) (bool, error) {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *binaryProperty) Replace(value interface{}) (bool, error) {
	if value == nil {
		return p.Delete()
	}

	if s, ok := value.(string); !ok {
		return false, p.errIncompatibleValue(value)
	} else if b64, err := base64.RawStdEncoding.DecodeString(s); err != nil {
		return false, p.errIncompatibleValue(value)
	} else {
		equal := p.compareByteArray(p.value, b64)
		if !equal {
			copy(p.value, b64)
			p.computeHash()
			p.mod++
		}
		return !equal, nil
	}
}

func (p *binaryProperty) Delete() (bool, error) {
	present := p.Present()
	p.value = nil
	p.computeHash()
	if p.mod == 0 || present {
		p.mod++
	}
	return present, nil
}

func (p *binaryProperty) Compact() {}

func (p *binaryProperty) String() string {
	return fmt.Sprintf("[%s] len=%d", p.attr.String(), len(p.value))
}

func (p *binaryProperty) computeHash() {
	h := fnv.New64a()
	_, err := h.Write(p.value)
	if err != nil {
		panic("error computing hash")
	}
	p.hash = h.Sum64()
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
