package crud

import (
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// Find the actual sort target inferred by the sortBy parameter. 'resource' points to the Resource whose value is being
// sorted. 'by' is the compiled sortBy path. Depending on the sortBy parameter, the actual sort target varies in several cases.
//
// First, if the sortBy parameter refers to a simple type inside a singular complex type, it will be the sort target. For instance,
//	sortBy=name.familyName
// Given the following resource snippet:
// 	{
// 		"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
// 		"id": "FA33AA7E-99DB-4446-A2CD-050133B442DA",
// 		"userName": "imulab",
// 		"name": {
// 			"familyName": "Qiu",
// 			"givenName": "David"
// 		}
// 	}
// The value "Qiu" will be used to sort the resource. This also applies to any top level singular simple types, as this
// project treats top-level container as a virtual complex property as well.
//
// Second, if the sortBy parameter refers to a simple type inside a multiValued complex type, the sort target will be
// the value whose accompanying primary attribute was set to 'true', or the first value. For instance,
//	sortBy=emails.value
// Given the following resource snippet:
// 	{
// 		"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
// 		"id": "FA33AA7E-99DB-4446-A2CD-050133B442DA",
// 		"userName": "imulab",
// 		"emails": [
// 			{
// 				"value": "foo@bar.com"
// 			},
// 			{
// 				"value": "bar@foo.com",
// 				"primary": true
// 			}
// 		]
// 	}
// The value "bar@foo.com" will be used to sort the resource, because its accompany primary attribute is true.
//
// However, if none of the primary attribute was set to true, like:
// 	{
// 		"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
// 		"id": "FA33AA7E-99DB-4446-A2CD-050133B442DA",
// 		"userName": "imulab",
// 		"emails": [
// 			{
// 				"value": "foo@bar.com"
// 			},
// 			{
// 				"value": "bar@foo.com"
// 			}
// 		]
// 	}
// In this case, "foo@bar.com" will be used to sort the resource, because it is the first value.
//
// Third, if the sortBy parameter refers to multiValued simple type, the sort target will be its first element. For instance,
//	sortBy=schemas
// Given the following resource snippet:
// 	{
// 		"schemas": ["urn:ietf:params:scim:schemas:core:2.0:User", "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"],
// 		"id": "FA33AA7E-99DB-4446-A2CD-050133B442DA",
// 		"userName": "imulab",
// 		"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
//  			"employeeNumber": "11250"
// 		}
// 	}
// The value "urn:ietf:params:scim:schemas:core:2.0:User" will be used to sort the resource because it is the first element.
//
// Other sortBy parameters that does not fall into the above three categories are considered to be invalid.
//
func SeekSortTarget(resource *prop.Resource, by *expr.Expression) (prop.Property, error) {
	if by.ContainsFilter() {
		return nil, fmt.Errorf("%w: sortBy attribute cannot contain filter", spec.ErrInvalidPath)
	}

	var candidates []prop.Property
	if err := primaryOrFirstTraverse(resource.RootProperty(), by, func(nav prop.Navigator) error {
		candidates = append(candidates, nav.Current())
		return nil
	}); err != nil {
		return nil, err
	}

	if len(candidates) != 1 {
		return nil, fmt.Errorf("%w: cannot determine sortBy candidate (%d candidate)", spec.ErrInvalidPath, len(candidates))
	}

	if candidates[0].Attribute().Type() == spec.TypeComplex {
		return nil, fmt.Errorf("%w: invalid sortBy target", spec.ErrInvalidPath)
	}

	if !candidates[0].Attribute().MultiValued() {
		return candidates[0], nil
	}

	if p, err := candidates[0].ChildAtIndex(0); err != nil {
		return nil, fmt.Errorf("%w: cannot determine sortBy candidate (no candidate)", spec.ErrNoTarget)
	} else {
		return p, nil
	}
}
