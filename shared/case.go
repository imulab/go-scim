package shared

import (
	"context"
	"reflect"
	"sync"
)

func CorrectCase(subj *Resource, sch *Schema, ctx context.Context) (err error) {
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

	caseCorrectionInstance.correctCaseWithReflection(reflect.ValueOf(subj.Complex), sch.ToAttribute(), ctx)

	err = nil
	return
}

type caseCorrection struct{}

func (cc *caseCorrection) correctCaseWithReflection(v reflect.Value, attr *Attribute, ctx context.Context) {
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
			cc.correctCaseWithReflection(v.Index(i), attr, ctx)
		}

	case reflect.Map:
		cc.resetMapKey(v, attr, ctx)
	}
}

func (cc *caseCorrection) resetMapKey(m reflect.Value, guide *Attribute, ctx context.Context) {
	if m.Kind() != reflect.Map {
		return
	}

	for _, k := range m.MapKeys() {
		p, err := NewPath(k.String())
		if err != nil {
			cc.throw(Error.NoAttribute(p.Value()), ctx)
		}

		attr := guide.GetAttribute(p, false)
		if attr == nil {
			cc.throw(Error.NoAttribute(p.Value()), ctx)
		}

		v := m.MapIndex(k)

		if k.String() != attr.Name {
			m.SetMapIndex(reflect.ValueOf(attr.Name), v)
			m.SetMapIndex(k, reflect.Value{})
		}

		cc.correctCaseWithReflection(v, attr, ctx)
	}
}

func (cc *caseCorrection) throw(err error, ctx context.Context) {
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
