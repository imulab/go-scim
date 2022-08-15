package scim

import "hash/fnv"

type decimalProperty struct {
	attr *Attribute
	vd   *float64
	hash uint64
}

func (p *decimalProperty) Attr() *Attribute {
	return p.attr
}

func (p *decimalProperty) Value() any {
	if p.Unassigned() {
		return nil
	}
	return *p.vd
}

func (p *decimalProperty) Unassigned() bool {
	return p.vd == nil
}

func (p *decimalProperty) Add(value any) error {
	return p.Set(value)
}

func (p *decimalProperty) Set(value any) error {
	if value == nil {
		p.Delete()
		return nil
	}

	f, ok := value.(float64)
	if !ok {
		return ErrInvalidValue
	}

	p.vd = &f
	p.updateHash(uint64ToBytes(uint64(f)))

	return nil
}

func (p *decimalProperty) Delete() {
	p.vd = nil
	p.hash = 0
}

func (p *decimalProperty) Hash() uint64 {
	return p.hash
}

func (p *decimalProperty) updateHash(bytes []byte) {
	h := fnv.New64a()
	_, _ = h.Write(bytes)
	p.hash = h.Sum64()
}

func (p *decimalProperty) Len() int {
	return 0
}

func (p *decimalProperty) ForEach(_ func(index int, child Property) error) error {
	return nil
}

func (p *decimalProperty) Find(_ func(child Property) bool) Property {
	return nil
}

func (p *decimalProperty) ByIndex(_ any) Property {
	return nil
}

func (p *decimalProperty) equalsTo(value any) bool {
	r, ok := p.compare(value)
	return ok && r == 0
}

func (p *decimalProperty) greaterThan(value any) bool {
	r, ok := p.compare(value)
	return ok && r > 0
}

func (p *decimalProperty) lessThan(value any) bool {
	r, ok := p.compare(value)
	return ok && r < 0
}

func (p *decimalProperty) isPresent() bool {
	return !p.Unassigned()
}

func (p *decimalProperty) compare(value any) (int, bool) {
	if p.Unassigned() || value == nil {
		return 0, false
	}

	f, ok := value.(float64)
	if !ok {
		return 0, false
	}

	switch {
	case *p.vd > f:
		return 1, true
	case *p.vd < f:
		return -1, true
	default:
		return 0, true
	}
}
