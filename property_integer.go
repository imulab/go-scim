package scim

import "hash/fnv"

// TODO complete the rest of properties
// TODO remove simpleProperty
// TODO evaluator
// TODO traverseQualifiedElements
// TODO crud methods on resources

type integerProperty struct {
	attr *Attribute
	vi   *int64
	hash uint64
}

func (p *integerProperty) Attr() *Attribute {
	return p.attr
}

func (p *integerProperty) Value() any {
	if p.Unassigned() {
		return nil
	}
	return *p.vi
}

func (p *integerProperty) Unassigned() bool {
	return p.vi == nil
}

func (p *integerProperty) Add(value any) error {
	return p.Add(value)
}

func (p *integerProperty) Set(value any) error {
	if value == nil {
		p.Delete()
		return nil
	}

	i, ok := value.(int64)
	if !ok {
		return ErrInvalidValue
	}

	p.vi = &i
	p.updateHash(uint64ToBytes(uint64(i)))

	return nil
}

func (p *integerProperty) Delete() {
	p.vi = nil
	p.hash = 0
}

func (p *integerProperty) Hash() uint64 {
	return p.hash
}

func (p *integerProperty) updateHash(bytes []byte) {
	h := fnv.New64a()
	_, _ = h.Write(bytes)
	p.hash = h.Sum64()
}

func (p *integerProperty) Len() int {
	return 0
}

func (p *integerProperty) ForEach(_ func(index int, child Property) error) error {
	return nil
}

func (p *integerProperty) Find(_ func(child Property) bool) Property {
	return nil
}

func (p *integerProperty) ByIndex(_ any) Property {
	return nil
}

func (p *integerProperty) equalsTo(value any) bool {
	r, ok := p.compare(value)
	return ok && r == 0
}

func (p *integerProperty) greaterThan(value any) bool {
	r, ok := p.compare(value)
	return ok && r > 0
}

func (p *integerProperty) lessThan(value any) bool {
	r, ok := p.compare(value)
	return ok && r < 0
}

func (p *integerProperty) isPresent() bool {
	return !p.Unassigned()
}

func (p *integerProperty) compare(value any) (int, bool) {
	if p.Unassigned() || value == nil {
		return 0, false
	}

	i, ok := value.(int64)
	if !ok {
		return 0, false
	}

	switch {
	case *p.vi > i:
		return 1, true
	case *p.vi < i:
		return -1, true
	default:
		return 0, true
	}
}
