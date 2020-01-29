package crud

import (
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

func Register(resourceType *spec.ResourceType) {
	expr.RegisterURN(resourceType.Schema().ID())
	resourceType.ForEachExtension(func(extension *spec.Schema, required bool) {
		expr.RegisterURN(extension.ID())
	})
}
