package json

import "bytes"

type serializer struct {
	// buffer
	bytes.Buffer
	// scratch buffer to assist in number conversions
	scratch [64]byte
	// stack to keep track of the index of the current element among its context
	elementIndexes []int
	// stack to keep track of the contexts
	contexts []int
}
