package mongo

import (
	"github.com/imulab/go-scim/core"
	"math"
	"strconv"
	"time"
)

// Entry point to serialize the resource to BSON that can be saved in MongoDB.
func Serialize(resource *core.Resource) ([]byte, error) {
	visitor := &serializer{
		buf:   make([]byte, 0),
		stack: make([]*context, 0),
	}
	visitor.push(mObject)

	err := resource.Visit(visitor)
	if err != nil {
		return nil, err
	}

	return visitor.buf, nil
}

// BSON serializer that implements core.Visitor interface.
type serializer struct {
	buf []byte
	// Stack to keep track of traversal context.
	// Context is pushed and popped only for container properties.
	stack []*context
}

func (s *serializer) ShouldVisit(property core.Property) bool {
	// we want to serialize the complete model
	return true
}

func (s *serializer) Visit(property core.Property) error {
	if property.Attribute().MultiValued {
		s.serializeMultiProperty(property)
		return nil
	}

	switch property.Attribute().Type {
	case core.TypeString, core.TypeReference, core.TypeBinary:
		s.serializeStringProperty(property)
	case core.TypeDateTime:
		s.serializeDateTimeProperty(property)
	case core.TypeInteger:
		s.serializeIntegerProperty(property)
	case core.TypeDecimal:
		s.serializeDecimalProperty(property)
	case core.TypeBoolean:
		s.serializeBooleanProperty(property)
	case core.TypeComplex:
		s.serializeComplexProperty(property)
	default:
		panic("invalid property type")
	}

	if property.Attribute().Type != core.TypeComplex {
		s.current().index++
	}

	return nil
}

func (s *serializer) serializeComplexProperty(property core.Property) {
	if property.IsUnassigned() {
		s.addName(0x0A, property.Attribute())
		return
	}

	s.addName(0x03, property.Attribute())
	// do not add sub property values here because they will be visited later.
}

func (s *serializer) serializeMultiProperty(property core.Property) {
	if property.IsUnassigned() {
		s.addName(0x0A, property.Attribute())
		return
	}

	s.addName(0x04, property.Attribute())
	// do not add element values here because they will be visited later.
}

func (s *serializer) serializeStringProperty(property core.Property) {
	if property.IsUnassigned() {
		s.addName(0x0A, property.Attribute())
		return
	}

	s.addName(0x02, property.Attribute())
	value := property.Raw().(string)
	s.addInt32(int32(len(value) + 1))
	s.addBytes([]byte(value)...)
	s.addBytes(0)
}

func (s *serializer) serializeDateTimeProperty(property core.Property) {
	if property.IsUnassigned() {
		s.addName(0x0A, property.Attribute())
		return
	}

	s.addName(0x09, property.Attribute())
	// mongodb stores milliseconds
	t, _ := time.Parse(core.ISO8601, property.Raw().(string))
	s.addInt64(t.Unix()*1000 + int64(t.Nanosecond()/1e6))
}

func (s *serializer) serializeIntegerProperty(property core.Property) {
	if property.IsUnassigned() {
		s.addName(0x0A, property.Attribute())
		return
	}

	s.addName(0x12, property.Attribute())
	s.addInt64(property.Raw().(int64))
}

func (s *serializer) serializeDecimalProperty(property core.Property) {
	if property.IsUnassigned() {
		s.addName(0x0A, property.Attribute())
		return
	}

	s.addName(0x01, property.Attribute())
	s.addInt64(int64(math.Float64bits(property.Raw().(float64))))
}

func (s *serializer) serializeBooleanProperty(property core.Property) {
	if property.IsUnassigned() {
		s.addName(0x0A, property.Attribute())
		return
	}

	s.addName(0x08, property.Attribute())
	if property.Raw().(bool) {
		s.addBytes(1)
	} else {
		s.addBytes(0)
	}
}

func (s *serializer) addInt64(v int64) {
	u := uint64(v)
	s.addBytes(byte(u), byte(u>>8), byte(u>>16), byte(u>>24),
		byte(u>>32), byte(u>>40), byte(u>>48), byte(u>>56))
}

func (s *serializer) BeginComplex() {
	s.push(mObject)
	s.current().start = s.reserveInt32()
}

func (s *serializer) EndComplex() {
	s.addBytes(0)
	start := s.current().start
	s.setInt32(start, int32(len(s.buf)-start))
	s.pop()
	s.current().index++
}

func (s *serializer) BeginMulti() {
	s.push(mArray)
	s.current().start = s.reserveInt32()
}

func (s *serializer) EndMulti() {
	s.addBytes(0)
	start := s.current().start
	s.setInt32(start, int32(len(s.buf)-start))
	s.pop()
	s.current().index++
}

func (s *serializer) addName(kind byte, attr *core.Attribute) {
	var name string
	{
		switch s.current().mode {
		case mArray:
			name = strconv.Itoa(s.current().index)
		case mObject:
			name = attr.Name
			if attr.Metadata != nil && len(attr.Metadata.DbAlias) > 0 {
				name = attr.Metadata.DbAlias
			}
		}
	}

	s.addBytes(kind)
	s.addBytes([]byte(name)...)
	s.addBytes(0)
}

// Reserve 4 bytes at current position to write the length of the data in the current context later.
func (s *serializer) reserveInt32() (pos int) {
	pos = len(s.buf)
	s.addBytes(0, 0, 0, 0)
	return
}

func (s *serializer) addInt32(v int32) {
	u := uint32(v)
	s.addBytes(byte(u), byte(u>>8), byte(u>>16), byte(u>>24))
}

// Add bytes to buffer
func (s *serializer) addBytes(v ...byte) {
	s.buf = append(s.buf, v...)
}

// Set the int32 value at the previously reserved position from reserveInt32. This shall be called after
// container elements have exhausted and we now know the length of them.
func (s *serializer) setInt32(pos int, v int32) {
	s.buf[pos+0] = byte(v)
	s.buf[pos+1] = byte(v >> 8)
	s.buf[pos+2] = byte(v >> 16)
	s.buf[pos+3] = byte(v >> 24)
}

func (s *serializer) push(m mode) {
	s.stack = append(s.stack, &context{
		mode:  m,
		index: 0,
		start: 0,
	})
}

func (s *serializer) pop() {
	s.stack = s.stack[:len(s.stack)-1]
}

func (s *serializer) current() *context {
	return s.stack[len(s.stack)-1]
}

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
