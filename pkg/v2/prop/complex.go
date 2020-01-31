package prop

import (
	"encoding/binary"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/annotation"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"hash/fnv"
	"strings"
)

// NewComplex creates a new complex property associated with attribute. All sub attributes are created.
func NewComplex(attr *spec.Attribute) Property {
	ensureSingularComplexType(attr)
	p := complexProperty{
		attr:        attr,
		subProps:    []Property{},
		nameIndex:   map[string]int{},
		subscribers: []Subscriber{},
	}
	attr.ForEachAnnotation(func(annotation string, params map[string]interface{}) {
		if subscriber, ok := SubscriberFactory().Create(annotation, &p, params); ok {
			p.subscribers = append(p.subscribers, subscriber)
		}
	})
	_ = attr.ForEachSubAttribute(func(subAttribute *spec.Attribute) error {
		p.subProps = append(p.subProps, NewProperty(subAttribute))
		p.nameIndex[strings.ToLower(subAttribute.Name())] = len(p.subProps) - 1
		return nil
	})
	return &p
}

// NewComplexOf creates a new complex property of given value associated with attribute.
func NewComplexOf(attr *spec.Attribute, value map[string]interface{}) Property {
	p := NewComplex(attr)
	_, err := p.Replace(value)
	if err != nil {
		panic(err)
	}
	return p
}

func ensureSingularComplexType(attr *spec.Attribute) {
	if attr.MultiValued() || attr.Type() != spec.TypeComplex {
		panic("invalid attribute for complex property")
	}
}

type complexProperty struct {
	attr        *spec.Attribute
	subProps    []Property     // array of sub properties to maintain determinate iteration order
	nameIndex   map[string]int // attribute's name (to lower case) to index in subProps to allow fast access
	subscribers []Subscriber
}

func (p *complexProperty) Attribute() *spec.Attribute {
	return p.attr
}

// Caution: slow operation
func (p *complexProperty) Raw() interface{} {
	values := map[string]interface{}{}
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

func (p *complexProperty) Dirty() bool {
	for _, subProp := range p.subProps {
		if subProp.Dirty() {
			return true
		}
	}
	return false
}

func (p *complexProperty) identitySubAttributes() map[*spec.Attribute]struct{} {
	idSubAttr := map[*spec.Attribute]struct{}{}
	_ = p.attr.ForEachSubAttribute(func(subAttribute *spec.Attribute) error {
		if _, ok := subAttribute.Annotation(annotation.Identity); ok {
			idSubAttr[subAttribute] = struct{}{}
		}
		return nil
	})
	return idSubAttr
}

// Caution: expensive
func (p *complexProperty) Hash() uint64 {
	if p.IsUnassigned() {
		return 0
	}

	var (
		h         = fnv.New64a()
		idSubAttr = p.identitySubAttributes()
	)
	if err := p.ForEachChild(func(_ int, child Property) error {
		if _, ok := idSubAttr[child.Attribute()]; !ok && len(idSubAttr) > 0 {
			return nil // do not include in computation if complex has identity attributes but this is not one of them.
		}

		if _, err := h.Write([]byte(child.Attribute().Name())); err != nil {
			return err
		}

		if child.IsUnassigned() {
			return nil // Skip the value hash if it is unassigned
		}

		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, child.Hash())
		if _, err := h.Write(b); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic("error computing hash")
	}

	return h.Sum64()
}

func (p *complexProperty) Matches(another Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}
	if p.CountChildren() != another.CountChildren() {
		return false // Usually this won't happen, but still check it to be sure.
	}
	return p.Hash() == another.Hash()
}

func (p *complexProperty) Clone() Property {
	c := complexProperty{
		attr:        p.attr,
		subProps:    make([]Property, 0, len(p.subProps)),
		nameIndex:   make(map[string]int),
		subscribers: p.subscribers,
	}
	for i, sp := range p.subProps {
		c.subProps = append(c.subProps, sp.Clone())
		c.nameIndex[strings.ToLower(sp.Attribute().Name())] = i
	}
	return &c
}

func (p *complexProperty) Add(value interface{}) (*Event, error) {
	if value == nil {
		return nil, nil
	}

	m, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: value is incompatible with '%s'", spec.ErrInvalidValue, p.attr.Path())
	}

	wasUnassigned := p.IsUnassigned()

	for k, v := range m {
		i, ok := p.nameIndex[strings.ToLower(k)]
		if !ok {
			continue
		}
		if _, err := p.subProps[i].Add(v); err != nil {
			return nil, err
		}
	}

	if wasUnassigned && !p.IsUnassigned() {
		return EventAssigned.NewFrom(p, nil), nil
	} else if !wasUnassigned && p.IsUnassigned() {
		return EventUnassigned.NewFrom(p, nil), nil
	}

	return nil, nil
}

func (p *complexProperty) Replace(value interface{}) (*Event, error) {
	if value == nil {
		return nil, nil
	}

	wasUnassigned := p.IsUnassigned()

	if _, err := p.Delete(); err != nil {
		return nil, err
	}

	if _, err := p.Add(value); err != nil {
		return nil, err
	}

	if wasUnassigned && !p.IsUnassigned() {
		return EventAssigned.NewFrom(p, nil), nil
	} else if !wasUnassigned && p.IsUnassigned() {
		return EventUnassigned.NewFrom(p, nil), nil
	}

	return nil, nil
}

func (p *complexProperty) Delete() (*Event, error) {
	if p.IsUnassigned() {
		return nil, nil
	}

	for _, sp := range p.subProps {
		if _, err := sp.Delete(); err != nil {
			return nil, err
		}
	}

	return EventUnassigned.NewFrom(p, nil), nil
}

func (p *complexProperty) Notify(events *Events) error {
	for _, sub := range p.subscribers {
		if err := sub.Notify(p, events); err != nil {
			return err
		}
	}
	return nil
}

func (p *complexProperty) CountChildren() int {
	return len(p.subProps)
}

func (p *complexProperty) ForEachChild(callback func(index int, child Property) error) error {
	for i, sp := range p.subProps {
		if err := callback(i, sp); err != nil {
			return err
		}
	}
	return nil
}

func (p *complexProperty) FindChild(criteria func(child Property) bool) Property {
	for _, sp := range p.subProps {
		if criteria(sp) {
			return sp
		}
	}
	return nil
}

func (p *complexProperty) ChildAtIndex(index interface{}) (Property, error) {
	switch i := index.(type) {
	case string:
		ni, ok := p.nameIndex[strings.ToLower(i)]
		if !ok {
			return nil, fmt.Errorf("%w: '%s' does not have child '%s'", spec.ErrInvalidPath, p.attr.Path(), i)
		}
		return p.subProps[ni], nil
	default:
		panic("invalid index type")
	}
}

func (p *complexProperty) Present() bool {
	// complex property is present iff it was modified and is currently not unassigned.
	// non-dirty unassigned complex property is an internal representation, it does not
	// indicate presence of user data or operations.
	return p.Dirty() && !p.IsUnassigned()
}

var (
	_ PrCapable = (*complexProperty)(nil)
)
