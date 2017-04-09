package validation

import (
	. "github.com/davidiamyou/go-scim/shared"
	"reflect"
	"sync"
)

func ValidateType(subj *Resource, sch *Schema) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case error:
				err = r.(error)
			default:
				err = ErrorCentral.Text("%v", r)
			}
		}
	}()

	typeValidatorInstance.validateTypeWithReflection(reflect.ValueOf(subj.Complex), sch.ToAttribute())
	err = nil
	return
}

var (
	singleTypeValidator	sync.Once
	typeValidatorInstance 	*typeValidator
)

func init() {
	singleTypeValidator.Do(func() {
		typeValidatorInstance = &typeValidator{}
	})
}

type typeValidator struct {}

func (tv *typeValidator) validateTypeWithReflection(v reflect.Value, attr *Attribute) {
	if attr.Mutability == ReadOnly {
		return
	}

	if !v.IsValid() {
		return
	}

	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		if !attr.ExpectsString() {
			panic(ErrorCentral.InvalidType(attr.Assist.FullPath, TypeString, v.Type().Name()))
		}

	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		if !attr.ExpectsInteger() {
			panic(ErrorCentral.InvalidType(attr.Assist.FullPath, TypeInteger, v.Type().Name()))
		}

	case reflect.Float32, reflect.Float64:
		if !attr.ExpectsFloat() {
			panic(ErrorCentral.InvalidType(attr.Assist.FullPath, TypeDecimal, v.Type().Name()))
		}

	case reflect.Bool:
		if !attr.ExpectsBool() {
			panic(ErrorCentral.InvalidType(attr.Assist.FullPath, TypeBoolean, v.Type().Name()))
		}

	case reflect.Array, reflect.Slice:
		if !attr.MultiValued {
			panic(ErrorCentral.InvalidType(attr.Assist.FullPath, "array", v.Type().Name()))
		}

		subAttr := attr.Clone()
		subAttr.MultiValued = false
		for i := 0; i < v.Len(); i++ {
			tv.validateTypeWithReflection(v.Index(i), subAttr)
		}

	case reflect.Map:
		if !attr.ExpectsComplex() {
			panic(ErrorCentral.InvalidType(attr.Assist.FullPath, TypeComplex, v.Type().Name()))
		}

		for _, k := range v.MapKeys() {
			p, err := NewPath(k.String())
			if err != nil {
				panic(err)
			}

			subAttr := attr.GetAttribute(p, false)
			if subAttr == nil {
				panic(ErrorCentral.NoAttribute(p.Value()))
			}

			tv.validateTypeWithReflection(v.MapIndex(k), subAttr)
		}

	default:
		panic(ErrorCentral.InvalidType(attr.Assist.FullPath, "unhandled type", v.Type().Name()))
	}
}
