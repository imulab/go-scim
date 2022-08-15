package scim

import "hash/fnv"

type multiProperty struct {
	attr *Attribute
	elem []Property
}

func (p *multiProperty) Attr() *Attribute {
	return p.attr
}

func (p *multiProperty) Value() any {
	var values []any
	for _, each := range p.elem {
		// there is no need to check for Unassigned here: multiProperty elements are
		// automatically deduplicated and compacted.
		values = append(values, each.Value())
	}

	if values == nil {
		return []any{}
	}

	return values
}

func (p *multiProperty) Unassigned() bool {
	// elements will NOT contain unassigned elements as they are automatically compacted.
	// Hence, non-zero length elements means this property is assigned; otherwise, it is unassigned.
	return len(p.elem) == 0
}

func (p *multiProperty) Clone() Property {
	p0 := p.attr.createProperty().(*multiProperty)
	for _, each := range p.elem {
		p0.elem = append(p0.elem, cloneProperty(each))
	}
	return p0
}

func (p *multiProperty) Add(value any) error {
	if value == nil {
		return nil
	}

	switch newValue := value.(type) {
	case []any:
		return addSliceToMultiValueProperty(p, newValue)
	case []string:
		return addSliceToMultiValueProperty(p, newValue)
	case []int64:
		return addSliceToMultiValueProperty(p, newValue)
	case []float64:
		return addSliceToMultiValueProperty(p, newValue)
	case []bool:
		return addSliceToMultiValueProperty(p, newValue)
	}

	newElem := p.attr.toSingleValued().createProperty()
	if err := newElem.Set(value); err != nil {
		return err
	}

	primaryGuard := p.primarySwitchGuard()

	// Must be invoked later in the order of primaryGuard -> deduplicate -> compact, because:
	// primaryGuard may produce duplicate elements, which can be deduplicated; deduplication may
	// produce unassigned elements, which can be cleaned up by compaction.
	defer p.compact()
	defer p.deduplicate()
	defer primaryGuard()

	p.elem = append(p.elem, newElem)

	return nil
}

func (p *multiProperty) appendElement() (index int, post func()) {
	p.elem = append(p.elem, p.attr.toSingleValued().createProperty())
	index = len(p.elem) - 1
	primaryGuard := p.primarySwitchGuard()
	post = func() {
		primaryGuard()
		p.deduplicate()
		p.compact()
	}
	return
}

func (p *multiProperty) deduplicate() {
	if len(p.elem) < 2 {
		return
	}

	hashSet := map[uint64]struct{}{}

	for _, each := range p.elem {
		hash := each.Hash()
		if _, ok := hashSet[hash]; ok {
			// just delete it, unassigned elements can later be removed by compact.
			each.Delete()
		} else {
			hashSet[hash] = struct{}{}
		}
	}
}

func (p *multiProperty) compact() {
	for i := p.Len() - 1; i >= 0; i-- {
		if p.elem[i].Unassigned() {
			p.elem = append(p.elem[:i], p.elem[i+1:]...)
		}
	}
}

func (p *multiProperty) primarySwitchGuard() (post func()) {
	if p.attr.typ != TypeComplex || p.Len() < 2 {
		post = func() {}
		return
	}

	before := p.Find(func(child Property) bool {
		return child.Attr().primary && child.Value() == true
	})
	if before == nil {
		post = func() {}
		return
	}

	post = func() {
		if before.Value() != true {
			return
		}

		_ = p.ForEach(func(_ int, child Property) error {
			if child == before {
				return nil
			}

			if !child.Attr().primary || child.Value() != true {
				return nil
			}

			_ = before.Set(false)

			return nil
		})
	}

	return
}

func (p *multiProperty) Set(value any) error {
	if value == nil {
		p.Delete()
		return nil
	}

	switch value.(type) {
	case []any, []string, []int64, []float64, []bool:
		p.Delete()
		return p.Add(value)
	default:
		return ErrInvalidValue
	}
}

func (p *multiProperty) Delete() {
	p.elem = []Property{}
}

func (p *multiProperty) Hash() uint64 {
	if len(p.elem) == 0 {
		return 0
	}

	var hashes []uint64
	_ = p.ForEach(func(_ int, child Property) error {
		if child.Unassigned() {
			return nil
		}

		hashes = append(hashes, child.Hash())

		// SCIM array does not have orders. We keep the hash array
		// sorted so that different multiValue properties containing
		// the same elements in different orders can be recognized as
		// the same, as they compute the same hash. We use insertion
		// sort here as we don't expect a large number of elements.
		for i := len(hashes) - 1; i > 0; i-- {
			if hashes[i-1] > hashes[i] {
				hashes[i-1], hashes[i] = hashes[i], hashes[i-1]
			}
		}

		return nil
	})

	h := fnv.New64a()
	for _, hash := range hashes {
		_, _ = h.Write(uint64ToBytes(hash))
	}

	return h.Sum64()
}

func (p *multiProperty) Len() int {
	return len(p.elem)
}

func (p *multiProperty) ForEach(fn func(index int, child Property) error) error {
	for i, each := range p.elem {
		if err := fn(i, each); err != nil {
			return err
		}
	}
	return nil
}

func (p *multiProperty) Find(criteria func(child Property) bool) Property {
	for _, each := range p.elem {
		if criteria(each) {
			return each
		}
	}
	return nil
}

func (p *multiProperty) ByIndex(index any) Property {
	i, ok := index.(int)
	if !ok {
		return nil
	}

	if i < 0 || i >= len(p.elem) {
		return nil
	}

	return p.elem[i]
}

func (p *multiProperty) equalsTo(value any) bool {
	// This implementation is counter-intuitive. It is implemented to allow for the
	// special scenario where SCIM uses 'eq' operator to match an element
	// within a multiValued property. Hence, consider this a special contains operation.
	if p.Unassigned() {
		return false
	}

	return p.Find(func(child Property) bool {
		if eqChild, ok := child.(eqTrait); !ok {
			return false
		} else {
			return eqChild.equalsTo(value)
		}
	}) != nil
}

func (p *multiProperty) isPresent() bool {
	return !p.Unassigned()
}

// addSliceToMultiValueProperty is here because methods cannot have generic type parameters as of Go1.18.1.
func addSliceToMultiValueProperty[E interface {
	any | string | int64 | float64 | bool
}](p *multiProperty, values []E) error {
	for _, each := range values {
		if err := p.Add(each); err != nil {
			return err
		}
	}
	return nil
}
