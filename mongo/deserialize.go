package mongo

import (
	"fmt"
	"github.com/imulab/go-scim/core"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"time"
)

// Construct a new resource unmarshaler, which could be feed to the unmarshal mechanism of the mongo driver.
func newResourceUnmarshaler(resourceType *core.ResourceType) *resourceUnmarshaler {
	resource := core.Resources.New(resourceType)
	navigator := core.NewNavigator(resource)
	return &resourceUnmarshaler{
		resource:  resource,
		navigator: navigator,
	}
}

// implementation of bson.Unmarshaler to circumvent the driver mechanism of decoder lookup and take
// full control of the deserialization process
type resourceUnmarshaler struct {
	resource  *core.Resource
	navigator core.Navigator
}

func (m *resourceUnmarshaler) GetResource() *core.Resource {
	return m.resource
}

func (m *resourceUnmarshaler) UnmarshalBSON(raw []byte) error {
	vr := bsonrw.NewBSONDocumentReader(raw)
	return m.unmarshalComplexProperty(vr, true)
}

func (m *resourceUnmarshaler) unmarshalComplexProperty(vr bsonrw.ValueReader, isTopLevel bool) error {
	// ensure property type
	prop := m.navigator.Current()
	if prop.Attribute().MultiValued || prop.Attribute().Type != core.TypeComplex {
		return m.errPropertyType(core.TypeComplex.String(), prop.Attribute().DescribeType())
	}

	// ensure value type is document
	if !isTopLevel && vr.Type() != bsontype.EmbeddedDocument {
		if vr.Type() == bsontype.Null {
			_ = vr.ReadNull()
			_ = m.navigator.Delete()
			return nil
		}
		return m.errInvalidDocType(bsontype.EmbeddedDocument, vr.Type())
	}

	dr, err := vr.ReadDocument()
	if err != nil {
		return m.errReadDoc(err)
	}

	for {
		name, evr, err := dr.ReadElement()
		if err != nil {
			if err == bsonrw.ErrEOD {
				break
			}
			return m.errReadDoc(err)
		}

		var subProp core.Property
		{
			// first try to directly focus with the dbName, if failed, try match with
			// dbAlias name of one sub attribute
			subProp, err = m.navigator.Focus(name)
			if err != nil {
				for _, subAttr := range prop.Attribute().SubAttributes {
					metadata := core.Meta.Get(subAttr.Id, MongoMetadataId)
					if metadata == nil {
						continue
					}

					if metadata.(*Metadata).DbAlias == name {
						subProp, err = m.navigator.Focus(subAttr.Name)
						break
					}
				}
				if err != nil {
					return m.errReadDoc(err)
				}
			}
		}

		if subProp.Attribute().MultiValued {
			err = m.unmarshalMultiValuedProperty(evr)
		} else {
			err = m.unmarshalSingleValuedProperty(evr)
		}
		if err != nil {
			return err
		}

		m.navigator.Release()
	}

	return nil
}

func (m *resourceUnmarshaler) unmarshalSingleValuedProperty(vr bsonrw.ValueReader) error {
	switch m.navigator.Current().Attribute().Type {
	case core.TypeString, core.TypeReference, core.TypeBinary:
		return m.unmarshalCommonStringProperty(vr)
	case core.TypeDateTime:
		return m.unmarshalDateTimeProperty(vr)
	case core.TypeInteger:
		return m.unmarshalIntegerProperty(vr)
	case core.TypeDecimal:
		return m.unmarshalDecimalProperty(vr)
	case core.TypeBoolean:
		return m.unmarshalBooleanProperty(vr)
	case core.TypeComplex:
		return m.unmarshalComplexProperty(vr, false)
	default:
		panic("invalid attribute type")
	}
}

func (m *resourceUnmarshaler) unmarshalMultiValuedProperty(vr bsonrw.ValueReader) error {
	// ensure property type
	prop := m.navigator.Current()
	if !prop.Attribute().MultiValued {
		return m.errPropertyType("multiValued", "singular")
	}

	// ensure value type is array
	if vr.Type() != bsontype.Array {
		if vr.Type() == bsontype.Null {
			_ = vr.ReadNull()
			_ = m.navigator.Delete()
			return nil
		}
		return m.errInvalidDocType(bsontype.Array, vr.Type())
	}

	ar, err := vr.ReadArray()
	if err != nil {
		return m.errReadDoc(err)
	}

	for {
		evr, err := ar.ReadValue()
		if err != nil {
			if err == bsonrw.ErrEOA {
				break
			}
			return m.errReadDoc(err)
		}

		// Construct and focus on a new element prototype to accept the new value
		idx := m.navigator.NewPrototype()
		elemProp, err := m.navigator.Focus(idx)
		if err != nil {
			return err
		}

		if elemProp.Attribute().MultiValued {
			err = m.unmarshalMultiValuedProperty(evr)
		} else {
			err = m.unmarshalSingleValuedProperty(evr)
		}
		if err != nil {
			return err
		}

		m.navigator.Release()
	}

	return nil
}

