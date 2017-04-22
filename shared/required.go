package shared

import (
	"reflect"
	"sync"
	"context"
)

func ValidateRequired(subj *Resource, sch *Schema, ctx context.Context) (err error) {
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

	requiredValidatorInstance.validateRequiredWithReflection(reflect.ValueOf(subj.Complex), sch.ToAttribute(), ctx)

	err = nil
	return
}

var (
	singleRequiredValidator   sync.Once
	requiredValidatorInstance *requiredValidator
)

func init() {
	singleRequiredValidator.Do(func() {
		requiredValidatorInstance = &requiredValidator{}
	})
}

type requiredValidator struct{}

func (rv *requiredValidator) validateRequiredWithReflection(v reflect.Value, attr *Attribute, ctx context.Context) {
	if !v.IsValid() {
		rv.checkValue(v, attr, ctx)
		return
	}

	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		rv.checkValue(v, attr, ctx)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rv.checkValue(v, attr, ctx)

	case reflect.Float32, reflect.Float64:
		rv.checkValue(v, attr, ctx)

	case reflect.Bool:
		rv.checkValue(v, attr, ctx)

	case reflect.Slice, reflect.Array:
		rv.checkValue(v, attr, ctx)
		if attr.ExpectsComplexArray() {
			subAttr := attr.Clone()
			subAttr.MultiValued = false
			for i := 0; i < v.Len(); i++ {
				rv.validateRequiredWithReflection(v.Index(i), subAttr, ctx)
			}
		}

	case reflect.Map:
		rv.checkValue(v, attr, ctx)
		for _, k := range v.MapKeys() {
			p, err := NewPath(k.String())
			if err != nil {
				rv.throw(err, ctx)
			}
			subAttr := attr.GetAttribute(p, false)
			if subAttr == nil {
				rv.throw(Error.NoAttribute(p.Value()), ctx)
			}
			rv.validateRequiredWithReflection(v.MapIndex(k), subAttr, ctx)
		}
	}
}

func (rv *requiredValidator) checkValue(v reflect.Value, attr *Attribute, ctx context.Context) {
	if attr.Required && !attr.Assigned(v) {
		switch attr.Mutability {
		case ReadOnly:
			// unassigned, required, readOnly property is allowed
		case Immutable:
			// nil, required, immutable property is allowed
			if v.IsValid() {
				rv.throw(Error.MissingRequiredProperty(attr.Assist.FullPath), ctx)
			}
		default:
			rv.throw(Error.MissingRequiredProperty(attr.Assist.FullPath), ctx)
		}
	}
}

func (rv *requiredValidator) throw(err error, ctx context.Context) {
	panic(err)
}
