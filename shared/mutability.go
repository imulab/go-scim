package shared

import (
	"reflect"
)

func ValidateMutability(subj *Resource, ref *Resource, sch *Schema) (err error) {
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

	validator := &mutabilityValidator{
		subjBaseStack: NewStackWithoutLimit(),
		refBaseStack:  NewStackWithoutLimit(),
	}
	validator.stepIn(reflect.ValueOf(subj.Complex), reflect.ValueOf(ref.Complex), sch.ToAttribute())

	err = nil
	return
}

type mutabilityValidator struct {
	subjBaseStack Stack
	refBaseStack  Stack
}

func (mv *mutabilityValidator) stepIn(sv, rv reflect.Value, attr *Attribute) {
	if !attr.Assigned(sv) || !attr.Assigned(rv) {
		return
	}

	switch sv.Kind() {
	case reflect.Interface, reflect.Ptr:
		sv = sv.Elem()
	}

	switch rv.Kind() {
	case reflect.Interface, reflect.Ptr:
		rv = rv.Elem()
	}

	mv.subjBaseStack.Push(sv)
	mv.refBaseStack.Push(rv)
	mv.validateMutabilityWithReflection(attr)
	mv.subjBaseStack.Pop()
	mv.refBaseStack.Pop()
}

func (mv *mutabilityValidator) validateMutabilityWithReflection(guide *Attribute) {
	for _, attr := range guide.SubAttributes {
		subjVal := mv.getSubjectValue(attr)
		refVal := mv.getReferenceValue(attr)

		switch attr.Type {
		case TypeComplex:
			if attr.MultiValued {
				mv.compareAndCopy(subjVal, refVal, attr)
				if attr.Assigned(subjVal) && attr.Assigned(refVal) {
					if subjVal.Kind() == reflect.Interface {
						subjVal = subjVal.Elem()
					}
					if refVal.Kind() == reflect.Interface {
						refVal = refVal.Elem()
					}
					elemAttr := attr.Clone()
					elemAttr.MultiValued = false
					for i := 0; i < subjVal.Len(); i++ {
						for j := 0; j < refVal.Len(); j++ {
							subjElemVal := subjVal.Index(i)
							refElemVal := refVal.Index(j)
							if mv.matches(subjElemVal, refElemVal, elemAttr.Assist.ArrayIndexKey) {
								mv.stepIn(subjElemVal, refElemVal, elemAttr)
							}
						}
					}
				}
			} else {
				mv.compareAndCopy(subjVal, refVal, attr)
				mv.stepIn(subjVal, refVal, attr)
			}
		default:
			mv.compareAndCopy(subjVal, refVal, attr)
		}
	}
}

func (mv *mutabilityValidator) compareAndCopy(sv, rv reflect.Value, attr *Attribute) {
	switch attr.Mutability {
	case ReadOnly:
		baseVal := mv.subjBaseStack.Peek().(reflect.Value)
		baseVal.SetMapIndex(reflect.ValueOf(attr.Name), rv)

	case Immutable:
		if !mv.safeIsNil(rv) {
			if !mv.safeIsNil(sv) {
				if !reflect.DeepEqual(sv.Interface(), rv.Interface()) {
					mv.throw(Error.MutabilityViolation(attr.Assist.FullPath))
				}
			} else {
				mv.throw(Error.MutabilityViolation(attr.Assist.FullPath))
			}
		}
	}
}

func (mv *mutabilityValidator) matches(sv, rv reflect.Value, keys []string) bool {
	if !sv.IsValid() || !rv.IsValid() {
		return false
	}

	switch sv.Kind() {
	case reflect.Interface, reflect.Ptr:
		sv = sv.Elem()
	}

	switch rv.Kind() {
	case reflect.Interface, reflect.Ptr:
		rv = rv.Elem()
	}

	for _, key := range keys {
		svKeyVal := sv.MapIndex(reflect.ValueOf(key))
		rvKeyVal := rv.MapIndex(reflect.ValueOf(key))

		if !svKeyVal.IsValid() || !rvKeyVal.IsValid() {
			return false
		} else if !reflect.DeepEqual(svKeyVal.Interface(), rvKeyVal.Interface()) {
			return false
		}
	}
	return true
}

func (mv *mutabilityValidator) getSubjectValue(attr *Attribute) reflect.Value {
	return mv.getValue(mv.subjBaseStack.Peek().(reflect.Value), attr)
}

func (mv *mutabilityValidator) getReferenceValue(attr *Attribute) reflect.Value {
	return mv.getValue(mv.refBaseStack.Peek().(reflect.Value), attr)
}

func (mv *mutabilityValidator) getValue(v reflect.Value, attr *Attribute) reflect.Value {
	if v.Kind() != reflect.Map {
		panic("invalid use of mutabilityValidator::getValue")
	}
	return v.MapIndex(reflect.ValueOf(attr.Name))
}

func (mv *mutabilityValidator) safeIsNil(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (mv *mutabilityValidator) throw(err error) {
	panic(err)
}
