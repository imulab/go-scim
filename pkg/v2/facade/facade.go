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

var (
	ErrNilInput           = errors.New("the input object is nil")
	ErrInputType          = errors.New("the input object has a wrong type")
	ErrDisallowedOperator = errors.New("a filter contains disallowed operators")
	ErrSCIMPath           = errors.New("the input object contains an invalid SCIM path")
)

// New is the constructor for Facade.
func New(resourceType *spec.ResourceType) *Facade {
	return &Facade{resourceType: resourceType}
}

// Facade is the frontend of any compatible structures that fully or partially conform to a spec.ResourceType. By using
// the Facade, traditional "flat" database objects can be adapted to SCIM API.
type Facade struct {
	resourceType *spec.ResourceType
}

// Export converts the given object to a prop.Resource. The supplied object must be a struct, or a non-nil pointer to
// a struct. Any other input object will be rejected. Depending on the availability of data in the input object, the
// returned prop.Resource may or may not conform to the constraints imposed by the spec.ResourceType.
//
// For the fields to be recognized by this method, they must be tagged with "scim" whose content is a comma delimited
// list of SCIM paths. Apart from having to be a legal path backed by the resource type, a filtered path may be allowed,
// provided that only the "and" and "eq" predicate is used inside the filter. A filtered path is essential in mapping
// one or more fields into a multi-valued complex property. The following is an example of legal paths under the
// User resource type with User schema and the Enterprise User schema extension:
//
//	1. id
//	2. meta.created
//	3. name.formatted
//	4. emails[type eq "work"].value
//	5. addresses[type eq "office" and primary eq true].value
//	6. urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager.value
//
// In addition to the "scim" tag definition, the types of tagging fields must also conform to the following rules:
//
//	1. SCIM String: string or *string
//	2. SCIM Integer: int64 or *int64
//	3. SCIM Decimal: float64 or *float64
//	4. SCIM Boolean: bool or *bool
//	5. SCIM DateTime: int64 or *int64, which contains a UNIX timestamp.
//	6. SCIM Reference: string or *string
//	7. SCIM Binary: string or *string, which contains the Base64 encoded data
//
// For multi-valued properties, the struct field can use the slice of the above non-pointer types. For instance, for a
// multi-valued string property, the corresponding type is []string. Nil slices and nil pointers are interpreted as
// "unassigned" and skipped. Because Facade is intended for traditional flat domain objects like SQL table domains, there
// is no type mapping for complex objects. Complex objects will be constructed by mapping a field to a nested SCIM path,
// hence creating the intended hierarchy.
//
// In addition to the user defined fields, some internal properties will be automatically assigned. The "schemas" property
// always reflects the schemas used in the "scim" tags. The "meta.resourceType" is always assigned to the name of the
// spec.ResourceType defined in the Facade.
//
// The following is a complete example of an object that can be converted to prop.Resource.
//
//	type User struct {
//		Id			string	`scim:"id"`
//		Email 		string	`scim:"userName,emails[type eq \"work\" and primary eq true].value"`
//		BackupEmail *string	`scim:"emails[type eq \"work\" and primary eq false].value"`
//		Name		string	`scim:"name.formatted"
//		NickName	*string	`scim:"nickName"`
//		CreatedAt	int64	`scim:"meta.created"`
//		UpdatedAt	int64	`scim:"meta.lastModified"
//		Active		bool	`scim:"active"`
//		Manager		*string	`scim:"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager.value"`
//	}
//
//	// ref is a pseudo function that returns reference to a string
//	var user = &User{
//		Id: "test",
//		Email: "john@gmail.com",
//		BackupEmail: ref("john@outlook.com"),
//		Name: "John Doe",
//		NickName: nil,
//		CreatedAt: 1608795238,
//		UpdatedAt: 1608795238,
//		Active: false,
//		Manager: ref("tom"),
//	}
//
//	// The above object can be converted to prop.Resource, which will in turn produce the following JSON when rendered:
//	{
//		"schemas": [
//			"urn:ietf:params:scim:schemas:core:2.0:User",
//			"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
//		],
//		"id": "test",
//		"meta": {
//			"resourceType": "User",
//			"created": "2020-12-24T07:33:58",
//			"lastModified": "2020-12-24T07:33:58"
//		},
//		"name": {
//			"formatted": "John Doe"
//		},
//		"emails": [{
//			"value": "john@gmail.com",
//			"type": "work",
//			"primary": true
//		}, {
//			"value": "john@outlook.com",
//			"type": "work",
//			"primary": false
//		}],
//		"active": false,
//		"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
//			"manager": {
//				"value": "tom"
//			}
//		}
//	}
//
// Some tips for designing the domain object structure. First, use concrete types when the data is known to be not nil,
// and use pointer types when data is nullable. Second, when adding two fields to distinct complex objects inside a
// multi-valued property, do not use overlapping filters. For example, [type eq "work" and primary eq true] overlaps
// with [type eq "work"], but it does not overlap with [type eq "work" and primary eq false]. If overlapping cannot be
// avoided, place the fields with the more general filter in front.
func (f *Facade) Export(obj interface{}) (*prop.Resource, error) {
	return f.export(reflect.ValueOf(obj))
}

