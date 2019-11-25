package prop

import (
	"encoding/base64"
	"fmt"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
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
		dirty: false,
	}
}

// Create a new binary property with given base64 encoded value. The method will panic if
// given attribute is not singular binary type. The property will be
// marked dirty at the start.
func NewBinaryOf(attr *core.Attribute, value string) core.Property {
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
	dirty bool
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

func (p *binaryProperty) IsUnassigned() (unassigned bool, dirty bool) {
	return len(p.value) == 0, p.dirty
}

func (p *binaryProperty) CountChildren() int {
	return 0
}

func (p *binaryProperty) ForEachChild(callback func(child core.Property)) {}

func (p *binaryProperty) Matches(another core.Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}

	if unassigned, _ := p.IsUnassigned(); unassigned {
		alsoUnassigned, _ := another.IsUnassigned()
		return alsoUnassigned
	}

	ok, err := p.EqualsTo(another.Raw())
	return ok && err == nil
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
			p.dirty = true
		}
		return !equal, nil
	}
}

func (p *binaryProperty) Delete() (bool, error) {
	present := p.Present()
	if present {
		p.value = nil
		p.dirty = true
	}
	return present, nil
}

func (p *binaryProperty) Compact() {}

func (p *binaryProperty) String() string {
	return fmt.Sprintf("[%s] len=%d", p.attr.String(), len(p.value))
}

func (p *binaryProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("%v is incompatible with attribute '%s'", value, p.attr.Path())
}

func (p *binaryProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}
