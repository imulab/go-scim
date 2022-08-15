package scim

import (
	"fmt"
	"hash/fnv"
	"strings"
)

type complexProperty struct {
	attr      *Attribute
	children  []Property     // modelled as slice for iteration stability
	nameIndex map[string]int // locates child property index by lower-cased name
}

func (p *complexProperty) Attr() *Attribute {
	return p.attr
}

func (p *complexProperty) Value() any {
	if p.Unassigned() {
		return nil
	}

	values := map[string]any{}
	for _, each := range p.children {
		if !each.Unassigned() {
			values[each.Attr().name] = each.Value()
		}
	}

	return values
}

func (p *complexProperty) Unassigned() bool {
	for _, each := range p.children {
		if !each.Unassigned() {
			return false
		}
	}
	return true
}

func (p *complexProperty) Clone() Property {
	p0 := p.attr.createProperty().(*complexProperty)
	for i := 0; i < len(p.children); i++ {
		if err := p0.children[i].Set(p.children[i].Value()); err != nil {
			panic(fmt.Errorf("failed to clone property: %s", err))
		}
	}
	return p0
}

func (p *complexProperty) Add(value any) error {
	if value == nil {
		return nil
	}

	values, ok := value.(map[string]any)
	if !ok {
		return ErrInvalidValue
	}

	for k, v := range values {
		if child := p.ByIndex(k); child != nil {
			if err := child.Set(v); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *complexProperty) Set(value any) error {
	p.Delete()
	return p.Add(value)
}

func (p *complexProperty) Delete() {
	for _, each := range p.children {
		each.Delete()
	}
}

func (p *complexProperty) Hash() uint64 {
	// only include idProps properties whose attribute is marked identity
	var idProps []Property
	_ = p.ForEach(func(_ int, child Property) error {
		if child.Attr().identity {
			idProps = append(idProps, child)
		}
		return nil
	})
	if len(idProps) == 0 {
		idProps = p.children
	}

	h := fnv.New64a()
	for _, each := range idProps {
		_, _ = h.Write([]byte(each.Attr().name))
		if !each.Unassigned() {
			_, _ = h.Write(uint64ToBytes(each.Hash()))
		}
	}

	return h.Sum64()
}

func (p *complexProperty) Len() int {
	return len(p.children)
}

func (p *complexProperty) ForEach(fn func(index int, child Property) error) error {
	for i, each := range p.children {
		if err := fn(i, each); err != nil {
			return err
		}
	}
	return nil
}

func (p *complexProperty) Find(criteria func(child Property) bool) Property {
	for _, each := range p.children {
		if criteria(each) {
			return each
		}
	}
	return nil
}

func (p *complexProperty) ByIndex(index any) Property {
	name, ok := index.(string)
	if !ok {
		return nil
	}

	i, ok := p.nameIndex[strings.ToLower(name)]
	if !ok {
		return nil
	}

	return p.children[i]
}

func (p *complexProperty) isPresent() bool {
	return !p.Unassigned()
}
