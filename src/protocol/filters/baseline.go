package filters

import (
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
)

// Create a resource filter that records the current hash of the reference resource as the baseline under the key
// BaselineHashKey.
func NewBaselineHashResourceFilter(order int) protocol.ResourceFilter {
	return &baselineHashFilter{order: order}
}

type (
	BaselineHashKey    struct{}
	baselineHashFilter struct {
		order int
	}
)

func (f *baselineHashFilter) Order() int {
	return f.order
}

func (f *baselineHashFilter) Filter(ctx *protocol.FilterContext, resource *prop.Resource) error {
	return nil
}

func (f *baselineHashFilter) FilterRef(ctx *protocol.FilterContext, resource *prop.Resource, ref *prop.Resource) error {
	ctx.Put(BaselineHashKey{}, ref.Hash())
	return nil
}
