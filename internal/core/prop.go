package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
)

var (
	ErrValue = errors.New("value is incompatible with attribute")
)

// Property abstracts a single node on the resource value tree. It may hold values, or other properties, as described
// by its Attribute.
type Property interface {
	// Attr returns the Attribute of this Property. The returned value should always be non-nil.
	Attr() *Attribute

	// Value returns the held value in Go's native type. For unassigned properties, the value returned should always be
	// nil. For assigned multivalued properties, the values returned should be of type []any. For other assigned properties,
	// the following relations between attributes and values apply:
	//
	//	string, dateTime, reference, binary -> string
	//	integer -> int64
	//	decimal -> float64
	//	boolean -> bool
	//	complex -> map[string]any
	//
	// Failure of adhering to the above rules may cause the library to panic.
	Value() any

	// Unassigned returns true if the property is unassigned. It is recommended to check the unassigned-ness of the
	// property before continuing to other operations, as implementations may choose to maintain inner structure
	// despite the lack of data for efficiency purposes.
	Unassigned() bool

	// Clone returns a deep copy of this property. The returned property can share the same instance of Attribute with
	// the original property (because attributes are considered readonly after setup), but must have distinct instances
	// of values.
	Clone() Property

	// Add modifies the value of the property by adding a new value to it. For singular non-complex properties, the
	// Add operation is identical to the Set operation. If the given value is incompatible with the attribute, implementations
	// should return ErrValue.
	Add(value any) error

	// Set modifies the value of the property by completely replace its value with the new value. Calling Set wil nil
	// is semantically identical to calling Delete. Otherwise, if the given value is incompatible with the attribute,
	// implementations should return ErrValue.
	Set(value any) error

	// Delete modifies the value of the property by removing it. After the operation, Unassigned should return true.
	Delete()

	// Hash returns the hash identity of this property. Properties with the same Attribute and hash value are considered
	// to be identical.
	Hash() uint64

	// Len returns the number of sub-properties contained by this property. For singular non-complex properties, Len
	// always returns 0.
	Len() int

	// ForEach iterates through the sub-properties of this property. The visitor function fn is invoked on each visited
	// sub-property. Any error returned by the visitor function terminates the traversal process. Implementations are
	// responsible to ensure stability of the traversal.
	ForEach(fn func(index int, child Property) error) error

	// Find returns the first sub-property that matches the given criteria. If no such property is found, nil is returned.
	Find(criteria func(child Property) bool) Property

	// ByIndex returns the sub property of this Property by an index. For complex properties, this index is expected to
	// be a string, containing name of the sub property; for multiValued properties, this index is expected to be a
	// number, containing array index of the target sub property. For singular non-complex properties, this method is
	// expected to always return nil.
	ByIndex(index any) Property
}

type simpleProperty struct {
	attr *Attribute
	vs   *string
	vi   *int64
	vf   *float64
	vb   *bool
	hash uint64
}

func (p *simpleProperty) Attr() *Attribute {
	return p.attr
}

func (p *simpleProperty) Value() any {
	if p.Unassigned() {
		return nil
	}

	switch p.attr.Type {
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
	switch p.attr.Type {
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
	p0 := p.attr.NewProperty()

	if p.Unassigned() {
		return p0
	}

	if err := p0.Set(p.Value()); err != nil {
		panic(fmt.Errorf("failed to clone property: %s", err))
	}

	return p0
}

func (p *simpleProperty) Add(value any) error {
	return p.Set(value)
}

func (p *simpleProperty) Set(value any) error {
	if value == nil {
		p.Delete()
		return nil
	}

	switch p.attr.Type {
	case TypeString, TypeBinary, TypeDateTime, TypeReference:
		if v, ok := value.(string); ok {
			p.vs = &v
			if p.attr.CaseExact {
				p.updateHash([]byte(*p.vs))
			} else {
				p.updateHash([]byte(strings.ToLower(*p.vs)))
			}
			return nil
		}

	case TypeInteger:
		if v, ok := value.(int64); ok {
			p.vi = &v
			p.updateHash(uint64ToBytes(uint64(v)))
			return nil
		}

	case TypeDecimal:
		if v, ok := value.(float64); ok {
			p.vf = &v
			p.updateHash(uint64ToBytes(uint64(v)))
			return nil
		}

	case TypeBoolean:
		if v, ok := value.(bool); ok {
			p.vb = &v
			if v {
				p.updateHash(uint64ToBytes(1))
			} else {
				p.updateHash(uint64ToBytes(0))
			}
			return nil
		}
	}

	return ErrValue
}

func (p *simpleProperty) Delete() {
	p.vs = nil
	p.vi = nil
	p.vf = nil
	p.vb = nil
	p.hash = 0
}

func (p *simpleProperty) Hash() uint64 {
	return p.hash
}

func (p *simpleProperty) updateHash(bytes []byte) {
	h := fnv.New64a()
	_, _ = h.Write(bytes)
	p.hash = h.Sum64()
}

func (p *simpleProperty) Len() int {
	return 0
}

func (p *simpleProperty) ForEach(_ func(index int, child Property) error) error {
	return nil
}

func (p *simpleProperty) Find(_ func(child Property) bool) Property {
	return nil
}

func (p *simpleProperty) ByIndex(_ any) Property {
	return nil
}

type complexProperty struct {
	attr *Attribute
	// children modelled as slice for iteration stability
	children []Property
	// nameIndex locates child property index by lower-cased name
	nameIndex map[string]int
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
			values[each.Attr().Name] = each.Value()
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
	p0 := p.attr.NewProperty().(*complexProperty)
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
		return ErrValue
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
		if child.Attr().Identity {
			idProps = append(idProps, child)
		}
		return nil
	})
	if len(idProps) == 0 {
		idProps = p.children
	}

	h := fnv.New64a()
	for _, each := range idProps {
		_, _ = h.Write([]byte(each.Attr().Name))
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
	p0 := p.attr.NewProperty().(*multiProperty)
	for _, each := range p.elem {
		p0.elem = append(p0.elem, each.Clone())
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

	newElem := p.attr.asSingleValued().NewProperty()
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
	if p.attr.Type != TypeComplex || p.Len() < 2 {
		post = func() {}
		return
	}

	before := p.Find(func(child Property) bool {
		return child.Attr().Primary && child.Value() == true
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

			if !child.Attr().Primary || child.Value() != true {
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
		return ErrValue
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

func uint64ToBytes(u uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, u)
	return b
}
