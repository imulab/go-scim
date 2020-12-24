package facade

import (
	"errors"
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ToResource converts the target structure to a prop.Resource.
//
// To become legible of being converted to a prop.Resource, the target must fulfill the following:
//	1. It must be a struct.
//	2. The fields intended to be transferred must have a "scim" tag.
//	3. The fields tagging the "scim" tag must be of type string, int64, float64, bool, their respective pointer types,
//	and their respective slice types.
//	4. The type of the tagged fields must match the type required by the corresponding attributes specified in the tag.
func ToResource(target interface{}, resourceType *spec.ResourceType) (*prop.Resource, error) {
	return toResource(reflect.ValueOf(target), resourceType)
}

func toResource(target reflect.Value, resourceType *spec.ResourceType) (*prop.Resource, error) {
	if target.Kind() == reflect.Ptr {
		return toResource(target.Elem(), resourceType)
	}

	if target.Kind() != reflect.Struct {
		return nil, errors.New("target is not of type struct")
	}

	resource := prop.NewResource(resourceType)

	for i := 0; i < target.NumField(); i++ {
		tag, ok := target.Type().Field(i).Tag.Lookup("scim")
		if !ok {
			continue
		}

		paths := strings.FieldsFunc(tag, func(r rune) bool { return r == ',' })
		for _, path := range paths {
			if err := assign(target.Field(i), path, resource); err != nil {
				return nil, err
			}
		}
	}

	crud.Add(resource, "meta.resourceType", resourceType.Name())

	return resource, nil
}

func assign(field reflect.Value, path string, resource *prop.Resource) error {
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return nil
		}
		return assign(field.Elem(), path, resource)
	}

	// _ = typeCheck(field)

	nav := resource.Navigator()

	head, err := expr.CompilePath(path)
	if err != nil {
		return err
	}

	for cur := head; cur != nil; cur = cur.Next() {
		if cur.IsPath() {
			attr := nav.Current().Attribute().SubAttributeForName(cur.Token())
			nav.Add(map[string]interface{}{
				cur.Token(): nil,
			})
			nav.Dot(cur.Token())

			if cur.Next() == nil {
				switch field.Kind() {
				case reflect.String:
					switch attr.Type() {
					case spec.TypeString, spec.TypeReference, spec.TypeBinary:
						nav.Replace(field.String())
					default:
						panic("incompatible type")
					}
				case reflect.Int64:
					switch attr.Type() {
					case spec.TypeInteger:
						nav.Replace(field.Int())
					case spec.TypeDateTime:
						nav.Replace(time.Unix(field.Int(), 0).Format(spec.ISO8601))
					default:
						panic("incompatible type")
					}
				case reflect.Float64:
					switch attr.Type() {
					case spec.TypeDecimal:
						nav.Replace(field.Float())
					default:
						panic("incompatible type")
					}
				case reflect.Bool:
					switch attr.Type() {
					case spec.TypeBoolean:
						nav.Replace(field.Bool())
					default:
						panic("incompatible type")
					}
				case reflect.Slice:
					panic("not implemented")
				default:
					panic("incompatible type")
				}
			}
		}

		if cur.IsRootOfFilter() {
			nav.Where(func(child prop.Property) bool {
				ok, _ := crud.EvaluateExpressionOnProperty(child, cur)
				return ok
			})
			if nav.HasError() {
				nav.ClearError()

				kv := map[string]string{}
				if err = collectKv(cur, kv); err != nil {
					return err
				}

				data := map[string]interface{}{}

				for k, v := range kv {
					kattr := nav.Current().Attribute().DeriveElementAttribute().SubAttributeForName(k)
					if kattr == nil {
						continue
					}

					switch kattr.Type() {
					case spec.TypeString, spec.TypeReference, spec.TypeDateTime, spec.TypeBinary:
						data[k] = v
					case spec.TypeInteger:
						i, err := strconv.ParseInt(v, 10, 64)
						if err != nil {
							return err
						}
						data[k] = i
					case spec.TypeDecimal:
						f, err := strconv.ParseFloat(v, 64)
						if err != nil {
							return err
						}
						data[k] = f
					case spec.TypeBoolean:
						b, err := strconv.ParseBool(v)
						if err != nil {
							return err
						}
						data[k] = b
					default:
						panic("unexpected types")
					}
				}

				nav.Add(data)
				nav.Where(func(child prop.Property) bool {
					ok, _ := crud.EvaluateExpressionOnProperty(child, cur)
					return ok
				})
				if nav.HasError() {
					return nav.Error()
				}
			}
		}
	}

	return nil
}

func collectKv(root *expr.Expression, collector map[string]string) error {
	if root.IsOperator() && (root.Token() != expr.And && root.Token() != expr.Eq) {
		return errors.New(`currently, only "and" and "eq" is supported`)
	}

	if root.IsLogicalOperator() {
		if err := collectKv(root.Left(), collector); err != nil {
			return err
		}
		if err := collectKv(root.Right(), collector); err != nil {
			return err
		}
		return nil
	}

	if root.IsRelationalOperator() {
		collector[root.Left().Token()] = strings.Trim(root.Right().Token(), "\"")
		return nil
	}

	panic("unreachable code")
}
