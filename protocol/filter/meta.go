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
	// Implementation of PropertyFilter to assign the resource type to the field annotated with '@meta.resourceType'.
	// When the reference resource is present, this filter also ensures the resource type of the resource and the reference
	// is the same. Since this filter carries out part of the responsibility of a readOnly check, the field annotated
	// with '@meta.resourceType' may also be annotated with '@skipReadOnly' to avoid multiple copying.
	metaResourceTypeFilter struct {}

	// Implementation of PropertyFilter to assign the current time to the property annotated with '@meta.created' only
	// when the field is unassigned. This filter is suggested to be used in combination of a readOnly filter, so any
	// existing created time from the reference resource may be copied over first.
	metaCreatedFilter struct {}

	// Implementation of PropertyFilter to assign the current time to the property annotated with '@meta.lastModified'.
	// Because this filter always generated a new value, it is suggested to combine this filter with the use of annotation
	// '@skipReadOnly' to avoid wasted copying.
	metaLastModifiedFilter struct {}

	// Implementation of PropertyFilter to assign the resource URL to the property annotated with '@meta.location' only
	// when the field is unassigned.
	//
	// The filter must be configured with all url formats corresponding to necessary resource type ids. The url format
	// must be a single argument string format that expects an id. This filter expects the resource id already been set
	// on the resource. The location generation only works with Filter, the method FilterWithRef only copied value over.
	metaLocationFilter struct {
		locationFormats map[string]string
	}

	// Implementation of PropertyFilter to assign a version to the property annotated with '@meta.version'. The filter
	// does not distinguish between modified and un-modified resource, and assigns a new version every time it's called.
	// The version will consist an sha256 hash of resource id, timestamp, and random number in range [0, 10000).
	// Because this filter always generated a new value, it is suggested to combine this filter with the use of annotation
	// '@skipReadOnly' to avoid wasted copying.
	metaVersionFilter struct {}
)

// --- metaResourceTypeFilter ---

func (f *metaResourceTypeFilter) Supports(attribute *core.Attribute) bool {
	return ContainsAnnotation(attribute, "@meta.resourceType")
}

func (f *metaResourceTypeFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaResourceTypeFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource) error {
	if ref != nil && resource.GetResourceType().Name != ref.GetResourceType().Name {
		return core.Errors.Internal("unexpected change of resource type")
	}
	return property.(core.Crud).Replace(nil, resource.GetResourceType().Name)
}

// --- metaCreatedFilter ---

func (f *metaCreatedFilter) Supports(attribute *core.Attribute) bool {
	return ContainsAnnotation(attribute, "@meta.created")
}

func (f *metaCreatedFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaCreatedFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property, ref *core.Resource) error {
	if property.IsUnassigned() {
		return property.(core.Crud).Replace(nil, time.Now().Format(core.ISO8601))
	}
	return nil
}

// --- metaLastModifiedFilter ---

func (metaLastModifiedFilter) Supports(attribute *core.Attribute) bool {
	return ContainsAnnotation(attribute, "@meta.lastModified")
}

func (metaLastModifiedFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (metaLastModifiedFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property, _ *core.Resource) error {
	return property.(core.Crud).Replace(nil, time.Now().Format(core.ISO8601))
}

// --- metaLocationFilter ---

func (f *metaLocationFilter) Supports(attribute *core.Attribute) bool {
	return ContainsAnnotation(attribute, "@meta.location")
}

func (f *metaLocationFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaLocationFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property, _ *core.Resource) error {
	if !property.IsUnassigned() {
		return nil
	}

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

// --- metaVersionFilter ---

func (f *metaVersionFilter) Supports(attribute *core.Attribute) bool {
	return ContainsAnnotation(attribute, "@meta.version")
}

func (f *metaVersionFilter) Order(attribute *core.Attribute) int {
	return 200
}

func (f *metaVersionFilter) Filter(ctx context.Context, resource *core.Resource, property core.Property, _ *core.Resource) error {
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
