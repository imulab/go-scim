package v2

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"time"
)

// Construct a new resource unmarshaler, which could be feed to the unmarshal mechanism of the mongo driver.
func newResourceUnmarshaler(resourceType *spec.ResourceType) *deserializer {
	resource := prop.NewResource(resourceType)
	navigator := resource.Navigator()
	return &deserializer{
		resource:  resource,
		navigator: navigator,
	}
}

type deserializer struct {
	resource  *prop.Resource
	navigator prop.Navigator
}

// Get the de-serialized resource. This should only be called after UnmarshalBSON has been called.
func (d *deserializer) Resource() *prop.Resource {
	return d.resource
}

func (d *deserializer) UnmarshalBSON(raw []byte) error {
	vr := bsonrw.NewBSONDocumentReader(raw)
	return d.deserializeComplex(vr, true)
}

func (d *deserializer) deserializeComplex(vr bsonrw.ValueReader, isTopLevel bool) error {
	// ensure property type
	p := d.navigator.Current()
	if p.Attribute().MultiValued() || p.Attribute().Type() != spec.TypeComplex {
		return d.errPropertyType(spec.TypeComplex.String(), d.describeType(p.Attribute()))
	}

	// ensure value type is document
	if !isTopLevel && vr.Type() != bsontype.EmbeddedDocument {
		if vr.Type() == bsontype.Null {
			_ = vr.ReadNull()
			_, _ = d.navigator.Current().Delete()
			return nil
		}
		return d.errInvalidDocType(bsontype.EmbeddedDocument, vr.Type())
	}

	dr, err := vr.ReadDocument()
	if err != nil {
		return d.errReadDoc(err)
	}

	for {
		name, evr, err := dr.ReadElement()
		if err != nil {
			if err == bsonrw.ErrEOD {
				break
			}
			return d.errReadDoc(err)
		}

		// special case, skip over the MongoDB internal id
		if name == "_id" {
			_, err = evr.ReadObjectID()
			if err != nil {
				return err
			}
			continue
		}

		var subProp prop.Property
		{
			// First try to directly focus with the name from MongoDB.
			// We try using navigator.Current().ChildAtIndex because using navigator.Dot would
			// potentially corrupt the navigator by registering a permanent error, preventing
			// us to retry if failed.
			//
			// If failed (err != nil), try to find a sub attribute whose had registered the matching
			// MongoDB field alias in its metadata. If found, focus using the name of the sub attribute.
			//
			// If failed again, return the true error
			if _, err := d.navigator.Current().ChildAtIndex(name); err == nil {
				subProp = d.navigator.Dot(name).Current()
			} else {
				// if failed, try to find a sub attribute who has a registered MongoDB attribute extension
				// that matches the name from MongoDB, and focus using the name of that sub attribute.
				if subAttr := p.Attribute().FindSubAttribute(func(subAttr *spec.Attribute) bool {
					if md, ok := metadataHub[subAttr.ID()]; !ok {
						return false
					} else {
						return md.MongoName == name
					}
				}); subAttr != nil {
					subProp = d.navigator.Dot(subAttr.Name()).Current()
					if d.navigator.HasError() {
						return d.navigator.Error()
					}
				} else {
					return d.errReadDoc(err)
				}
			}
		}

		if subProp.Attribute().MultiValued() {
			err = d.deserializeMultiValued(evr)
		} else {
			err = d.deserializeSingleValued(evr)
		}
		if err != nil {
			return err
		}

		d.navigator.Retract()
	}

	return nil
}

func (d *deserializer) deserializeSingleValued(vr bsonrw.ValueReader) error {
	switch d.navigator.Current().Attribute().Type() {
	case spec.TypeString, spec.TypeReference, spec.TypeBinary:
		return d.deserializeString(vr)
	case spec.TypeDateTime:
		return d.deserializeDateTime(vr)
	case spec.TypeInteger:
		return d.deserializeInteger(vr)
	case spec.TypeDecimal:
		return d.deserializeDecimal(vr)
	case spec.TypeBoolean:
		return d.deserializeBoolean(vr)
	case spec.TypeComplex:
		return d.deserializeComplex(vr, false)
	default:
		panic("invalid attribute type")
	}
}

