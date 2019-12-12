package crud

import (
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/expr"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/core/spec"
	"sort"
)

type SortOrder string

const (
	SortDefault SortOrder = ""
	SortAsc     SortOrder = "ascending"
	SortDesc    SortOrder = "descending"
)

type (
	Sort struct {
		By    string
		Order SortOrder
	}
	// Include or exclude attributes in the return. At most one can be specified.
	Projection struct {
		Attributes         []string
		ExcludedAttributes []string
	}
	Pagination struct {
		// 1-based start index
		StartIndex int
		Count      int
	}
)

// Sort the given list of resources according to the sort options.
func (s Sort) Sort(resources []*prop.Resource) error {
	if len(resources) <= 1 {
		return nil
	} else if len(s.By) == 0 {
		return nil
	}

	head, err := expr.CompilePath(s.By)
	if err != nil {
		return err
	}

	sort.Sort(&sortWrapper{
		resources: resources,
		by:        head,
		dir:       s.Order,
	})
	return nil
}

type sortWrapper struct {
	by        *expr.Expression
	dir       SortOrder
	resources []*prop.Resource
}

func (s *sortWrapper) Len() int {
	return len(s.resources)
}

func (s *sortWrapper) Less(i, j int) bool {
	var (
		iProp prop.Property
		jProp prop.Property
		err   error
	)

	iProp, err = SeekSortTarget(s.resources[i], s.by)
	if err != nil {
		switch s.dir {
		case SortDefault, SortAsc:
			return false
		case SortDesc:
			return true
		default:
			panic("invalid sortOrder")
		}
	}

	jProp, err = SeekSortTarget(s.resources[j], s.by)
	if err != nil {
		switch s.dir {
		case SortDefault, SortAsc:
			return true
		case SortDesc:
			return false
		default:
			panic("invalid sortOrder")
		}
	}

	less, _ := iProp.LessThan(jProp.Raw())
	switch s.dir {
	case SortDefault, SortAsc:
		return less
	case SortDesc:
		return !less
	default:
		panic("invalid sortOrder")
	}
}

func (s *sortWrapper) Swap(i, j int) {
	s.resources[i], s.resources[j] = s.resources[j], s.resources[i]
}

// Find the actual sort target inferred by the sortBy parameter. 'resource' points to the Resource whose value is being
// sorted. 'by' is the compiled sortBy path. Depending on the sortBy parameter, the actual sort target varies in several cases.
//
// First, if the sortBy parameter refers to a simple type inside a singular complex type, it will be the sort target.
//
//		For instance, sortBy=name.familyName. Given the following resource snippet,
//			{
//				"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
//				"id": "FA33AA7E-99DB-4446-A2CD-050133B442DA",
//				"userName": "imulab",
//				"name": {
//					"familyName": "Qiu",
//					"givenName": "David"
//				}
//			}
//		The value "Qiu" will be used to sort the resource. This also applies to any top level singular simple types, as
// 		this project treats top-level container as a virtual complex property as well.
//
// Second, if the sortBy parameter refers to a simple type inside a multiValued complex type, the sort target will be
// the value whose accompanying primary attribute was set to 'true', or the first value.
//
//		For instance, sortBy=emails.value. Given the following resource snippet,
//			{
//				"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
//				"id": "FA33AA7E-99DB-4446-A2CD-050133B442DA",
//				"userName": "imulab",
//				"emails": [
//					{
//						"value": "foo@bar.com"
//					},
//					{
//						"value": "bar@foo.com",
//						"primary": true
//					}
//				]
//			}
// 		The value "bar@foo.com" will be used to sort the resource, because its accompany primary attribute is true.
//
//		However, if none of the primary attribute was set to true, like:
//			{
//				"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
//				"id": "FA33AA7E-99DB-4446-A2CD-050133B442DA",
//				"userName": "imulab",
//				"emails": [
//					{
//						"value": "foo@bar.com"
//					},
//					{
//						"value": "bar@foo.com"
//					}
//				]
//			}
//		In this case, "foo@bar.com" will be used to sort the resource, because it is the first value.
//
// Third, if the sortBy parameter refers to multiValued simple type, the sort target will be its first element.
//
//		For instance, sortBy=schemas. Given the following resource snippet,
//			{
//				"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User", "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"],
//				"id": "FA33AA7E-99DB-4446-A2CD-050133B442DA",
//				"userName": "imulab",
//				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
// 					"employeeNumber": "11250"
//				}
//			}
//		The value "urn:ietf:params:scim:schemas:core:2.0:User" will be used to sort the resource because it is the first element.
//
// Other sortBy parameters that does not fall into the above three categories are considered to be invalid.
//
func SeekSortTarget(resource *prop.Resource, by *expr.Expression) (prop.Property, error) {
	candidates := make([]prop.Property, 0)
	_ = traverse(resource.NewNavigator(), by, func(target prop.Property) error {
		candidates = append(candidates, target)
		return nil
	})
	if len(candidates) == 0 {
		return nil, errors.NoTarget("sortBy yields no target")
	}
	if candidates[0].Attribute().Type() == spec.TypeComplex {
		return nil, errors.InvalidPath("sortBy must not point to a complex type")
	}
	if len(candidates) == 1 {
		if candidates[0].Attribute().SingleValued() {
			return candidates[0], nil
		} else {
			return prop.NewNavigator(candidates[0]).FocusIndex(0)
		}
	}
	for _, each := range candidates {
		if each.Attribute().IsPrimary() {
			if true == each.Raw() {
				return each, nil
			}
		} else {
			if p, err := prop.NewNavigator(each.Parent()).FocusCriteria(func(child prop.Property) bool {
				return child.Attribute().IsPrimary()
			}); err != nil {
				return nil, err
			} else if true == p.Raw() {
				return each, nil
			}
		}
	}
	return candidates[0], nil
}