func (m *resourceUnmarshaler) unmarshalCommonStringProperty(vr bsonrw.ValueReader) error {
	// ensure property type
	prop := m.navigator.Current()
	if prop.Attribute().MultiValued || (
		prop.Attribute().Type != core.TypeString &&
			prop.Attribute().Type != core.TypeReference &&
			prop.Attribute().Type != core.TypeBinary) {
		return m.errPropertyType("string|reference|binary", prop.Attribute().DescribeType())
	}

	// ensure value type is string
	if vr.Type() != bsontype.String {
		if vr.Type() == bsontype.Null {
			_ = vr.ReadNull()
			_ = m.navigator.Delete()
			return nil
		}
		return m.errInvalidDocType(bsontype.String, vr.Type())
	}

	value, err := vr.ReadString()
	if err != nil {
		return m.errReadDoc(err)
	}

	return m.navigator.Replace(value)
}

func (m *resourceUnmarshaler) unmarshalDateTimeProperty(vr bsonrw.ValueReader) error {
	// ensure property type
	prop := m.navigator.Current()
	if prop.Attribute().MultiValued || prop.Attribute().Type != core.TypeDateTime {
		return m.errPropertyType(core.TypeDateTime.String(), prop.Attribute().DescribeType())
	}

	// ensure value type is dateTime
	if vr.Type() != bsontype.DateTime {
		if vr.Type() == bsontype.Null {
			_ = vr.ReadNull()
			_ = m.navigator.Delete()
			return nil
		}
		return m.errInvalidDocType(bsontype.DateTime, vr.Type())
	}

	milliSeconds, err := vr.ReadDateTime()
	if err != nil {
		return m.errReadDoc(err)
	}

	t := time.Unix(0, milliSeconds*int64(time.Millisecond))
	return m.navigator.Replace(t.Format(core.ISO8601))
}

func (m *resourceUnmarshaler) unmarshalIntegerProperty(vr bsonrw.ValueReader) error {
	// ensure property type
	prop := m.navigator.Current()
	if prop.Attribute().MultiValued || prop.Attribute().Type != core.TypeInteger {
		return m.errPropertyType(core.TypeInteger.String(), prop.Attribute().DescribeType())
	}

	// ensure value type is int64
	if vr.Type() != bsontype.Int64 {
		if vr.Type() == bsontype.Null {
			_ = vr.ReadNull()
			_ = m.navigator.Delete()
			return nil
		}
		return m.errInvalidDocType(bsontype.Int64, vr.Type())
	}

	i64, err := vr.ReadInt64()
	if err != nil {
		return m.errReadDoc(err)
	}

	return m.navigator.Replace(i64)
}

func (m *resourceUnmarshaler) unmarshalDecimalProperty(vr bsonrw.ValueReader) error {
	// ensure property type
	prop := m.navigator.Current()
	if prop.Attribute().MultiValued || prop.Attribute().Type != core.TypeDecimal {
		return m.errPropertyType(core.TypeDecimal.String(), prop.Attribute().DescribeType())
	}

	// ensure value type is double
	if vr.Type() != bsontype.Double {
		if vr.Type() == bsontype.Null {
			_ = vr.ReadNull()
			_ = m.navigator.Delete()
			return nil
		}
		return m.errInvalidDocType(bsontype.Double, vr.Type())
	}

	f64, err := vr.ReadDouble()
	if err != nil {
		return m.errReadDoc(err)
	}

	return m.navigator.Replace(f64)
}

func (m *resourceUnmarshaler) unmarshalBooleanProperty(vr bsonrw.ValueReader) error {
	// ensure property type
	prop := m.navigator.Current()
	if prop.Attribute().MultiValued || prop.Attribute().Type != core.TypeBoolean {
		return m.errPropertyType(core.TypeBoolean.String(), prop.Attribute().DescribeType())
	}

	// ensure value type is boolean
	if vr.Type() != bsontype.Boolean {
		if vr.Type() == bsontype.Null {
			_ = vr.ReadNull()
			_ = m.navigator.Delete()
			return nil
		}
		return m.errInvalidDocType(bsontype.Boolean, vr.Type())
	}

	b, err := vr.ReadBoolean()
	if err != nil {
		return m.errReadDoc(err)
	}

	return m.navigator.Replace(b)
}

func (m *resourceUnmarshaler) errInvalidDocType(expect, actual bsontype.Type) error {
	return core.Errors.Internal(fmt.Sprintf("unexpected type in document: expect %s, got %s",
		m.bsonTypeToString(expect), m.bsonTypeToString(actual)))
}

func (m *resourceUnmarshaler) errPropertyType(expect, actual string) error {
	return core.Errors.Internal(fmt.Sprintf("unexpected property type: expect %s, got %s", expect, actual))
}

func (m *resourceUnmarshaler) errReadDoc(err error) error {
	return core.Errors.Internal(fmt.Sprintf("failed to read MongoDB document: %s", err))
}

func (m *resourceUnmarshaler) bsonTypeToString(t bsontype.Type) string {
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
