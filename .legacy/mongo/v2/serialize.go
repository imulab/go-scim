package v2

import (
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"go.mongodb.org/mongo-driver/bson"
	"math"
	"strconv"
	"time"
)

// Create an adapter to BSON that implements the bson.Marshaler interface so it can be directly
// feed to MongoDB driver methods.
func newBsonAdapter(resource *prop.Resource) bson.Marshaler {
	return &bsonAdapter{resource: resource}
}

// Adapter of resource to bson.Marshaler
type bsonAdapter struct {
	resource *prop.Resource
}

func (d *bsonAdapter) MarshalBSON() ([]byte, error) {
	visitor := &serializer{
		buf:   make([]byte, 0),
		stack: make([]*frame, 0),
	}

	err := d.resource.Visit(visitor)
	if err != nil {
		return nil, err
	}

	return visitor.buf, nil
}

// BSON serializer that implements prop.Visitor interface.
type serializer struct {
	buf []byte
	// Stack to keep track of traversal context.
	// Context is pushed and popped only for container properties.
	stack []*frame
}

func (s *serializer) ShouldVisit(property prop.Property) bool {
	// In case of a top level unassigned complex property, which pushes the state mObject onto the frame,
	// we cannot prevent the complex property itself from calling Visit/BeginComplex/EndComplex,
	// because we still want it to be serialized as NULL. However, we wouldn't want to visit its sub properties,
	// because we already know that are all unassigned.
	if s.current().mode == mObject {
		return !property.IsUnassigned()
	}
	return true
}

func (s *serializer) Visit(property prop.Property) error {
	if property.Attribute().MultiValued() {
		s.serializeMultiProperty(property)
		return nil
	}

	switch property.Attribute().Type() {
	case spec.TypeString, spec.TypeReference, spec.TypeBinary:
		s.serializeStringProperty(property)
	case spec.TypeDateTime:
		s.serializeDateTimeProperty(property)
	case spec.TypeInteger:
		s.serializeIntegerProperty(property)
	case spec.TypeDecimal:
		s.serializeDecimalProperty(property)
	case spec.TypeBoolean:
		s.serializeBooleanProperty(property)
	case spec.TypeComplex:
		s.serializeComplexProperty(property)
	default:
		panic("invalid property type")
	}

	if property.Attribute().Type() != spec.TypeComplex {
		s.current().index++
	}

	return nil
}

func (s *serializer) serializeComplexProperty(property prop.Property) {
	// We only serialize the name and the kind here, because sub properties, if any, will be visited later.
	if property.IsUnassigned() {
		s.addName(0x0A, property.Attribute())
		return
	}

	s.addName(0x03, property.Attribute())
}

func (s *serializer) serializeMultiProperty(property prop.Property) {
	// First off, we only serialize the name and the kind here, because elements, if any, will
	// be visited later.
	//
	// Secondly, we always serialize the multiValued property as an BSON array even if it's empty.
	// This is consistent with the concept of "unassigned" in the specification: an empty
	// multiValued property is considered to be unassigned while all other nulls are considered
	// unassigned.
	//
	// By doing this, we also avoid one other problem: if we serialize an unassigned multiValued
	// property as NULL, we will create a problem where the visitor calls BeginMulti and EndMulti
	// while attempting to traverse an empty array. BeginMulti would have to reserve an int32
	// element length marker, which is not consistent to the NULL type.
	s.addName(0x04, property.Attribute())
}

func (s *serializer) serializeStringProperty(property prop.Property) {
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

func (s *serializer) serializeDateTimeProperty(property prop.Property) {
	if property.IsUnassigned() {
		s.addName(0x0A, property.Attribute())
		return
	}

	s.addName(0x09, property.Attribute())
	// mongodb stores milliseconds
	t, _ := time.Parse(spec.ISO8601, property.Raw().(string))
	s.addInt64(t.Unix()*1000 + int64(t.Nanosecond()/1e6))
}

func (s *serializer) serializeIntegerProperty(property prop.Property) {
	if property.IsUnassigned() {
		s.addName(0x0A, property.Attribute())
		return
	}

	s.addName(0x12, property.Attribute())
	s.addInt64(property.Raw().(int64))
}

func (s *serializer) serializeDecimalProperty(property prop.Property) {
	if property.IsUnassigned() {
		s.addName(0x0A, property.Attribute())
		return
	}

	s.addName(0x01, property.Attribute())
	s.addInt64(int64(math.Float64bits(property.Raw().(float64))))
}

func (s *serializer) serializeBooleanProperty(property prop.Property) {
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

func (s *serializer) BeginChildren(container prop.Property) {
	if container.Attribute().MultiValued() {
		s.push(mArray)
		s.current().start = s.reserveInt32()
	} else {
		if len(s.stack) == 0 {
			s.push(mTop)
		} else {
			s.push(mObject)
		}
		if !container.IsUnassigned() {
			s.current().start = s.reserveInt32()
		}
	}
}

func (s *serializer) EndChildren(container prop.Property) {
	if container.Attribute().MultiValued() {
		s.addBytes(0)
		start := s.current().start
		s.setInt32(start, int32(len(s.buf)-start))
		s.pop()
		s.current().index++
	} else {
		if !container.IsUnassigned() {
			s.addBytes(0)
			start := s.current().start
			s.setInt32(start, int32(len(s.buf)-start))
		}
		s.pop()
		if len(s.stack) > 0 {
			s.current().index++
		}
	}
}

func (s *serializer) addName(kind byte, attr *spec.Attribute) {
	var name string
	{
		switch s.current().mode {
		case mArray:
			name = strconv.Itoa(s.current().index)
		case mObject, mTop:
			name = attr.Name()
			if md, ok := metadataHub[attr.ID()]; ok {
				name = md.MongoName
			}
		}
	}

	s.addBytes(kind)
	s.addBytes([]byte(name)...)
	s.addBytes(0)
}

func (s *serializer) addInt64(v int64) {
	u := uint64(v)
	s.addBytes(byte(u), byte(u>>8), byte(u>>16), byte(u>>24),
		byte(u>>32), byte(u>>40), byte(u>>48), byte(u>>56))
}

// Reserve 4 bytes at current position to write the length of the data in the current frame later.
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
	s.stack = append(s.stack, &frame{
		mode:  m,
		index: 0,
		start: 0,
	})
}

func (s *serializer) pop() {
	s.stack = s.stack[:len(s.stack)-1]
}

func (s *serializer) current() *frame {
	return s.stack[len(s.stack)-1]
}
