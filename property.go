package scim

import (
	"encoding/binary"
	"hash/fnv"
	"strings"
)

// Property describes a property of SCIM Resource.
type Property interface {
	// Attribute returns a non-nil object to describe the data of this property.
	Attribute() *Attribute

	// Value returns the property's native Go value.
	//
	// The value returned are constrained by a few rules. First, unassigned property should return nil as value.
	// Second, multiValued property should return []any type. Third, properties with other attribute types must
	// return a Go type that satisfies their corresponding SCIM type:
	//
	//	string, dateTime, reference, binary -> string
	//	integer -> int64
	//	decimal -> float64
	//	boolean -> bool
	//	complex -> map[string]any
	Value() any

	// Unassigned returns true when this property is unassigned.
	//
	// For singular attributes, the property is unassigned when Value returns nil; for multiValued attributes, the
	// property is unassigned when it has no elements; for complex attributes, the property is unassigned if and only if
	// all its sub properties are unassigned.
	Unassigned() bool

	// Clone returns a deep copy of this Property. The returned property may share the same instance of Root and
	// Attribute, but must return distinct instance of Value.
	Clone() Property

	// Set replaces the given value for this Property. The type of the value varies according to attribute type and
	// implementations. Setting a Property with nil will make this property Unassigned.
	Set(value any) error

	// Add adds the given value to this Property. For singular non-complex types, it is semantically identical to Set.
	// For singular complex types, it sets values on a subset of all sub properties. For multiValued types, it simply
	// appends a new element property.
	Add(value any) error

	// Delete removes the value of this Property. A deleted Property is always Unassigned.
	Delete()

	// Hash returns the hash identity of this Property. Unassigned properties should always return zero.
	Hash() uint64

	// Len counts the number of sub properties hosted under this Property. This method is meaningful for properties with
	// complex and/or multiValued Attribute. For singular non-complex attributes, this method should always return zero.
	Len() int

	// Iterate iterates the sub properties of this Property and invokes the callback function. Any error returned by the
	// callback function will be immediately returned. For complex properties, the iteration index is always zero.
	Iterate(fn func(index int, child Property) error) error

	// Find searches for the first sub property of this Property that meets the criteria. If no sub property matches
	// the criteria. Find should return nil.
	Find(criteria func(child Property) bool) Property

	// ByIndex returns the sub property of this Property by an index. For complex properties, this index is expected to
	// be a string, containing name of the sub property; for multiValued properties, this index is expected to be a
	// number, containing array index of the target sub property. For singular non-complex properties, this method is
	// expected to always return nil.
	ByIndex(index any) Property
}

// simpleProperty implements Property for all properties singular non-complex attribute.
type simpleProperty struct {
	attr *Attribute
	vs   *string
	vi   *int64
	vf   *float64
	vb   *bool
	hash uint64
}

func (p *simpleProperty) Attribute() *Attribute {
	return p.attr
}

func (p *simpleProperty) Value() any {
	if p.Unassigned() {
		return nil
	}

	switch p.attr.typ {
	case TypeString, TypeBinary, TypeDateTime, TypeReference:
		return *p.vs
	case TypeInteger:
		return *p.vi
	case TypeDecimal:
		return *p.vf
	case TypeBoolean:
		return *p.vb
	default:
		panic("unexpected simple property type")
	}
}

func (p *simpleProperty) Unassigned() bool {
	switch p.attr.typ {
	case TypeString, TypeBinary, TypeDateTime, TypeReference:
		return p.vs == nil
	case TypeInteger:
		return p.vi == nil
	case TypeDecimal:
		return p.vf == nil
	case TypeBoolean:
		return p.vb == nil
	default:
		panic("unexpected simple property type")
	}
}

func (p *simpleProperty) Clone() Property {
	p0 := createProperty(p.attr).(*simpleProperty)
	_ = p0.Set(p.Value())
	return p0
}

func (p *simpleProperty) Set(value any) error {
	if value == nil {
		p.Delete()
		return nil
	}

	switch p.attr.typ {
	case TypeString, TypeBinary, TypeDateTime, TypeReference:
		if v, ok := value.(string); ok {
			p.vs = &v

			if p.attr.caseExact {
				p.setHash([]byte(*p.vs))
			} else {
				p.setHash([]byte(strings.ToLower(*p.vs)))
			}

			return nil
		}

	case TypeInteger:
		if v, ok := value.(int64); ok {
			p.vi = &v

			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(v))
			p.setHash(b)

			return nil
		}

	case TypeDecimal:
		if v, ok := value.(float64); ok {
			p.vf = &v

			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(v))
			p.setHash(b)

			return nil
		}

	case TypeBoolean:
		if v, ok := value.(bool); ok {
			p.vb = &v

			b := make([]byte, 8)
			if v {
				binary.LittleEndian.PutUint64(b, uint64(1))
			} else {
				binary.LittleEndian.PutUint64(b, uint64(0))
			}
			p.setHash(b)

			return nil
		}
	}

	return ErrInvalidValue
}

