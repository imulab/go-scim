package crud

import (
	"github.com/elvsn/scim.go/crud/internal"
	"github.com/elvsn/scim.go/spec"
)

func Register(resourceType *spec.ResourceType) {
	internal.RegisterURN(resourceType.Schema().ID())
	resourceType.ForEachExtension(func(extension *spec.Schema, required bool) {
		internal.RegisterURN(extension.ID())
	})
}
