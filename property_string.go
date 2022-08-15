package scim

import (
	"hash/fnv"
	"strings"
)

type stringProperty struct {
	attr *Attribute
	vs   *string
	hash uint64
}

func (p *stringProperty) Attr() *Attribute {
	return p.attr
}

func (p *stringProperty) Value() any {
	if p.Unassigned() {
		return nil
	}
	return *p.vs
}

func (p *stringProperty) Unassigned() bool {
	return p.vs == nil
}

func (p *stringProperty) Add(value any) error {
	return p.Set(value)
}

func (p *stringProperty) Set(value any) error {
	if value == nil {
		p.Delete()
		return nil
	}

	v, ok := value.(string)
	if !ok {
		return ErrInvalidValue
	}

	p.vs = &v
	if p.attr.caseExact {
		p.updateHash([]byte(*p.vs))
	} else {
		p.updateHash([]byte(strings.ToLower(*p.vs)))
	}

	return nil
}

func (p *stringProperty) Delete() {
	p.vs = nil
	p.hash = 0
}

func (p *stringProperty) Hash() uint64 {
	return p.hash
}

func (p *stringProperty) updateHash(bytes []byte) {
	h := fnv.New64a()
	_, _ = h.Write(bytes)
	p.hash = h.Sum64()
}

func (p *stringProperty) Len() int {
	return 0
}

func (p *stringProperty) ForEach(_ func(index int, child Property) error) error {
	return nil
}

func (p *stringProperty) Find(_ func(child Property) bool) Property {
	return nil
}

func (p *stringProperty) ByIndex(_ any) Property {
	return nil
}

func (p *stringProperty) equalsTo(value any) bool {
	if p.Unassigned() || value == nil {
		return false
	}

	s, ok := value.(string)
	if !ok {
		return false
	}

	return p.attr.compareString(*p.vs, s) == 0
}

func (p *stringProperty) greaterThan(value any) bool {
	if p.Unassigned() || value == nil {
		return false
	}

	s, ok := value.(string)
	if !ok {
		return false
	}

	return p.attr.compareString(*p.vs, s) > 0
}

func (p *stringProperty) lessThan(value any) bool {
	if p.Unassigned() || value == nil {
		return false
	}

	s, ok := value.(string)
	if !ok {
		return false
	}

	return p.attr.compareString(*p.vs, s) < 0
}

func (p *stringProperty) startsWith(value string) bool {
	return !p.Unassigned() && p.attr.hasPrefix(*p.vs, value)
}

func (p *stringProperty) endsWith(value string) bool {
	return !p.Unassigned() && p.attr.hasSuffix(*p.vs, value)
}

func (p *stringProperty) contains(value string) bool {
	return !p.Unassigned() && p.attr.containsString(*p.vs, value)
}

func (p *stringProperty) isPresent() bool {
	return !p.Unassigned() && len(*p.vs) > 0
}
