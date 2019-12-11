package json

import "github.com/imulab/go-scim/pkg/core/prop"

type ResourceMarshalAdapter struct {
	Resource	*prop.Resource
	Include		[]string
	Exclude		[]string
}

func (r ResourceMarshalAdapter) MarshalJSON() ([]byte, error) {
	return Serialize(r.Resource, Options().Include(r.Include...).Exclude(r.Exclude...))
}
