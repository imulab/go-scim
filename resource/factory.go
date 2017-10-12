package resource

import "github.com/pkg/errors"

var (
	ErrorSchemaNotSupported = errors.New("schema is not supported.")
)

// A factory that initiates the desired types.
type Factory interface {

	// Instantiate SCIM resource by inspecting the schema types.
	// For instance, if the schema types contains 'urn:ietf:params:scim:schemas:core:2.0:User', the factory
	// shall instantiate a standard or extended user resource.
	NewResource(schemas []string) (interface{}, error)

	//NewArray(attr *Attribute) (interface{}, error)
}

// Default implementation of Factory. It can handle schema types of
// - urn:ietf:params:scim:schemas:core:2.0:User
// - urn:ietf:params:scim:schemas:core:2.0:Group
type DefaultFactory struct {}

func (f DefaultFactory) NewResource(schemas []string) (interface{}, error) {
	hash := f.buildHash(schemas)
	switch {
	case f.containsAll(hash, UserUrn):
		return &User{
			DisplayName: "",
		}, nil
	default:
		return nil, ErrorSchemaNotSupported
	}
}

func (f DefaultFactory) buildHash(schemas []string) map[string]struct{} {
	hash := make(map[string]struct{})
	for _, each := range schemas {
		hash[each] = struct{}{}
	}
	return hash
}

func (f DefaultFactory) containsAll(hash map[string]struct{}, targets ...string) bool {
	for _, each := range targets {
		if _, ok := hash[each]; !ok {
			return false
		}
	}
	return true
}