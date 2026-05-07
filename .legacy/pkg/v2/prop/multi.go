package prop

import (
	"encoding/binary"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"hash/fnv"
)

// NewMulti creates a new multiValued property associated with attribute. All sub attributes are created.
func NewMulti(attr *spec.Attribute) Property {
	ensureMultiType(attr)
	p := multiValuedProperty{
		attr:        attr,
		elements:    []Property{},
		subscribers: []Subscriber{},
	}
	attr.ForEachAnnotation(func(annotation string, params map[string]interface{}) {
		if subscriber, ok := SubscriberFactory().Create(annotation, &p, params); ok {
			p.subscribers = append(p.subscribers, subscriber)
		}
	})
	return &p
}

// NewMultiOf creates a new multiValued property of given value associated with attribute.
func NewMultiOf(attr *spec.Attribute, value []interface{}) Property {
	p := NewMulti(attr)
	_, err := p.Replace(value)
	if err != nil {
		panic(err)
	}
	return p
}

func ensureMultiType(attr *spec.Attribute) {
	if !attr.MultiValued() {
		panic("invalid attribute for multi property")
	}
}

type multiValuedProperty struct {
	attr        *spec.Attribute
	dirty       bool
	elements    []Property
	subscribers []Subscriber
}

func (p *multiValuedProperty) Attribute() *spec.Attribute {
	return p.attr
}

// Caution: slow operation
func (p *multiValuedProperty) Raw() interface{} {
	if len(p.elements) == 0 {
		return nil
	}
	var values []interface{}
	for _, elem := range p.elements {
		values = append(values, elem.Raw())
	}
	return values
}

func (p *multiValuedProperty) IsUnassigned() bool {
	return len(p.elements) == 0
}

func (p *multiValuedProperty) Dirty() bool {
	return p.dirty
}

// Caution: expensive operation
func (p *multiValuedProperty) Hash() uint64 {
	if len(p.elements) == 0 {
		return 0
	}

	var hashes []uint64
	_ = p.ForEachChild(func(index int, child Property) error {
		if child.IsUnassigned() {
			return nil
		}

		// SCIM array does not have orders. We keep the hash array
		// sorted so that different multiValue properties containing
		// the same elements in different orders can be recognized as
		// the same, as they compute the same hash. We use insertion
		// sort here as we don't expect a large number of elements.
		hashes = append(hashes, child.Hash())
		for i := len(hashes) - 1; i > 0; i-- {
			if hashes[i-1] > hashes[i] {
				hashes[i-1], hashes[i] = hashes[i], hashes[i-1]
			}
		}
		return nil
	})

	h := fnv.New64a()
	for _, hash := range hashes {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, hash)
		_, err := h.Write(b)
		if err != nil {
			panic("error computing hash")
		}
	}

	return h.Sum64()
}

func (p *multiValuedProperty) Matches(another Property) bool {
	if !p.Attribute().Equals(another.Attribute()) {
		return false
	}
	if p.IsUnassigned() {
		return another.IsUnassigned()
	}
	return p.Hash() == another.Hash()
}

func (p *multiValuedProperty) Clone() Property {
	c := multiValuedProperty{
		attr:        p.attr,
		elements:    []Property{},
		dirty:       p.dirty,
		subscribers: p.subscribers,
	}
	for _, elem := range p.elements {
		c.elements = append(c.elements, elem.Clone())
	}
	return &c
}

func (p *multiValuedProperty) Add(value interface{}) (*Event, error) {
	if value == nil {
		return nil, nil
	}

	// transform value into properties to add
	var (
		toAdd = make([]Property, 0)
		p0    Property
		err   error
	)
	{
		switch val := value.(type) {
		case []interface{}:
			for _, v := range val {
				if v == nil {
					continue
				}
				p0, err = p.newElementProperty(v)
				if err != nil {
					return nil, err
				}
				toAdd = append(toAdd, p0)
			}
		default:
			p0, err = p.newElementProperty(val)
			if err != nil {
				return nil, err
			}
			toAdd = append(toAdd, p0)
		}
	}
	if len(toAdd) == 0 {
		return nil, nil
	}

	// Add each candidate only if they do not match existing elements
	for _, eachToAdd := range toAdd {
		match := false
		for _, elem := range p.elements {
			if elem.Matches(eachToAdd) {
				match = true
				break
			}
		}
		if !match {
			p.elements = append(p.elements, eachToAdd)
			p.dirty = true
		}
	}

	return nil, nil
}

