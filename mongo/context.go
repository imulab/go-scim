package mongo

// Context preserves vital information within the traversal stack frame during BSON serialization
type context struct {
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
)
