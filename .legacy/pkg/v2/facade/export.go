package facade

import (
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Export exports the object as a prop.Resource. For each field and the corresponding path specified in the "scim" tag,
// it creates a property with the field value at the specified path.
func Export(obj interface{}, resourceType *spec.ResourceType) (*prop.Resource, error) {
	r := prop.NewResource(resourceType)
	if err := crud.Add(r, "schemas", resourceType.Schema().ID()); err != nil {
		return nil, err
	}
	if err := crud.Add(r, "meta.resourceType", resourceType.Name()); err != nil {
		return nil, err
	}

	exp := exporter{}
	forEachMapping(reflect.ValueOf(obj), func(field reflect.Value, path string) error {
		return exp.assign(r, field, path)
	})

	return r, nil
}

type exporter struct{}

func (f exporter) assign(r *prop.Resource, field reflect.Value, path string) error {
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return nil
		}
		return f.assign(r, field.Elem(), path)
	}

	head, err := expr.CompilePath(path)
	if err != nil {
		return err
	}

	nav := r.Navigator()

	for cur := head; cur != nil; cur = cur.Next() {
		switch {
		case cur.IsPath():
			if err := f.stepIn(nav, cur.Token()); err != nil {
				return err
			}
			if cur.Next() == nil {
				if err := f.set(nav, field); err != nil {
					return err
				}
			}
		case cur.IsRootOfFilter():
			if err := f.selectElem(nav, cur); err != nil {
				return err
			}
		default:
			return ErrSCIMPath
		}
	}

	return nil
}

func (f exporter) stepIn(nav prop.Navigator, path string) error {
	nav.Add(map[string]interface{}{path: nil})
	nav.Dot(path)
	return nav.Error()
}

func (f exporter) selectElem(nav prop.Navigator, filter *expr.Expression) error {
	nav.Where(func(child prop.Property) bool {
		ok, _ := crud.EvaluateExpressionOnProperty(child, filter)
		return ok
	})
	if !nav.HasError() {
		return nil
	}

	// Navigator errors because it didn't find such element, clear the
	// error and create it!
	nav.ClearError()

	filterPropValues := map[string]string{}
	if err := f.collectLeafProps(filter, filterPropValues); err != nil {
		return err
	}

	complexData := map[string]interface{}{}
	for k, v := range filterPropValues {
		attr := nav.Current().Attribute().DeriveElementAttribute().SubAttributeForName(k)
		switch attr.Type() {
		case spec.TypeString, spec.TypeReference, spec.TypeDateTime, spec.TypeBinary:
			complexData[k] = v
		case spec.TypeInteger:
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return err
			}
			complexData[k] = i
		case spec.TypeDecimal:
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return err
			}
			complexData[k] = f
		case spec.TypeBoolean:
			b, err := strconv.ParseBool(v)
			if err != nil {
				return err
			}
			complexData[k] = b
		default:
			panic("unexpected type")
		}
	}

	nav.Add(complexData)
	if nav.HasError() {
		return nav.Error()
	}

	nav.Where(func(child prop.Property) bool {
		ok, _ := crud.EvaluateExpressionOnProperty(child, filter)
		return ok
	})
	return nav.Error()
}

func (f exporter) collectLeafProps(root *expr.Expression, collector map[string]string) error {
	if root.IsOperator() {
		if root.Token() != expr.And && root.Token() != expr.Eq {
			return ErrDisallowedOperator
		}
	}

	if root.IsLogicalOperator() {
		if err := f.collectLeafProps(root.Left(), collector); err != nil {
			return err
		}
		return f.collectLeafProps(root.Right(), collector)
	}

	if root.IsRelationalOperator() {
		k := root.Left().Token()
		v := strings.Trim(root.Right().Token(), "\"")
		collector[k] = v
		return nil
	}

	panic("unreachable code")
}

func (f exporter) set(nav prop.Navigator, field reflect.Value) error {
	attr := nav.Current().Attribute()

	if err := typeCheck(attr, field.Type()); err != nil {
		return err
	}

	switch field.Kind() {
	case reflect.String:
		nav.Replace(field.String())
		return nav.Error()
	case reflect.Int64:
		switch attr.Type() {
		case spec.TypeInteger:
			nav.Replace(field.Int())
			return nav.Error()
		case spec.TypeDateTime:
			nav.Replace(time.Unix(field.Int(), 0).UTC().Format(spec.ISO8601))
			return nav.Error()
		}
	case reflect.Float64:
		nav.Replace(field.Float())
		return nav.Error()
	case reflect.Bool:
		nav.Replace(field.Bool())
		return nav.Error()
	case reflect.Slice:
		if attr.MultiValued() {
			var list []interface{}
			for i := 0; i < field.Len(); i++ {
				list = append(list, field.Index(i).Interface())
			}
			nav.Replace(list)
			return nav.Error()
		}
	}

	return ErrInputType
}
