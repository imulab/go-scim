package filters

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
	"github.com/imulab/go-scim/src/core/prop"
	"github.com/imulab/go-scim/src/protocol"
	"math/rand"
	"time"
)

const (
	fieldMeta         = "meta"
	fieldResourceType = "resourceType"
	fieldCreated      = "created"
	fieldLastModified = "lastModified"
	fieldLocation     = "location"
	fieldVersion      = "version"
)

// Create a new meta resource filter. The filter assigns metadata to the meta field on Filter. On FilterWithRef, the filter
// updates lastModified time and version only if hash has changed since baseline (key BaselineHashKey in context).
func NewMetaResourceFilter(order int) protocol.ResourceFilter {
	return &metaResourceFilter{order: order}
}

type metaResourceFilter struct {
	order int
}

func (f *metaResourceFilter) Order() int {
	return f.order
}

func (f *metaResourceFilter) Filter(ctx *protocol.FilterContext, resource *prop.Resource) error {
	meta, err := resource.NewNavigator().FocusName(fieldMeta)
	if err != nil {
		return err
	}

	if err := f.assignResourceType(meta, resource.ResourceType()); err != nil {
		return err
	}

	if err := f.assignCreatedTimeToNow(meta); err != nil {
		return err
	}

	if err := f.assignLastModifiedTimeToNow(meta); err != nil {
		return err
	}

	if err := f.assignLocation(meta, resource); err != nil {
		return err
	}

	if err := f.updateVersion(meta, resource); err != nil {
		return err
	}

	return nil
}

func (f *metaResourceFilter) FilterRef(ctx *protocol.FilterContext, resource *prop.Resource, ref *prop.Resource) error {
	resourceHash := resource.Hash()
	baselineHash, ok := ctx.Get(BaselineHashKey{})

	if !ok || resourceHash != baselineHash.(uint64) {
		meta, err := resource.NewNavigator().FocusName(fieldMeta)
		if err != nil {
			return err
		}

		if err := f.assignLastModifiedTimeToNow(meta); err != nil {
			return err
		}

		if err := f.updateVersion(meta, resource); err != nil {
			return err
		}
	}

	return nil
}

func (f *metaResourceFilter) assignResourceType(meta core.Property, resourceType *core.ResourceType) error {
	p, err := prop.NewNavigator(meta).FocusName(fieldResourceType)
	if err != nil {
		return err
	}

	_, err = prop.Internal(p).Replace(resourceType.Name())
	if err != nil {
		return err
	}

	return nil
}

func (f *metaResourceFilter) assignCreatedTimeToNow(meta core.Property) error {
	p, err := prop.NewNavigator(meta).FocusName(fieldCreated)
	if err != nil {
		return err
	}

	_, err = prop.Internal(p).Replace(time.Now().Format(prop.ISO8601))
	if err != nil {
		return err
	}

	return nil
}

func (f *metaResourceFilter) assignLastModifiedTimeToNow(meta core.Property) error {
	p, err := prop.NewNavigator(meta).FocusName(fieldLastModified)
	if err != nil {
		return err
	}

	_, err = prop.Internal(p).Replace(time.Now().Format(prop.ISO8601))
	if err != nil {
		return err
	}

	return nil
}

func (f *metaResourceFilter) assignLocation(meta core.Property, resource *prop.Resource) error {
	id := resource.ID()
	if len(id) == 0 {
		return errors.Internal("id is not assigned to resource")
	}

	p, err := prop.NewNavigator(meta).FocusName(fieldLocation)
	if err != nil {
		return err
	}

	_, err = prop.Internal(p).Replace(fmt.Sprintf("%s/%s", resource.ResourceType().Endpoint(), id))
	if err != nil {
		return err
	}

	return nil
}

func (f *metaResourceFilter) updateVersion(meta core.Property, resource *prop.Resource) error {
	id := resource.ID()
	if len(id) == 0 {
		return errors.Internal("id is not assigned to resource")
	}

	ts := rand.Uint64()
	tsBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(tsBuf, ts)

	sha := sha1.New()
	sha.Write([]byte(id))
	sha.Write(tsBuf)
	sum := sha.Sum(nil)

	p, err := prop.NewNavigator(meta).FocusName(fieldVersion)
	if err != nil {
		return err
	}

	_, err = prop.Internal(p).Replace(fmt.Sprintf("W/\"%x\"", sum))
	if err != nil {
		return err
	}

	return nil
}
