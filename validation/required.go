package validation

import (
	. "github.com/davidiamyou/go-scim/shared"
	"reflect"
	"sync"
)

func ValidateRequired(subj *Resource, sch *Schema) (err error) {
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

	requiredValidatorInstance.validateRequiredWithReflection(reflect.ValueOf(subj.Complex), sch.ToAttribute())

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

func (rv *requiredValidator) validateRequiredWithReflection(v reflect.Value, attr *Attribute) {
	if !v.IsValid() {
		rv.checkValue(v, attr)
		return
	}

	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		rv.checkValue(v, attr)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		rv.checkValue(v, attr)

	case reflect.Float32, reflect.Float64:
		rv.checkValue(v, attr)

	case reflect.Bool:
		rv.checkValue(v, attr)

	case reflect.Slice, reflect.Array:
		rv.checkValue(v, attr)
		if attr.ExpectsComplexArray() {
			subAttr := attr.Clone()
			subAttr.MultiValued = false
			for i := 0; i < v.Len(); i++ {
				rv.validateRequiredWithReflection(v.Index(i), subAttr)
			}
		}

	case reflect.Map:
		rv.checkValue(v, attr)
		for _, k := range v.MapKeys() {
			p, err := NewPath(k.String())
			if err != nil {
				rv.throw(err)
			}
			subAttr := attr.GetAttribute(p, false)
			if subAttr == nil {
				rv.throw(Error.NoAttribute(p.Value()))
			}
			rv.validateRequiredWithReflection(v.MapIndex(k), subAttr)
		}
	}
}

func (rv *requiredValidator) checkValue(v reflect.Value, attr *Attribute) {
	if attr.Required && !attr.Assigned(v) {
		switch attr.Mutability {
		case ReadOnly:
			// unassigned, required, readOnly property is allowed
		case Immutable:
			// nil, required, immutable property is allowed
			if v.IsValid() {
				rv.throw(Error.MissingRequiredProperty(attr.Assist.FullPath))
			}
		default:
			rv.throw(Error.MissingRequiredProperty(attr.Assist.FullPath))
		}
	}
}

func (rv *requiredValidator) throw(err error) {
	panic(err)
}
