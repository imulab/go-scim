package crud

import (
	"github.com/elvsn/scim.go/crud/expr"
	"github.com/elvsn/scim.go/spec"
)

func Register(resourceType *spec.ResourceType) {
	expr.RegisterURN(resourceType.Schema().ID())
	resourceType.ForEachExtension(func(extension *spec.Schema, required bool) {
		expr.RegisterURN(extension.ID())
	})
}
