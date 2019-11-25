package prop

import (
	"fmt"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
	"time"
)

// Create a new unassigned string property. The method will panic if
// given attribute is not singular dateTime type.
func NewDateTime(attr *core.Attribute) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeDateTime {
		panic("invalid attribute for dateTime property")
	}
	return &dateTimeProperty{
		attr:  attr,
		value: nil,
		dirty: false,
	}
}

// Create a new string property with given value. The method will panic if
// given attribute is not singular dateTime type or the value is not of ISO8601 format.
// The property will be marked dirty at the start.
func NewDateTimeOf(attr *core.Attribute, value string) core.Property {
	p := NewDateTime(attr)
	err := p.Replace(value)
	if err != nil {
		panic(err)
	}
	return p
}

type dateTimeProperty struct {
	attr	*core.Attribute
	value	*time.Time
	dirty	bool
}

func (p *dateTimeProperty) Attribute() *core.Attribute {
	return p.attr
}

func (p *dateTimeProperty) Raw() interface{} {
	if p.value == nil {
		return nil
	}
	return p.mustToISO8601()
}

func (p *dateTimeProperty) IsUnassigned() (unassigned bool, dirty bool) {
	return p.value == nil, p.dirty
}

func (p *dateTimeProperty) CountChildren() int {
	return 0
}

func (p *dateTimeProperty) ForEachChild(callback func(child core.Property)) {}

func (p *dateTimeProperty) Matches(another core.Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}
	ok, err := p.EqualsTo(another.Raw())
	return ok && err == nil
}

func (p *dateTimeProperty) EqualsTo(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	if s, ok := value.(string); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		t, err := p.fromISO8601(s)
		if err != nil {
			return false, err
		}
		return (*(p.value)).Equal(t), nil
	}
}

func (p *dateTimeProperty) StartsWith(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *dateTimeProperty) EndsWith(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *dateTimeProperty) Contains(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *dateTimeProperty) GreaterThan(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	if s, ok := value.(string); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		t, err := p.fromISO8601(s)
		if err != nil {
			return false, err
		}
		return (*(p.value)).After(t), nil
	}
}

func (p *dateTimeProperty) LessThan(value interface{}) (bool, error) {
	if p.value == nil || value == nil {
		return false, nil
	}

	if s, ok := value.(string); !ok {
		return false, p.errIncompatibleValue(value)
	} else {
		t, err := p.fromISO8601(s)
		if err != nil {
			return false, err
		}
		return (*(p.value)).Before(t), nil
	}
}

func (p *dateTimeProperty) Present() bool {
	return p.value != nil
}

func (p *dateTimeProperty) DFS(callback func(property core.Property)) {
	callback(p)
}

func (p *dateTimeProperty) Add(value interface{}) error {
	if value == nil {
		return p.Delete()
	}
	return p.Replace(value)
}

func (p *dateTimeProperty) Replace(value interface{}) error {
	if value == nil {
		return p.Delete()
	}

	if s, ok := value.(string); !ok {
		return p.errIncompatibleValue(value)
	} else if t, err := p.fromISO8601(s); err != nil {
		return err
	} else {
		p.value = &t
		p.dirty = true
		return nil
	}
}

func (p *dateTimeProperty) Delete() error {
	p.value = nil
	p.dirty = true
	return nil
}

func (p *dateTimeProperty) Compact() {}

func (p *dateTimeProperty) String() string {
	return fmt.Sprintf("[%s] %v", p.attr.String(), p.Raw())
}

func (p *dateTimeProperty) mustToISO8601() string {
	if p.value == nil {
		panic("do not call this method when value is nil")
	}
	return (*(p.value)).Format("2006-01-02T15:04:05")
}

func (p *dateTimeProperty) fromISO8601(value string) (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05", value)
	if err != nil {
		return time.Time{}, p.errIncompatibleValue(value)
	}
	return t, nil
}

func (p *dateTimeProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("'%v' is not in ISO8601 format required by dateTime property '%s'", value, p.attr.Path())
}

func (p *dateTimeProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}