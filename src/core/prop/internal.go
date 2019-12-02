package prop

import "github.com/imulab/go-scim/src/core"

// Create a delegate property of p that invokes addInternal/replaceInternal/deleteInternal in place of
// Add/Replace/Delete methods, if they exist. If these methods do not, the method panics. These internal modification
// methods are intentionally not declared in the Property API because the user must be conscious that it is limited to
// internal only. All other property methods are directly delegated to the underlying property.
func Internal(p core.Property) core.Property {
	if t, ok := p.(interface{
		core.Property
		addInternal(value interface{}) error
		replaceInternal(value interface{}) error
		deleteInternal() error
	}); !ok {
		panic("target does not implement internal methods")
	} else {
		return &internal{p: t}
	}
}

type internal struct {
	p interface{
		core.Property
		addInternal(value interface{}) error
		replaceInternal(value interface{}) error
		deleteInternal() error
	}
}

func (i *internal) Attribute() *core.Attribute {
	return i.p.Attribute()
}

func (i *internal) Raw() interface{} {
	return i.p.Raw()
}

func (i *internal) IsUnassigned() bool {
	return i.p.IsUnassigned()
}

func (i *internal) ModCount() int {
	return i.p.ModCount()
}

func (i *internal) Matches(another core.Property) bool {
	return i.p.Matches(another)
}

func (i *internal) Hash() uint64 {
	return i.p.Hash()
}

func (i *internal) EqualsTo(value interface{}) (bool, error) {
	return i.p.EqualsTo(value)
}

func (i *internal) StartsWith(value string) (bool, error) {
	return i.p.StartsWith(value)
}

func (i *internal) EndsWith(value string) (bool, error) {
	return i.p.EndsWith(value)
}

func (i *internal) Contains(value string) (bool, error) {
	return i.p.Contains(value)
}

func (i *internal) GreaterThan(value interface{}) (bool, error) {
	return i.p.GreaterThan(value)
}

func (i *internal) LessThan(value interface{}) (bool, error) {
	return i.p.LessThan(value)
}

func (i *internal) Present() bool {
	return i.p.Present()
}

func (i *internal) Add(value interface{}) (bool, error) {
	return false, i.p.addInternal(value)
}

func (i *internal) Replace(value interface{}) (bool, error) {
	return false, i.p.replaceInternal(value)
}

func (i *internal) Delete() (bool, error) {
	return false, i.p.deleteInternal()
}
