package facade

import (
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
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

	switch field.Type().Kind() {
	case reflect.String: // string: String, Reference, Binary
		field.SetString(nav.Current().Raw().(string))
		return nil
	case reflect.Int64:
		switch nav.Current().Attribute().Type() {
		case spec.TypeInteger: // int64: Integer
			field.SetInt(nav.Current().Raw().(int64))
			return nil
		case spec.TypeDateTime: // int64: DateTime
			if t, err := time.Parse(spec.ISO8601, nav.Current().Raw().(string)); err != nil {
				return err
			} else {
				field.SetInt(t.UTC().Unix())
				return nil
			}
		}
	case reflect.Float64: // float64: Decimal
		field.SetFloat(nav.Current().Raw().(float64))
		return nil
	case reflect.Bool: // bool: Boolean
		field.SetBool(nav.Current().Raw().(bool))
		return nil
	case reflect.Ptr:
		switch field.Type().Elem().Kind() {
		case reflect.String: // *string: String, Reference, Binary
			v := nav.Current().Raw().(string)
			field.Set(reflect.ValueOf(&v))
			return nil
		case reflect.Int64:
			switch nav.Current().Attribute().Type() {
			case spec.TypeInteger: // *int64: Integer
				v := nav.Current().Raw().(int64)
				field.Set(reflect.ValueOf(&v))
				return nil
			case spec.TypeDateTime: // *int64: DateTime
				if t, err := time.Parse(spec.ISO8601, nav.Current().Raw().(string)); err != nil {
					return err
				} else {
					v := t.UTC().Unix()
					field.Set(reflect.ValueOf(&v))
					return nil
				}
			}
		case reflect.Float64: // *float64: Decimal
			v := nav.Current().Raw().(float64)
			field.Set(reflect.ValueOf(&v))
			return nil
		case reflect.Bool: // *bool: Boolean
			v := nav.Current().Raw().(bool)
			field.Set(reflect.ValueOf(&v))
			return nil
		}
	case reflect.Slice:
		switch field.Type().Elem().Kind() {
		case reflect.String: // []string: String, Reference, Binary
			var array []string
			for _, elem := range nav.Current().Raw().([]interface{}) {
				array = append(array, elem.(string))
			}
			field.Set(reflect.ValueOf(array))
			return nil
		case reflect.Int64:
			switch nav.Current().Attribute().Type() {
			case spec.TypeInteger: // []int64: Integer
				var array []int64
				for _, elem := range nav.Current().Raw().([]interface{}) {
					array = append(array, elem.(int64))
				}
				field.Set(reflect.ValueOf(array))
				return nil
			case spec.TypeDateTime: // []int64: DateTime
				var array []int64
				for _, elem := range nav.Current().Raw().([]interface{}) {
					if t, err := time.Parse(spec.ISO8601, elem.(string)); err != nil {
						return err
					} else {
						array = append(array, t.UTC().Unix())
					}
				}
				field.Set(reflect.ValueOf(array))
				return nil
			}
		case reflect.Float64: // []float64: Decimal
			var array []float64
			for _, elem := range nav.Current().Raw().([]interface{}) {
				array = append(array, elem.(float64))
			}
			field.Set(reflect.ValueOf(&array))
			return nil
		case reflect.Bool: // []bool: Boolean
			var array []bool
			for _, elem := range nav.Current().Raw().([]interface{}) {
				array = append(array, elem.(bool))
			}
			field.Set(reflect.ValueOf(&array))
			return nil
		}
	}

	return ErrInputType
}
