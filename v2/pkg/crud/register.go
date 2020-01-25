package crud

import (
	"github.com/imulab/go-scim/v2/pkg/crud/expr"
	"github.com/imulab/go-scim/v2/pkg/spec"
)

func Register(resourceType *spec.ResourceType) {
	expr.RegisterURN(resourceType.Schema().ID())
	resourceType.ForEachExtension(func(extension *spec.Schema, required bool) {
		expr.RegisterURN(extension.ID())
	})
}
