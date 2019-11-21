package protocol

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/imulab/go-scim/core"
	"github.com/satori/go.uuid"
	"math/rand"
	"strings"
	"time"
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

// Create a new ID filter. The filter is responsible of processing field id. It is intended to generate a new UUID
// for the resource to serve as its id, hence the annotation is usually marked on the id field. The filter replaces
// the id field value with a new UUID when Filter is called; it does nothing when FilterWithRef is called.
func NewIDFilter() PropertyFilter {
	return &idFilter{}
}

// Create a meta resource type filter. The filter is responsible of assigning resource's resource type to the field
// 'meta.resourceType'. The filter only assigns the resource type when Filter is called. The filter is a no-op when
// FilterWithRef is called.
func NewMetaResourceTypeFilter() PropertyFilter {
	return &metaResourceTypeFilter{}
}

// Create a meta created filter. The filter is responsible of assigning the current time to the field 'meta.created'
// when Filter is called. The filter is a no-op when FilterWithRef is called.
func NewMetaCreatedFilter() PropertyFilter {
	return &metaCreatedFilter{}
}

// Create a meta lastModified filter. The filter is responsible of assigning the current time to the field 'meta.lastModified'
// when either Filter or FilterWithRef is called.
func NewMetaLastModifiedFilter() PropertyFilter {
	return &metaLastModifiedFilter{}
}

// Create a meta location filter. The filter is responsible of generating the resource location url and assign it to field
// 'meta.location'. Id must have been generated and bulkId is not accepted. Generation only happens when Filter is called;
// when FilterWithRef is called, this is a no-op.
func NewMetaLocationFilter(locationFormats map[string]string) PropertyFilter {
	return &metaLocationFilter{
		locationFormats: locationFormats,
	}
}

// Create a meta version filter. The filter is responsible of assigning a new version based on an sha256 hash of the
// resource's id, current time and a random number in range of [0, 10000). Naturally, id must have been generated.
// The version assignment happens on both Filter and FilterWithRef call.
func NewMetaVersionFilter() PropertyFilter {
	return &metaVersionFilter{}
}

// Create a mutability filter. The filter is responsible for scanning properties with readOnly or immutable attributes.
// However, readOnly attributes can be skipped by marking it with annotation '@skipReadOnly'. In general, property values
// with readOnly attribute is deleted in absence of a reference property, while they are overwritten when in presence of
// a reference property; on the other hand, property values with an immutable attribute is ignored in absence of a reference
// property, while in the presence of a reference property, they are matched with the reference property value to ensure
// values have not changed.
func NewMutabilityFilter() PropertyFilter {
	return &mutabilityFilter{}
}

// Create an required filter. The filter is responsible for checking any attribute whose required is set true that they
// are not unassigned.
func NewRequiredFilter() PropertyFilter {
	return &requiredFilter{}
}

// Create an uniqueness filter. This filter is responsible for checking properties whose attribute's uniqueness constraint
// has value 'server'. It will make sure, for the given value, no other resource in the database has that value.
func NewUniquenessFilter(providers []PersistenceProvider) PropertyFilter {
	return &uniquenessFilter{
		providers: providers,
	}
}

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
		// As a general rule of thumb, in the many stages the filters may be conceptually divide into, stage 1 filters
		// start with index 100; stage 2 filters start with index 200; and so on...
		Order(attribute *core.Attribute) int
		// Process the given property, with access to the owning resource.
		Filter(ctx context.Context, resource *core.Resource, property core.Property) error
		// Process the given property, with access to the owning and reference resource, and property.
		FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error
	}

	idFilter struct{}
	metaResourceTypeFilter struct{}
	metaCreatedFilter struct{}
	metaLastModifiedFilter struct{}
	metaLocationFilter struct {
		locationFormats map[string]string
	}
	metaVersionFilter struct{}
	mutabilityFilter struct{}
	requiredFilter struct{}
	uniquenessFilter struct {
		providers []PersistenceProvider
	}
)

// --- idFilter ---

func (f *idFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Id == "id"
}

