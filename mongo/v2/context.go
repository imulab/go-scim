package v2

// Frame preserves vital information within the traversal stack frame during BSON serialization
type frame struct {
	// mode (or context) of the current context
	mode mode
	// index keeps track of the number of element visited in
	// a container (i.e. complex, multiValued). It could be used
	// as a key for array elements.
	index int
	// keeps track of the byte index where the current container started
	start int
}

type mode int

const (
	_ mode = iota
	mObject
	mArray
	// mTop denotes the top level object. It is distinguish from mObject in that
	// top level object is still fully serialized as object even when all its sub
	// properties are unassigned, whereas mObject is serialized as NULL when it is
	// considered unassigned (all sub properties are unassigned).
	mTop
)
