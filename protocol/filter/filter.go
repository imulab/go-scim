package filter

import (
	"context"
	"github.com/imulab/go-scim/core"
)

// A property filter is the main processing stage that the resource go through after being parsed and before being
// sent to a persistence provider. The implementations can carry out works like annotation processing, validation,
// value generation, etc.
type (
	PropertyFilter interface {
		// Returns true if this filter supports processing the given attribute.
		Supports(attribute *core.Attribute) bool
		// Returns an integer based order value, so that different filters working on the same attribute can be sorted
		// and called in sequence. The filter can choose to return the same order value or different order value for
		// different attributes.
		//
		// As a general rule of thumb, in the many stages the filters may be conceptually divide into, stage 1 filters
		// start with index 100; stage 2 filters start with index 200; and so on...
		Order(attribute *core.Attribute) int
		// Process the given property, with access to the owning resource and the reference resource. The reference
		// resource is optional, and when present, will present a view of the data in the persistence state.
		Filter(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource) error
	}
)

// Build an index map of attribute id corresponding a sorted list of property filters, based on their PropertyFilter.Order
// reaction to the attribute. All unique derived attributes will be tried with filters, only only those that is supported
// by at least one of the filters will be present in the final result.
//
// This method uses a slow insertion sort to perform the ordering. Since this method is a setup phase method, and the
// number of filters corresponding to each attribute id is not expected to be high, this slow sorting method poses no
// immediate threat to performance. To enhance performance, provide an already sorted filters array to this method.
func BuildIndex(resourceTypes []*core.ResourceType, filters []PropertyFilter) map[string][]PropertyFilter {
	var attributes map[*core.Attribute]struct{}
	{
		// build a unique set of attributes, to make sure PropertyFilter.Supports is not called twice.
		attributes = make(map[*core.Attribute]struct{})
		for _, resourceType := range resourceTypes {
			for _, attribute := range resourceType.DerivedAttributes() {
				attributes[attribute] = struct{}{}
			}
		}
	}

	var index map[*core.Attribute][]PropertyFilter
	{
		index = make(map[*core.Attribute][]PropertyFilter)
		for attribute := range attributes {
			for _, filter := range filters {
				if filter.Supports(attribute) {
					supported, ok := index[attribute]
					if !ok {
						supported = make([]PropertyFilter, 0)
					}
					supported = append(supported, filter)
					index[attribute] = supported
				}
			}
		}
	}

	var result map[string][]PropertyFilter
	{
		result = make(map[string][]PropertyFilter)
		for attribute, filters := range index {
			if len(filters) > 1 {
				// Here we usually have a small number (< 5) of filters corresponding to each attribute, and this
				// method is only expected to be called during the initialization phase. Hence, we use the O(N^2)
				// but simple insertion sort here.
				orders := make([]int, len(filters), len(filters))
				for i, filter := range filters {
					orders[i] = filter.Order(attribute)
				}
				for i := 1; i < len(orders); i++ {
					for j := i; j > 0; j-- {
						if orders[j-1] > orders[j] {
							orders[j-1], orders[j] = orders[j], orders[j-1]
							filters[j-1], filters[j] = filters[j], filters[j-1]
						}
					}
				}
			}
			result[attribute.Id] = filters
		}
	}

	return result
}

// Return true if the attribute's metadata contains the queried annotation. The annotation is case sensitive.
func ContainsAnnotation(attr *core.Attribute, annotation string) bool {
	metadata := core.Meta.Get(attr.Id, core.DefaultMetadataId)
	if metadata == nil {
		return false
	}

	annotations := metadata.(*core.DefaultMetadata).Annotations
	for _, each := range annotations {
		if each == annotation {
			return true
		}
	}

	return false
}