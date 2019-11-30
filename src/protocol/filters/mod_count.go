package filters

import (
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
)


func NewInitialModCountFieldFilter() protocol.FieldFilter {
	return &initialModCountFieldFilter{}
}

type (
	initialModCountFieldFilter    struct {}
	InitialModCountFieldFilterKey struct {}
)

func (f *initialModCountFieldFilter) Supports(attribute *core.Attribute) bool {
	return true
}

func (f *initialModCountFieldFilter) Order() int {
	return 0
}

func (f *initialModCountFieldFilter) Filter(ctx *protocol.FieldFilterContext, resource *prop.Resource, property core.Property) error {
	f.putModCount(ctx, property)
	return nil
}

func (f *initialModCountFieldFilter) FieldRef(ctx *protocol.FieldFilterContext, resource *prop.Resource, property core.Property, refResource *prop.Resource, refProperty core.Property) error {
	f.putModCount(ctx, property)
	return nil
}

func (f *initialModCountFieldFilter) putModCount(ctx *protocol.FieldFilterContext, property core.Property) {
	var stats map[string]int
	{
		m, ok := ctx.Get(InitialModCountFieldFilterKey{})
		if !ok {
			stats = make(map[string]int)
		} else {
			stats = m.(map[string]int)
		}
	}

	stats[property.Attribute().ID()] = property.ModCount()
	ctx.Put(InitialModCountFieldFilterKey{}, stats)
}