func (f *idFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *idFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	return property.(core.Crud).Replace(nil, strings.ToLower(uuid.NewV4().String()))
}

func (f *idFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	return nil
}

// --- metaResourceTypeFilter ---

func (f *metaResourceTypeFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Id == "meta.resourceType"
}

func (f *metaResourceTypeFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaResourceTypeFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	return property.(core.Crud).Replace(nil, resource.GetResourceType().Name)
}

func (f *metaResourceTypeFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	return nil
}

// --- metaCreatedFilter ---

func (f *metaCreatedFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Id == "meta.created"
}

func (f *metaCreatedFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaCreatedFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	return property.(core.Crud).Replace(nil, time.Now().Format(core.ISO8601))
}

func (f *metaCreatedFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	return nil
}

// --- metaLastModifiedFilter ---

func (metaLastModifiedFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Id == "meta.lastModified"
}

func (metaLastModifiedFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaLastModifiedFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	return property.(core.Crud).Replace(nil, time.Now().Format(core.ISO8601))
}

func (f *metaLastModifiedFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	return property.(core.Crud).Replace(nil, time.Now().Format(core.ISO8601))
}

// --- metaLocationFilter ---

func (f *metaLocationFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Id == "meta.location"
}

func (f *metaLocationFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaLocationFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	raw, err := resource.Get(core.Steps.NewPath("id"))
	if err != nil || raw == nil {
		return core.Errors.Internal("failed to obtain resource id")
	}

	id, ok := raw.(string)
	if !ok || len(id) == 0 {
		return core.Errors.Internal("invalid id")
	} else if strings.HasPrefix(id, "bulkId:") {
		return core.Errors.Internal("location filter failed: cannot process bulkId")
	}

	format := f.locationFormats[resource.GetResourceType().Id]
	if len(format) == 0 {
		panic("location url formats for all resource types must be set in metaFilter")
	}

	return property.(core.Crud).Replace(nil, fmt.Sprintf(format, id))
}

func (f *metaLocationFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	return nil
}

// --- metaVersionFilter ---

func (f *metaVersionFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Id == "meta.version"
}

func (f *metaVersionFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaVersionFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	return f.assignNewVersion(resource, property)
}

func (f *metaVersionFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	return f.assignNewVersion(resource, property)
}

func (f *metaVersionFilter) assignNewVersion(resource *core.Resource, property core.Property) error {
	id, err := resource.Get(core.Steps.NewPath("id"))
	if err != nil || id == nil || len(id.(string)) == 0 {
		return core.Errors.Internal("could not locate id")
	}

	sha := sha256.New()
	sum := sha.Sum([]byte(fmt.Sprintf("%s:%d:%d", id, time.Now().Unix(), rand.Intn(10000))))

	return property.(core.Crud).Replace(nil, fmt.Sprintf("W/\"%x\"", sum))
}

// --- mutabilityFilter ---

func (f *mutabilityFilter) Supports(attribute *core.Attribute) bool {
	if ContainsAnnotation(attribute, "@mutability:skip") {
		// Because this filter copies reference property values to resource property when mutability is readOnly,
		// the property handled by this filter may have already been handled before by the same filter running against
		// a container attribute. For example, the 'meta' attribute is normally readOnly, one does not wish to double
		// process all its sub attributes when the 'meta' attribute has already been copied. Hence, we suggest to
		// mark any readOnly sub properties whose container property is also readOnly with '@mutability:skip'. This way,
		// we avoid the double processing.
		return false
	}
	return attribute.Mutability == core.MutabilityReadOnly || attribute.Mutability == core.MutabilityImmutable
}

func (f *mutabilityFilter) Order(attribute *core.Attribute) int {
	return 100
}

func (f *mutabilityFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	if property.Attribute().Mutability == core.MutabilityReadOnly {
		return property.(core.Crud).Delete(nil)
	}
	return nil
}

