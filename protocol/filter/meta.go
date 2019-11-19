package filter

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/imulab/go-scim/core"
	"math/rand"
	"time"
)

type (
	// Implementation of PropertyFilter to assign the resource type to the filed meta.resourceType. This filter
	// copies the resource's resource type name to the field. It only works on the field annotated with
	// '@meta.resourceType'. Behaviour is identical with or without reference property.
	metaResourceTypeFilter struct {}

	// Implementation of PropertyFilter to assign the current time to the property annotated with '@meta.created'.
	// The time generation only works in Filter method. In FilterWithRef, the value is copied over.
	metaCreatedFilter struct {}

	// Implementation of PropertyFilter to assign the current time to the property annotated with '@meta.lastModified'.
	// The time generation works in both Filter and FilterWithRef method.
	metaLastModifiedFilter struct {}

	// Implementation of PropertyFilter to assign the resource URL to the property annotated with '@meta.location'.
	// The filter must be configured with all url formats corresponding to necessary resource type ids. The url format
	// must be a single argument string format that expects an id. This filter expects the resource id already been set
	// on the resource. The location generation only works with Filter, the method FilterWithRef only copied value over.
	metaLocationFilter struct {
		locationFormats map[string]string
	}

	// Implementation of PropertyFilter to assign a version to the property annotated with '@meta.version'. The filter
	// does not distinguish between modified and un-modified resource, and assigns a new version every time it's called.
	// The version will consist an sha256 hash of resource id, timestamp, and random number in range [0, 10000).
	metaVersionFilter struct {}
)

// --- metaResourceTypeFilter ---

func (f *metaResourceTypeFilter) Supports(attribute *core.Attribute) bool {
	return ContainsAnnotation(attribute, "@meta.resourceType")
}

func (f *metaResourceTypeFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaResourceTypeFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	return property.(core.Crud).Replace(nil, resource.GetResourceType().Name)
}

func (f *metaResourceTypeFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, _ *core.Resource, _ core.Property) error {
	return f.Filter(ctx, resource, property)
}

// --- metaCreatedFilter ---

func (f *metaCreatedFilter) Supports(attribute *core.Attribute) bool {
	return ContainsAnnotation(attribute, "@meta.created")
}

func (f *metaCreatedFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaCreatedFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	return property.(core.Crud).Replace(nil, time.Now().Format(core.ISO8601))
}

func (f *metaCreatedFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, refResource *core.Resource, ref core.Property) error {
	return property.(core.Crud).Replace(nil, ref.Raw())
}

// --- metaLastModifiedFilter ---

func (metaLastModifiedFilter) Supports(attribute *core.Attribute) bool {
	return ContainsAnnotation(attribute, "@meta.lastModified")
}

func (metaLastModifiedFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (metaLastModifiedFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	return property.(core.Crud).Replace(nil, time.Now().Format(core.ISO8601))
}

func (metaLastModifiedFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, refResource *core.Resource, ref core.Property) error {
	return property.(core.Crud).Replace(nil, time.Now().Format(core.ISO8601))
}

// --- metaLocationFilter ---

func (f *metaLocationFilter) Supports(attribute *core.Attribute) bool {
	return ContainsAnnotation(attribute, "@meta.location")
}

func (f *metaLocationFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaLocationFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	id, err := resource.Get(core.Steps.NewPath("id"))
	if err != nil || id == nil || len(id.(string)) == 0 {
		return core.Errors.Internal("could not locate resource id")
	}

	format := f.locationFormats[resource.GetResourceType().Id]
	if len(format) == 0 {
		panic("location url formats for all resource types must be set in metaFilter")
	}

	return property.(core.Crud).Replace(nil, fmt.Sprintf(format, id))
}

func (f *metaLocationFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, refResource *core.Resource, ref core.Property) error {
	return property.(core.Crud).Replace(nil, ref.Raw())
}

// --- metaVersionFilter ---

func (f *metaVersionFilter) Supports(attribute *core.Attribute) bool {
	return ContainsAnnotation(attribute, "@meta.version")
}

func (f *metaVersionFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaVersionFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property) error {
	return f.generateVersion(resource, property)
}

func (f *metaVersionFilter) FilterWithRef(ctx context.Context, resource *core.Resource, property core.Property, refResource *core.Resource, ref core.Property) error {
	return f.generateVersion(resource, property)
}

func (f *metaVersionFilter) generateVersion(resource *core.Resource, property core.Property) error {
	id, err := resource.Get(core.Steps.NewPath("id"))
	if err != nil || id == nil || len(id.(string)) == 0 {
		return core.Errors.Internal("could not locate id")
	}

	sha := sha256.New()
	sum := sha.Sum([]byte(fmt.Sprintf("%s:%d:%d", id, time.Now().Unix(), rand.Intn(10000))))

	return property.(core.Crud).Replace(nil, fmt.Sprintf("W/\"%x\"", sum))
}

// --- factories ---

func NewMetaResourceTypeFilter() PropertyFilter {
	return &metaResourceTypeFilter{}
}

func NewMetaCreatedFilter() PropertyFilter {
	return &metaCreatedFilter{}
}

func NewMetaLastModifiedFilter() PropertyFilter  {
	return &metaLastModifiedFilter{}
}

func NewMetaLocationFilter(resourceTypeIdToFormat map[string]string) PropertyFilter {
	return &metaLocationFilter{
		locationFormats: resourceTypeIdToFormat,
	}
}

func NewMetaVersionFilter() PropertyFilter {
	return &metaVersionFilter{}
}
