package shared

import (
	"reflect"
	"sync"
	"context"
)

func ValidateType(subj *Resource, sch *Schema, ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case error:
				err = r.(error)
			default:
				err = Error.Text("%v", r)
			}
		}
	}()

	typeValidatorInstance.validateTypeWithReflection(reflect.ValueOf(subj.Complex), sch.ToAttribute(), ctx)
	err = nil
	return
}

var (
	singleTypeValidator   sync.Once
	typeValidatorInstance *typeValidator
)

func init() {
	singleTypeValidator.Do(func() {
		typeValidatorInstance = &typeValidator{}
	})
}

type typeValidator struct{}

func (tv *typeValidator) validateTypeWithReflection(v reflect.Value, attr *Attribute, ctx context.Context) {
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
			tv.throw(Error.InvalidType(attr.Assist.FullPath, TypeString, v.Type().Name()), ctx)
		}

	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		if !attr.ExpectsInteger() {
			tv.throw(Error.InvalidType(attr.Assist.FullPath, TypeInteger, v.Type().Name()), ctx)
		}

	case reflect.Float32, reflect.Float64:
		if !attr.ExpectsFloat() {
			tv.throw(Error.InvalidType(attr.Assist.FullPath, TypeDecimal, v.Type().Name()), ctx)
		}

	case reflect.Bool:
		if !attr.ExpectsBool() {
			tv.throw(Error.InvalidType(attr.Assist.FullPath, TypeBoolean, v.Type().Name()), ctx)
		}

	case reflect.Array, reflect.Slice:
		if !attr.MultiValued {
			tv.throw(Error.InvalidType(attr.Assist.FullPath, "array", v.Type().Name()), ctx)
		}

		subAttr := attr.Clone()
		subAttr.MultiValued = false
		for i := 0; i < v.Len(); i++ {
			tv.validateTypeWithReflection(v.Index(i), subAttr, ctx)
		}

	case reflect.Map:
		if !attr.ExpectsComplex() {
			tv.throw(Error.InvalidType(attr.Assist.FullPath, TypeComplex, v.Type().Name()), ctx)
		}

		for _, k := range v.MapKeys() {
			p, err := NewPath(k.String())
			if err != nil {
				tv.throw(err, ctx)
			}

			subAttr := attr.GetAttribute(p, false)
			if subAttr == nil {
				tv.throw(Error.NoAttribute(p.Value()), ctx)
			}

			tv.validateTypeWithReflection(v.MapIndex(k), subAttr, ctx)
		}

	default:
		tv.throw(Error.InvalidType(attr.Assist.FullPath, "unhandled type", v.Type().Name()), ctx)
	}
}

func (tv *typeValidator) throw(err error, ctx context.Context) {
	panic(err)
}