func (f *mutabilityFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	switch property.Attribute().Mutability {
	case core.MutabilityReadOnly:
		if ref == nil || refProp == nil {
			return property.(core.Crud).Delete(nil)
		}
		return property.(core.Crud).Replace(nil, refProp.Raw())
	case core.MutabilityImmutable:
		if ref == nil || refProp == nil || refProp.IsUnassigned() {
			// no reference or reference does not contain a value, then
			// property is free to have any value, given immutable.
			return nil
		}
		if !property.(core.EqualAware).Matches(refProp) {
			return core.Errors.Mutability("'%s' is immutable, but value has changed", property.Attribute().DisplayName())
		}
		return nil
	default:
		return nil
	}
}

// --- requiredFilter ---

func (f *requiredFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Required
}

func (f *requiredFilter) Order(attribute *core.Attribute) int {
	return 500
}

func (f *requiredFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	return f.required(property)
}

func (f *requiredFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	return f.required(property)
}

func (f *requiredFilter) required(property core.Property) error {
	if !property.IsUnassigned() {
		return nil
	}
	return core.Errors.InvalidValue("'%s' is required, but is unassigned", property.Attribute().DisplayName())
}

// --- uniquenessFilter ---

func (f *uniquenessFilter) Supports(attribute *core.Attribute) bool {
	return attribute.Uniqueness == core.UniquenessServer
}

func (f *uniquenessFilter) Order(attribute *core.Attribute) int {
	return 600
}

func (f *uniquenessFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	return f.unique(ctx, resource, property)
}

func (f *uniquenessFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource, refProp core.Property) error {
	return f.unique(ctx, resource, property)
}

func (f *uniquenessFilter) unique(ctx context.Context, resource *core.Resource, property core.Property) error {
	if property.Attribute().MultiValued || property.Attribute().Type == core.TypeComplex {
		// uniqueness check does not apply to multiValued or complex attribute
		return nil
	}

	if property.IsUnassigned() {
		// no need to check uniqueness for unassigned values
		return nil
	}

	var provider PersistenceProvider
	{
		for _, each := range f.providers {
			if each.IsResourceTypeSupported(resource.GetResourceType()) && each.IsFilterSupported() {
				provider = each
				break
			}
		}
		// no provider configured to check uniqueness
		// silently exit
		if provider == nil {
			return nil
		}
	}

	var (
		id         string
		path       string
		scimFilter string
	)
	{
		path = core.Meta.Get(property.Attribute().Id, core.DefaultMetadataId).(*core.DefaultMetadata).Path
		if path == "id" {
			// special case: because we will use id in other queries
			scimFilter = fmt.Sprintf("id eq \"%s\"", property.Raw())
		} else {
			if v, err := resource.Get(core.Steps.NewPath("id")); err != nil {
				return core.Errors.Internal("failed to obtain id")
			} else if v == nil || len(v.(string)) == 0 {
				return core.Errors.Internal("invalid resource id")
			} else {
				id = v.(string)
			}

			switch property.Attribute().Type {
			case core.TypeString, core.TypeReference, core.TypeBinary, core.TypeDateTime:
				scimFilter = fmt.Sprintf("(id ne \"%s\") and (%s eq \"%s\")", id, path, property.Raw())
			case core.TypeInteger:
				scimFilter = fmt.Sprintf("(id ne \"%s\") and (%s eq %d)", id, path, property.Raw())
			case core.TypeDecimal:
				scimFilter = fmt.Sprintf("(id ne \"%s\") and (%s eq %f)", id, path, property.Raw())
			case core.TypeBoolean:
				scimFilter = fmt.Sprintf("(id ne \"%s\") and (%s eq %t)", id, path, property.Raw())
			default:
				panic("invalid attribute")
			}
		}
	}

	n, err := provider.Count(ctx, scimFilter)
	if err != nil {
		return core.Errors.Uniqueness("failed to check uniqueness for '%s': %s",
			property.Attribute().DisplayName(),
			err.Error(),
		)
	} else if n > 0 {
		return core.Errors.Uniqueness("value of '%s' does not satisfy constraint uniqueness=%s",
			property.Attribute().DisplayName(),
			property.Attribute().Uniqueness.String(),
		)
	}

	return nil
}