func (p *simpleProperty) Add(value any) error {
	return p.Set(value)
}

func (p *simpleProperty) Delete() {
	p.vs = nil
	p.vi = nil
	p.vf = nil
	p.vb = nil
	p.hash = 0
}

func (p *simpleProperty) setHash(bytes []byte) {
	h := fnv.New64a()
	_, _ = h.Write(bytes)
	p.hash = h.Sum64()
}

func (p *simpleProperty) Hash() uint64 {
	return p.hash
}

func (p *simpleProperty) Len() int {
	return 0
}

func (p *simpleProperty) Iterate(_ func(index int, child Property) error) error {
	return nil
}

func (p *simpleProperty) Find(_ func(child Property) bool) Property {
	return nil
}

func (p *simpleProperty) ByIndex(_ any) Property {
	return nil
}

// complexProperty implements Property for properties with singular complex attribute.
type complexProperty struct {
	attr      *Attribute
	sub       []Property
	nameIndex map[string]int // lower-cased sub attribute name to sub index for fast lookup
}

func (p *complexProperty) Attribute() *Attribute {
	return p.attr
}

func (p *complexProperty) Value() any {
	if p.Unassigned() {
		return nil
	}

	values := map[string]any{}

	for _, each := range p.sub {
		if !each.Unassigned() {
			values[each.Attribute().name] = each.Value()
		}
	}

	return values
}

func (p *complexProperty) Unassigned() bool {
	for _, each := range p.sub {
		if !each.Unassigned() {
			return false
		}
	}
	return true
}

func (p *complexProperty) Clone() Property {
	p0 := createProperty(p.attr).(*complexProperty)
	for i := 0; i < len(p.sub); i++ {
		_ = p0.sub[i].Set(p.sub[i].Value())
	}
	return p0
}

func (p *complexProperty) Set(value any) error {
	p.Delete()
	return p.Add(value)
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
		if subProp := p.dot(k); subProp != nil {
			if err := subProp.Set(v); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *complexProperty) Delete() {
	p.resetChildren()
}

func (p *complexProperty) Hash() uint64 {
	// only include sub properties whose attribute is marked identity
	var sub []Property
	_ = p.Iterate(func(_ int, child Property) error {
		if child.Attribute().identity {
			sub = append(sub, child)
		}
		return nil
	})
	if len(sub) == 0 {
		sub = p.sub
	}

	h := fnv.New64a()
	for _, each := range sub {
		_, _ = h.Write([]byte(each.Attribute().name))
		if !each.Unassigned() {
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, each.Hash())
			_, _ = h.Write(b)
		}
	}

	return h.Sum64()
}

func (p *complexProperty) Len() int {
	return len(p.sub)
}

func (p *complexProperty) Iterate(fn func(index int, child Property) error) error {
	for i, each := range p.sub {
		if err := fn(i, each); err != nil {
			return err
		}
	}

	return nil
}

func (p *complexProperty) Find(criteria func(child Property) bool) Property {
	for _, each := range p.sub {
		if ok := criteria(each); ok {
			return each
		}
	}

	return nil
}

func (p *complexProperty) ByIndex(index any) Property {
	switch name := index.(type) {
	case string:
		return p.dot(name)
	default:
		panic("unexpected index type for complex property")
	}
}

func (p *complexProperty) dot(name string) Property {
	i, ok := p.nameIndex[strings.ToLower(name)]
	if !ok {
		return nil
	}
	return p.sub[i]
}

func (p *complexProperty) resetChildren() {
	for _, each := range p.sub {
		each.Delete()
	}
}

// multiValuedProperty implements Property for properties with multiValued attribute.
type multiValuedProperty struct {
	attr *Attribute
	elem []Property
}

func (p *multiValuedProperty) Attribute() *Attribute {
	return p.attr
}

func (p *multiValuedProperty) Value() any {
	var values []any
	for _, each := range p.elem {
		if v := each.Value(); v != nil {
			values = append(values, v)
		}
	}

	if values == nil {
		return []any{}
	}

	return values
}

