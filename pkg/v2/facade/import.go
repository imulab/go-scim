package facade

import (
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/facade/internal"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"reflect"
	"time"
)

// Import imports the values of the resource into the destination object. For each field and its corresponding path
// specified in the "scim" tag, it assigns the value at the specified path from the resource.
func Import(res *prop.Resource, dest interface{}) error {
	imp := importer{}
	return forEachMapping(reflect.ValueOf(dest), func(field reflect.Value, path string) error {
		return imp.assign(res, path, field)
	})
}

type importer struct{}

func (f importer) assign(resource *prop.Resource, path string, field reflect.Value) error {
	head, err := expr.CompilePath(path)
	if err != nil {
		return err
	}

	nav := resource.Navigator()

	for cur := head; cur != nil; cur = cur.Next() {
		switch {
		case cur.IsPath():
			nav.Dot(cur.Token())
			if nav.HasError() {
				return nav.Error()
			}
		case cur.IsRootOfFilter():
			nav.Where(func(child prop.Property) bool {
				ok, _ := crud.EvaluateExpressionOnProperty(child, cur)
				return ok
			})
			if nav.HasError() {
				return nav.Error()
			}
		default:
			return ErrSCIMPath
		}
	}

	if nav.Current().IsUnassigned() {
		return nil
	}

	err = typeCheck(nav.Current().Attribute(), field.Type())
	if err != nil {
		return err
	}

	attr := nav.Current().Attribute()
	if attr.MultiValued() {
		slice := internal.Slice(nav.Current().Raw().([]interface{}))
		switch attr.Type() {
		case spec.TypeString, spec.TypeReference, spec.TypeBinary:
			field.Set(reflect.ValueOf(slice.StringTyped()))
		case spec.TypeInteger:
			field.Set(reflect.ValueOf(slice.Int64Typed()))
		case spec.TypeDecimal:
			field.Set(reflect.ValueOf(slice.Float64Typed()))
		case spec.TypeBoolean:
			field.Set(reflect.ValueOf(slice.BoolTyped()))
		case spec.TypeDateTime:
			var timestamps []int64
			for _, each := range slice {
				var t time.Time
				t, err = time.Parse(spec.ISO8601, each.(string))
				if err != nil {
					return err
				}
				timestamps = append(timestamps, t.UTC().Unix())
			}
			field.Set(reflect.ValueOf(timestamps))
		}
	} else {
		switch attr.Type() {
		case spec.TypeString, spec.TypeReference, spec.TypeBinary:
			err = internal.SetString(field, nav.Current().Raw().(string))
		case spec.TypeInteger:
			err = internal.SetInt64(field, nav.Current().Raw().(int64))
		case spec.TypeDecimal:
			err = internal.SetFloat64(field, nav.Current().Raw().(float64))
		case spec.TypeBoolean:
			err = internal.SetBool(field, nav.Current().Raw().(bool))
		case spec.TypeDateTime:
			var t time.Time
			t, err = time.Parse(spec.ISO8601, nav.Current().Raw().(string))
			if err != nil {
				break
			}
			err = internal.SetInt64(field, t.UTC().Unix())
		}
	}

	if err != nil {
		return ErrInputType
	}

	return nil
}