func (f *Facade) export(v reflect.Value) (*prop.Resource, error) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, ErrNilInput
		}
		return f.export(v.Elem())
	}

	if v.Kind() != reflect.Struct {
		return nil, ErrInputType
	}

	r := prop.NewResource(f.resourceType)
	if err := crud.Add(r, "schemas", f.resourceType.Schema().ID()); err != nil {
		return nil, err
	}
	if err := crud.Add(r, "meta.resourceType", f.resourceType.Name()); err != nil {
		return nil, err
	}

	for i := 0; i < v.NumField(); i++ {
		scimTag, ok := v.Type().Field(i).Tag.Lookup("scim")
		if !ok {
			continue
		}

		paths := strings.FieldsFunc(scimTag, func(r rune) bool { return r == ',' })
		for _, path := range paths {
			if err := f.assignFieldValueAtPath(r, v.Field(i), path); err != nil {
				return nil, err
			}
		}
	}

	return r, nil
}

func (f *Facade) assignFieldValueAtPath(r *prop.Resource, field reflect.Value, path string) error {
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return nil
		}
		return f.assignFieldValueAtPath(r, field.Elem(), path)
	}

	head, err := expr.CompilePath(path)
	if err != nil {
		return err
	}

	nav := r.Navigator()

	for cur := head; cur != nil; cur = cur.Next() {
		switch {
		case cur.IsPath():
			if err := stepIntoPath(nav, cur.Token()); err != nil {
				return err
			}
			if cur.Next() == nil {
				if err := setField(nav, field); err != nil {
					return err
				}
			}
		case cur.IsRootOfFilter():
			if err := selectElement(nav, cur); err != nil {
				return err
			}
		default:
			return ErrSCIMPath
		}
	}

	return nil
}

func stepIntoPath(nav prop.Navigator, path string) error {
	nav.Add(map[string]interface{}{path: nil})
	nav.Dot(path)
	return nav.Error()
}

func setField(nav prop.Navigator, field reflect.Value) error {
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
			nav.Replace(time.Unix(field.Int(), 0).Format(spec.ISO8601))
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

func selectElement(nav prop.Navigator, filter *expr.Expression) error {
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
	if err := collectPropsImpliedByFilter(filter, filterPropValues); err != nil {
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

func collectPropsImpliedByFilter(root *expr.Expression, collector map[string]string) error {
	if root.IsOperator() {
		if root.Token() != expr.And && root.Token() != expr.Eq {
			return ErrDisallowedOperator
		}
	}

	if root.IsLogicalOperator() {
		if err := collectPropsImpliedByFilter(root.Left(), collector); err != nil {
			return err
		}
		return collectPropsImpliedByFilter(root.Right(), collector)
	}

	if root.IsRelationalOperator() {
		k := root.Left().Token()
		v := strings.Trim(root.Right().Token(), "\"")
		collector[k] = v
		return nil
	}

	panic("unreachable code")
}

func typeCheck(attr *spec.Attribute, t reflect.Type) error {
	switch t.Kind() {
	case reflect.String:
		switch attr.Type() {
		case spec.TypeString, spec.TypeReference, spec.TypeBinary:
			return nil
		}
	case reflect.Int64:
		switch attr.Type() {
		case spec.TypeInteger, spec.TypeDateTime:
			return nil
		}
	case reflect.Float64:
		if spec.TypeDecimal == attr.Type() {
			return nil
		}
	case reflect.Bool:
		if spec.TypeBoolean == attr.Type() {
			return nil
		}
	case reflect.Slice:
		if attr.MultiValued() {
			return typeCheck(attr.DeriveElementAttribute(), t.Elem())
		}
	}

	return ErrInputType
}