func (p *multiValuedProperty) Unassigned() bool {
	if len(p.elem) == 0 {
		return true
	}

	// TODO(q)
	// since all additions are performed through Add (Set is delegated to Add) where compaction is performed, it should
	// be asserted that non-empty elements are always assigned, essentially reduce this method to len(p.elem) == 0 and
	// boost its performance.
	for _, each := range p.elem {
		if !each.Unassigned() {
			return false
		}
	}

	return true
}

func (p *multiValuedProperty) Clone() Property {
	p0 := createProperty(p.attr).(*multiValuedProperty)
	for _, each := range p.elem {
		p0.elem = append(p0.elem, each.Clone())
	}
	return p0
}

func (p *multiValuedProperty) Set(value any) error {
	p.Delete()

	if value == nil {
		return nil
	}

	switch value.(type) {
	case []any, []string, []int64, []float64, []bool:
		return p.Add(value)
	default:
		return ErrInvalidValue
	}
}

func (p *multiValuedProperty) Add(value any) error {
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

	newElem := createProperty(p.attr.toSingleValued())
	if err := newElem.Set(value); err != nil {
		return err
	}

	primaryGuard := p.primarySwitch()

	// Must be invoked later in the order of primaryGuard -> deduplicate -> compact, because:
	// primaryGuard may produce duplicate elements, which can be deduplicated; deduplication may
	// produce unassigned elements, which can be cleaned up by compaction.
	defer p.compact()
	defer p.deduplicate()
	defer primaryGuard()

	p.elem = append(p.elem, newElem)

	return nil
}

func addSliceToMultiValueProperty[E interface {
	any | string | int64 | float64 | bool
}](p *multiValuedProperty, values []E) error {
	for _, each := range values {
		if err := p.Add(each); err != nil {
			return err
		}
	}
	return nil
}

func (p *multiValuedProperty) Delete() {
	p.elem = []Property{}
}

func (p *multiValuedProperty) Hash() uint64 {
	if len(p.elem) == 0 {
		return 0
	}

	var hashes []uint64
	_ = p.Iterate(func(_ int, child Property) error {
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
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, hash)
		_, _ = h.Write(b)
	}

	return h.Sum64()
}

func (p *multiValuedProperty) Len() int {
	return len(p.elem)
}

func (p *multiValuedProperty) Iterate(fn func(index int, child Property) error) error {
	for i, each := range p.elem {
		if err := fn(i, each); err != nil {
			return err
		}
	}
	return nil
}

func (p *multiValuedProperty) Find(criteria func(child Property) bool) Property {
	for _, each := range p.elem {
		if criteria(each) {
			return each
		}
	}
	return nil
}

func (p *multiValuedProperty) ByIndex(index any) Property {
	switch i := index.(type) {
	case int:
		if i < 0 || i >= len(p.elem) {
			panic("index of out range")
		} else {
			return p.elem[i]
		}
	default:
		panic("unexpected index type for multiValued property")
	}
}

func (p *multiValuedProperty) deduplicate() {
	if len(p.elem) < 2 {
		return
	}

	hashSet := map[uint64]struct{}{}

	for _, each := range p.elem {
		hash := each.Hash()
		if _, ok := hashSet[hash]; ok {
			each.Delete()
		} else {
			hashSet[hash] = struct{}{}
		}
	}
}

func (p *multiValuedProperty) compact() {
	for i := p.Len() - 1; i >= 0; i-- {
		if p.elem[i].Unassigned() {
			p.elem = append(p.elem[:i], p.elem[i+1:]...)
		}
	}
}

func (p *multiValuedProperty) primarySwitch() func() {
	if p.attr.typ != TypeComplex {
		return func() {}
	}

	if len(p.elem) < 2 {
		return func() {} // no need to guard the primary property if there's less than two.
	}

	pre := p.Find(func(child Property) bool {
		return child.Attribute().primary && child.Value() == true
	})

	return func() {
		if pre.Value() != true {
			return
		}

		_ = p.Iterate(func(_ int, child Property) error {
			if child == pre {
				return nil
			}

			if !child.Attribute().primary || child.Value() != true {
				return nil
			}

			_ = pre.Set(false)

			return nil
		})
	}
}

func createProperty(attr *Attribute) Property {
	if attr.multiValued {
		return &multiValuedProperty{attr: attr, elem: []Property{}}
	}

	if attr.typ == TypeComplex {
		p := &complexProperty{attr: attr, nameIndex: map[string]int{}}
		for i, subAttr := range p.attr.subAttrs {
			p.sub = append(p.sub, createProperty(subAttr))
			p.nameIndex[strings.ToLower(subAttr.name)] = i
		}
		return p
	}

	return &simpleProperty{attr: attr}
}