func (p *multiValuedProperty) Replace(value interface{}) (*Event, error) {
	if _, err := p.Delete(); err != nil {
		return nil, err
	}

	if _, err := p.Add(value); err != nil {
		return nil, err
	}

	return nil, nil
}

func (p *multiValuedProperty) Delete() (*Event, error) {
	if p.IsUnassigned() {
		return nil, nil
	}

	ev := Event{typ: EventUnassigned, source: p, pre: p.Raw()}
	p.dirty = true
	p.elements = make([]Property, 0)
	return &ev, nil
}

func (p *multiValuedProperty) Notify(events *Events) error {
	for _, sub := range p.subscribers {
		if err := sub.Notify(p, events); err != nil {
			return err
		}
	}
	return nil
}

func (p *multiValuedProperty) CountChildren() int {
	return len(p.elements)
}

func (p *multiValuedProperty) ForEachChild(callback func(index int, child Property) error) error {
	for i, elem := range p.elements {
		if err := callback(i, elem); err != nil {
			return err
		}
	}
	return nil
}

func (p *multiValuedProperty) FindChild(criteria func(child Property) bool) Property {
	for _, elem := range p.elements {
		if criteria(elem) {
			return elem
		}
	}
	return nil
}

func (p *multiValuedProperty) ChildAtIndex(index interface{}) (Property, error) {
	switch i := index.(type) {
	case int:
		if i < 0 || i >= len(p.elements) {
			return nil, fmt.Errorf("%w: no element at index '%d' of '%s'", spec.ErrNoTarget, i, p.attr.Path())
		}
		return p.elements[i], nil
	default:
		panic("invalid index type")
	}
}

func (p *multiValuedProperty) newElementProperty(singleValue interface{}) (prop Property, err error) {
	defer func() {
		if r := recover(); r != nil {
			prop = nil
			err = fmt.Errorf("%w: value incompatible with '%s' element", spec.ErrInvalidValue, p.attr.Path())
		}
	}()

	prop = NewProperty(p.attr.DeriveElementAttribute())
	if singleValue != nil {
		_, err = prop.Replace(singleValue)
	}

	return
}

func (p *multiValuedProperty) EqualsTo(value interface{}) bool {
	// This implementation is counter intuitive. It is implemented to allow for the
	// special scenario where SCIM uses 'eq' operator to match an element
	// within a multiValued property. Hence, consider this a special contains operation.
	if p.IsUnassigned() {
		return false
	}

	if _, ok := p.elements[0].(EqCapable); !ok {
		return false
	}

	for _, elem := range p.elements {
		if elem.(EqCapable).EqualsTo(value) {
			return true
		}
	}

	return false
}

func (p *multiValuedProperty) Present() bool {
	return !p.IsUnassigned()
}

var (
	_ PrCapable = (*complexProperty)(nil)
)

// NewChild is a hidden API to append a new prototype element in this multiValued property and return the index of
// the created property. Use property.(interface{ AppendElement() int }) to check for applicability.
func (p *multiValuedProperty) AppendElement() int {
	c, err := p.newElementProperty(nil)
	if err != nil {
		return -1
	}
	p.elements = append(p.elements, c)
	return len(p.elements) - 1
}

// Compact is a hidden API to remove unassigned elements from this multiValued property and effectively de-fragment
// the content of this property. Use property.(interface{ Compact() }) to check for applicability.
func (p *multiValuedProperty) Compact() {
	if len(p.elements) == 0 {
		return
	}

	var i int
	for i = len(p.elements) - 1; i >= 0; i-- {
		if p.elements[i].IsUnassigned() {
			if i == len(p.elements)-1 {
				p.elements = p.elements[:i]
			} else if i == 0 {
				p.elements = p.elements[i+1:]
			} else {
				p.elements = append(p.elements[:i], p.elements[i+1:]...)
			}
		}
	}
}

var (
	_ PrCapable = (*multiValuedProperty)(nil)
)
