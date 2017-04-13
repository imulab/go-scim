package shared

import (
	"github.com/satori/go.uuid"
	"time"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
)

func NewReadOnlyAssignment(resourceType, locationTemplate string) ReadOnlyAssignment {
	return &readOnlyAssignment{resourceType:resourceType, locationTemplate:locationTemplate}
}

type ReadOnlyAssignment interface {
	AssignId(r *Resource)
	AssignMeta(r *Resource)
	UpdateMeta(r *Resource)
}

type readOnlyAssignment struct {
	resourceType string
	locationTemplate string
}

func (ro *readOnlyAssignment) AssignId(r *Resource) {
	r.Complex["id"] = uuid.NewV4().String()
}

func (ro *readOnlyAssignment) AssignMeta(r *Resource) {
	now := ro.timestamp()
	id := r.Complex["id"].(string)
	meta := map[string]interface{}{
		"created": now,
		"lastModified": now,
		"version": ro.generateVersion(id, now),
		"resourceType": ro.resourceType,
		"location": fmt.Sprintf(ro.locationTemplate, id),
	}
	r.Complex["meta"] = meta
}

func (ro *readOnlyAssignment) UpdateMeta(r *Resource) {
	id := r.Complex["id"].(string)
	meta := r.Complex["meta"].(map[string]interface{})

	now := ro.timestamp()
	meta["lastModified"] = now
	meta["version"] = ro.generateVersion(id, now)
	r.Complex["meta"] = meta
}

func (ro *readOnlyAssignment) timestamp() string {
	return time.Now().Format("2006-01-02T15:04:05Z")
}

func (ro *readOnlyAssignment) generateVersion(args ...string) string {
	hash := sha1.New()
	for _, arg := range args {
		hash.Write([]byte(arg))
	}
	return fmt.Sprintf("W/\"%s\"", base64.StdEncoding.EncodeToString(hash.Sum(nil)))
}