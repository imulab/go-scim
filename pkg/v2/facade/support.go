package facade

import (
	"errors"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"reflect"
	"strings"
)

var (
	ErrNilInput           = errors.New("the input object is nil")
	ErrInputType          = errors.New("the input object has a wrong type")
	ErrDisallowedOperator = errors.New("a filter contains disallowed operators")
	ErrSCIMPath           = errors.New("the input object contains an invalid SCIM path")
)

func forEachMapping(target reflect.Value, callback func(field reflect.Value, path string) error) error {
	if target.Kind() == reflect.Ptr {
		if target.IsNil() {
			return ErrNilInput
		}
		return forEachMapping(target.Elem(), callback)
	}

	if target.Kind() != reflect.Struct {
		return ErrInputType
	}

	for i := 0; i < target.NumField(); i++ {
		scimTag, ok := target.Type().Field(i).Tag.Lookup("scim")
		if !ok {
			continue
		}

		paths := strings.FieldsFunc(scimTag, func(r rune) bool { return r == ',' })
		for _, path := range paths {
			if err := callback(target.Field(i), path); err != nil {
				return err
			}
		}
	}

	return nil
}

func typeCheck(attr *spec.Attribute, t reflect.Type) error {
	switch t.Kind() {
	case reflect.String:
		switch attr.Type() {
		case spec.TypeString, spec.TypeReference, spec.TypeBinary:
			return nil
		}
	case reflect.Int64:
		switch attr.Type() {
		case spec.TypeInteger, spec.TypeDateTime:
			return nil
		}
	case reflect.Float64:
		if spec.TypeDecimal == attr.Type() {
			return nil
		}
	case reflect.Bool:
		if spec.TypeBoolean == attr.Type() {
			return nil
		}
	case reflect.Ptr:
		return typeCheck(attr, t.Elem())
	case reflect.Slice:
		if attr.MultiValued() {
			return typeCheck(attr.DeriveElementAttribute(), t.Elem())
		}
	}

	return ErrInputType
}
