package filter

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"math/rand"
	"strings"
	"time"
)

// MetaFilter returns a ByResource filter that assigns and updates the meta core attribute.
func MetaFilter() ByResource {
	return metaFilter{}
}

type metaFilter struct{}

func (f metaFilter) Filter(_ context.Context, resource *prop.Resource) error {
	nav := resource.Navigator()
	if nav.Dot("meta").HasError() {
		return nav.Error()
	}

	if err := f.assignResourceType(nav, resource.ResourceType()); err != nil {
		return err
	}
	if err := f.assignCreatedTimeToNow(nav); err != nil {
		return err
	}
	if err := f.assignLastModifiedToNow(nav); err != nil {
		return err
	}
	if err := f.assignLocation(nav, resource); err != nil {
		return err
	}
	if err := f.assignNewVersion(nav, resource); err != nil {
		return err
	}

	return nil
}

func (f metaFilter) FilterRef(_ context.Context, resource *prop.Resource, ref *prop.Resource) error {
	if resource.Hash() == ref.Hash() {
		return nil
	}

	nav := resource.Navigator()
	if nav.Dot("meta").HasError() {
		return nav.Error()
	}

	if err := f.assignLastModifiedToNow(nav); err != nil {
		return err
	}
	if err := f.assignNewVersion(nav, resource); err != nil {
		return err
	}

	return nil
}

func (f metaFilter) assignResourceType(nav prop.Navigator, resourceType *spec.ResourceType) error {
	if nav.Dot("resourceType").HasError() {
		return nav.Error()
	}
	defer nav.Retract()

	return nav.Replace(resourceType.ID()).Error()
}

func (f metaFilter) assignCreatedTimeToNow(nav prop.Navigator) error {
	if nav.Dot("created").HasError() {
		return nav.Error()
	}
	defer nav.Retract()

	return nav.Replace(time.Now().Format(spec.ISO8601)).Error()
}

func (f metaFilter) assignLastModifiedToNow(nav prop.Navigator) error {
	if nav.Dot("lastModified").HasError() {
		return nav.Error()
	}
	defer nav.Retract()

	return nav.Replace(time.Now().Format(spec.ISO8601)).Error()
}

func (f metaFilter) assignLocation(nav prop.Navigator, resource *prop.Resource) error {
	if nav.Dot("location").HasError() {
		return nav.Error()
	}
	defer nav.Retract()

	id := resource.IdOrEmpty()
	if len(id) == 0 {
		return fmt.Errorf("%w: empty id", spec.ErrInternal)
	}

	location := strings.TrimSuffix(resource.ResourceType().Endpoint(), "/") + "/" + id
	return nav.Replace(location).Error()
}

func (f metaFilter) assignNewVersion(nav prop.Navigator, resource *prop.Resource) error {
	if nav.Dot("version").HasError() {
		return nav.Error()
	}
	defer nav.Retract()

	id := resource.IdOrEmpty()
	if len(id) == 0 {
		return fmt.Errorf("%w: empty id", spec.ErrInternal)
	}

	ts := rand.Uint64()
	tsBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(tsBuf, ts)

	sha := sha1.New()
	sha.Write([]byte(id))
	sha.Write(tsBuf)
	sum := sha.Sum(nil)

	return nav.Replace(fmt.Sprintf("W/\"%x\"", sum)).Error()
}
