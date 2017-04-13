package shared

import (
	"reflect"
	"sync"
)

func CorrectCase(subj *Resource, sch *Schema) (err error) {
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

	caseCorrectionInstance.correctCaseWithReflection(reflect.ValueOf(subj.Complex), sch.ToAttribute())

	err = nil
	return
}

type caseCorrection struct{}

func (cc *caseCorrection) correctCaseWithReflection(v reflect.Value, attr *Attribute) {
	if !v.IsValid() {
		return
	}

	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			cc.correctCaseWithReflection(v.Index(i), attr)
		}

	case reflect.Map:
		cc.resetMapKey(v, attr)
	}
}

func (cc *caseCorrection) resetMapKey(m reflect.Value, guide *Attribute) {
	if m.Kind() != reflect.Map {
		return
	}

	for _, k := range m.MapKeys() {
		p, err := NewPath(k.String())
		if err != nil {
			cc.throw(Error.NoAttribute(p.Value()))
		}

		attr := guide.GetAttribute(p, false)
		if attr == nil {
			cc.throw(Error.NoAttribute(p.Value()))
		}

		v := m.MapIndex(k)

		if k.String() != attr.Name {
			m.SetMapIndex(reflect.ValueOf(attr.Name), v)
			m.SetMapIndex(k, reflect.Value{})
		}

		cc.correctCaseWithReflection(v, attr)
	}
}

func (cc *caseCorrection) throw(err error) {
	panic(err)
}

var (
	singleCaseCorrection   sync.Once
	caseCorrectionInstance *caseCorrection
)

func init() {
	singleCaseCorrection.Do(func() {
		caseCorrectionInstance = &caseCorrection{}
	})
}