func (d *deserializer) deserializeMultiValued(vr bsonrw.ValueReader) error {
	// ensure property type
	p := d.navigator.Current()
	if !p.Attribute().MultiValued() {
		return d.errPropertyType("multiValued", "singular")
	}

	// ensure value type is array
	if vr.Type() != bsontype.Array {
		if vr.Type() == bsontype.Null {
			_ = vr.ReadNull()
			_, _ = d.navigator.Current().Delete()
			return nil
		}
		return d.errInvalidDocType(bsontype.Array, vr.Type())
	}

	ar, err := vr.ReadArray()
	if err != nil {
		return d.errReadDoc(err)
	}

	for {
		evr, err := ar.ReadValue()
		if err != nil {
			if err == bsonrw.ErrEOA {
				break
			}
			return d.errReadDoc(err)
		}

		// Construct and focus on a new element prototype to accept the new value
		var elemProp prop.Property
		{
			if mv, ok := d.navigator.Current().(interface {
				AppendElement() int
			}); !ok {
				return fmt.Errorf("%w: expect property to implement AppendElement", spec.ErrInternal)
			} else {
				idx := mv.AppendElement()
				elemProp = d.navigator.At(idx).Current()
				if d.navigator.HasError() {
					return d.navigator.Error()
				}
			}
		}

		if elemProp.Attribute().MultiValued() {
			err = d.deserializeMultiValued(evr)
		} else {
			err = d.deserializeSingleValued(evr)
		}
		if err != nil {
			return err
		}

		d.navigator.Retract()
	}

	return nil
}

func (d *deserializer) deserializeString(vr bsonrw.ValueReader) error {
	// ensure property type
	p := d.navigator.Current()
	if p.Attribute().MultiValued() || (p.Attribute().Type() != spec.TypeString &&
		p.Attribute().Type() != spec.TypeReference &&
		p.Attribute().Type() != spec.TypeBinary) {
		return d.errPropertyType("string|reference|binary", d.describeType(p.Attribute()))
	}

	// ensure value type is string
	if vr.Type() != bsontype.String {
		if vr.Type() == bsontype.Null {
			_ = vr.ReadNull()
			_, _ = d.navigator.Current().Delete()
			return nil
		}
		return d.errInvalidDocType(bsontype.String, vr.Type())
	}

	value, err := vr.ReadString()
	if err != nil {
		return d.errReadDoc(err)
	}

	if _, err := d.navigator.Current().Replace(value); err != nil {
		return err
	}

	return nil
}

func (d *deserializer) deserializeDateTime(vr bsonrw.ValueReader) error {
	// ensure property type
	p := d.navigator.Current()
	if p.Attribute().MultiValued() || p.Attribute().Type() != spec.TypeDateTime {
		return d.errPropertyType(spec.TypeDateTime.String(), d.describeType(p.Attribute()))
	}

	// ensure value type is dateTime
	if vr.Type() != bsontype.DateTime {
		if vr.Type() == bsontype.Null {
			_ = vr.ReadNull()
			_, _ = d.navigator.Current().Delete()
			return nil
		}
		return d.errInvalidDocType(bsontype.DateTime, vr.Type())
	}

	milliSeconds, err := vr.ReadDateTime()
	if err != nil {
		return d.errReadDoc(err)
	}

	t := time.Unix(0, milliSeconds*int64(time.Millisecond))

	if _, err := d.navigator.Current().Replace(t.Format(spec.ISO8601)); err != nil {
		return err
	}

	return nil
}

func (d *deserializer) deserializeInteger(vr bsonrw.ValueReader) error {
	// ensure property type
	p := d.navigator.Current()
	if p.Attribute().MultiValued() || p.Attribute().Type() != spec.TypeInteger {
		return d.errPropertyType(spec.TypeInteger.String(), d.describeType(p.Attribute()))
	}

	// ensure value type is int64
	if vr.Type() != bsontype.Int64 {
		if vr.Type() == bsontype.Null {
			_ = vr.ReadNull()
			_, _ = d.navigator.Current().Delete()
			return nil
		}
		return d.errInvalidDocType(bsontype.Int64, vr.Type())
	}

	i64, err := vr.ReadInt64()
	if err != nil {
		return d.errReadDoc(err)
	}

	if _, err := d.navigator.Current().Replace(i64); err != nil {
		return err
	}

	return nil
}

