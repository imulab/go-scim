// This package serves as a frontend of custom structures that are mappable to SCIM schemas.
//
// Export and Import are the two main entrypoints. For structures to be recognized by these entrypoints, the intended
// fields must be tagged with "scim", whose content is a comma delimited list of SCIM paths. Apart from having to be a
// legal path backed by the resource type, a filtered path may be allowed, provided that only the "and" and "eq" predicate
// is used inside the filter. A filtered path is essential in mapping one or more fields into a multi-valued complex
// property. The following is an example of legal paths under the User resource type with User schema and the Enterprise
// User schema extension:
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
//		Name		string	`scim:"name.formatted"`
//		NickName	*string	`scim:"nickName"`
//		CreatedAt	int64	`scim:"meta.created"`
//		UpdatedAt	int64	`scim:"meta.lastModified"`
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
package facade
