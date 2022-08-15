package scim

import (
	"encoding/binary"
	"fmt"
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

	// Add modifies the value of the property by adding a new value to it. For singular non-complex properties, the
	// Add operation is identical to the Set operation. If the given value is incompatible with the attribute,
	// implementations should return ErrInvalidValue.
	Add(value any) error

	// Set modifies the value of the property by completely replace its value with the new value. Calling Set wil nil
	// is semantically identical to calling Delete. Otherwise, if the given value is incompatible with the attribute,
	// implementations should return ErrInvalidValue.
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

type appendElement interface {
	appendElement() (index int, post func())
}

// trait interfaces for property implementations so they can participate in expression evaluations.
type (
	eqTrait interface{ equalsTo(value any) bool }
	swTrait interface{ startsWith(value string) bool }
	ewTrait interface{ endsWith(value string) bool }
	coTrait interface{ contains(value string) bool }
	gtTrait interface{ greaterThan(value any) bool }
	ltTrait interface{ lessThan(value any) bool }
	prTrait interface{ isPresent() bool }
)

func cloneProperty(p Property) Property {
	p0 := p.Attr().createProperty()

	if p.Unassigned() {
		return p0
	}

	if err := p0.Set(p.Value()); err != nil {
		panic(fmt.Errorf("failed to clone property: %s", err))
	}

	return p0
}

func uint64ToBytes(u uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, u)
	return b
}