func (d *deserializer) deserializeDecimal(vr bsonrw.ValueReader) error {
	// ensure property type
	p := d.navigator.Current()
	if p.Attribute().MultiValued() || p.Attribute().Type() != spec.TypeDecimal {
		return d.errPropertyType(spec.TypeDecimal.String(), d.describeType(p.Attribute()))
	}

	// ensure value type is double
	if vr.Type() != bsontype.Double {
		if vr.Type() == bsontype.Null {
			_ = vr.ReadNull()
			_, _ = d.navigator.Current().Delete()
			return nil
		}
		return d.errInvalidDocType(bsontype.Double, vr.Type())
	}

	f64, err := vr.ReadDouble()
	if err != nil {
		return d.errReadDoc(err)
	}

	if _, err := d.navigator.Current().Replace(f64); err != nil {
		return err
	}

	return nil
}

func (d *deserializer) deserializeBoolean(vr bsonrw.ValueReader) error {
	// ensure property type
	p := d.navigator.Current()
	if p.Attribute().MultiValued() || p.Attribute().Type() != spec.TypeBoolean {
		return d.errPropertyType(spec.TypeBoolean.String(), d.describeType(p.Attribute()))
	}

	// ensure value type is boolean
	if vr.Type() != bsontype.Boolean {
		if vr.Type() == bsontype.Null {
			_ = vr.ReadNull()
			_, _ = d.navigator.Current().Delete()
			return nil
		}
		return d.errInvalidDocType(bsontype.Boolean, vr.Type())
	}

	b, err := vr.ReadBoolean()
	if err != nil {
		return d.errReadDoc(err)
	}

	if _, err := d.navigator.Current().Replace(b); err != nil {
		return err
	}

	return nil
}

func (d *deserializer) errInvalidDocType(expect, actual bsontype.Type) error {
	return fmt.Errorf("%w: unexpected type in document: expect %s, got %s",
		spec.ErrInternal, bsonTypeToString(expect), bsonTypeToString(actual))
}

func (d *deserializer) errPropertyType(expect, actual string) error {
	return fmt.Errorf("%w: unexpected property type: expect %s, got %s", spec.ErrInternal, expect, actual)
}

func (d *deserializer) errReadDoc(err error) error {
	return fmt.Errorf("%w: failed to read MongoDB document due to %v", spec.ErrInternal, err)
}

func bsonTypeToString(t bsontype.Type) string {
	switch t {
	case bsontype.Double:
		return "double"
	case bsontype.String:
		return "string"
	case bsontype.EmbeddedDocument:
		return "doc"
	case bsontype.Array:
		return "array"
	case bsontype.Binary:
		return "binary"
	case bsontype.Undefined:
		return "undefined"
	case bsontype.ObjectID:
		return "object_id"
	case bsontype.Boolean:
		return "boolean"
	case bsontype.DateTime:
		return "dateTime"
	case bsontype.Null:
		return "null"
	case bsontype.Regex:
		return "regex"
	case bsontype.DBPointer:
		return "dbPointer"
	case bsontype.JavaScript:
		return "javascript"
	case bsontype.Symbol:
		return "symbol"
	case bsontype.CodeWithScope:
		return "codeWithScope"
	case bsontype.Int32:
		return "int32"
	case bsontype.Timestamp:
		return "timestamp"
	case bsontype.Int64:
		return "int64"
	case bsontype.Decimal128:
		return "decimal128"
	case bsontype.MinKey:
		return "minKey"
	case bsontype.MaxKey:
		return "maxKey"
	default:
		return "?"
	}
}

func (d *deserializer) describeType(attr *spec.Attribute) string {
	if attr.MultiValued() {
		return fmt.Sprintf("multiValued %s", attr.Type().String())
	}
	return attr.Type().String()
}

var (
	_ bson.Unmarshaler = (*deserializer)(nil)
)
