package json

import (
	"github.com/davidiamyou/go-scim/resource"
	parser "github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"reflect"
	"fmt"
)

var (
	ErrorJson = errors.New("Malformed JSON")
)

type ScimDecoder struct {
	Attributes 		resource.AttributeVault
	Factory 		resource.Factory
}

func (dec *ScimDecoder) Decode(data []byte) (interface{}, error) {
	opt := &decodeOpt{
		schemas: []string{},
	}

	schemaCollector := func(value []byte, dataType parser.ValueType, offset int, err error){
		opt.schemas = append(opt.schemas, string(value))
	}

	if _, err := parser.ArrayEach(data, schemaCollector, "schemas"); err != nil {
		return nil, ErrorJson
	} else if v, err := dec.Factory.NewResource(opt.schemas); err != nil {
		return nil, err
	} else {
		return dec.decode(data, reflect.ValueOf(v), opt)
	}
}

func (dec *ScimDecoder) decode(data []byte, v reflect.Value, opt *decodeOpt) (interface{}, error) {
	if err := dec.forEachTag(v, func(field reflect.StructField, tag string) error {
		if attr, err := dec.Attributes.Get(tag); err != nil {
			return err
		} else {

			// TODO if attribute's requiredSchema isn't present, skip this.

			if attr.IsSimpleField() {
				if val, err := dec.attemptGetValue(data, attr); err == nil && val.IsValid() {
					dec.setValue(v, field, val)
				} else {
					return err
				}
			} else if attr.IsSimpleArray() {
				if val, err := dec.attemptGetArrayValue(data, attr); err == nil && val.IsValid() {
					dec.setValue(v, field, val)
				} else {
					return err
				}
			} else if attr.IsObject() {

			} else if attr.IsObjectArray() {

			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return v, nil
}

// Traverse through all possible aliases of the field and see if the incoming json bytes contain
// such a value. Returns the first match. If there's no match and the path is not required, return
// zero value, otherwise return RequiredPathNotFoundError.
func (dec *ScimDecoder) attemptGetValue(data []byte, attr *resource.Attribute) (reflect.Value, error) {
	for _, key := range attr.Guide.Aliases {
		switch attr.Type {
		case resource.TypeString, resource.TypeReference, resource.TypeDateTime, resource.TypeBinary:
			if str, err := parser.GetString(data, key); err == nil {
				return reflect.ValueOf(str), nil
			}
		case resource.TypeInteger:
			if i64, err := parser.GetInt(data, key); err == nil {
				return reflect.ValueOf(i64), nil
			}
		case resource.TypeDecimal:
			if f64, err := parser.GetFloat(data, key); err == nil {
				return reflect.ValueOf(f64), nil
			}
		case resource.TypeBoolean:
			if b, err := parser.GetBoolean(data, key); err == nil {
				return reflect.ValueOf(b), nil
			}
		}
	}

	if attr.Required {
		return reflect.Value{}, &RequiredPathNotFoundError{Path: attr.Guide.Tag}
	}

	return reflect.Value{}, nil
}

func (dec *ScimDecoder) attemptGetArrayValue(data []byte, attr *resource.Attribute) (reflect.Value, error) {
	for _, key := range attr.Guide.Aliases {
		switch attr.Type {
		case resource.TypeString, resource.TypeReference, resource.TypeDateTime, resource.TypeBinary:
			strArray := []string{}
			if _, err := parser.ArrayEach(data, func(value []byte, dataType parser.ValueType, offset int, err error){
				strArray = append(strArray, string(value))
			}, key); err == nil {
				return reflect.ValueOf(strArray), nil
			}
		case resource.TypeInteger:
			// TODO
		case resource.TypeDecimal:
			// TODO
		case resource.TypeBoolean:
			// TODO
		}
	}

	if attr.Required {
		return reflect.Value{}, &RequiredPathNotFoundError{Path: attr.Guide.Tag}
	}

	return reflect.Value{}, nil
}

func (dec *ScimDecoder) setValue(base reflect.Value, field reflect.StructField, val reflect.Value) {
	fmt.Println(field.Name + " set")
	if base.Elem().FieldByName(field.Name).CanSet() {
		base.Elem().FieldByName(field.Name).Set(val)
	}
}

func (dec *ScimDecoder) forEachTag(v reflect.Value, cb func(f reflect.StructField, tag string) error) error {
	if v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var t reflect.Type = v.Type()
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if scimTag, ok := f.Tag.Lookup("scim"); ok {
				if err := cb(f, scimTag); err != nil {
					return err
				}
			} else if f.Type.Kind() == reflect.Struct {
				if err := dec.forEachTag(v.Field(i), cb); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type decodeOpt struct {
	schemas 	[]string
}

type RequiredPathNotFoundError struct {
	Path 	string
}

func (e *RequiredPathNotFoundError) Error() string {
	return fmt.Sprintf("No value was found at required path %s", e.Path)
}