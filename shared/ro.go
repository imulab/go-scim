package shared

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/satori/go.uuid"
	"math"
	"strings"
	"time"
)

func NewIdAssignment() ReadOnlyAssignment {
	return &idAssignment{}
}

func NewMetaAssignment(properties PropertySource, resourceType string) ReadOnlyAssignment {
	return &metaAssignment{PropertySource: properties, resourceType: resourceType}
}

func NewGroupAssignment(groupRepository Repository) ReadOnlyAssignment {
	return &groupAssignment{groupRepo: groupRepository}
}

type ReadOnlyAssignment interface {
	AssignValue(r *Resource, ctx context.Context) error
}

// Generates id value with UUID v4
type idAssignment struct{}

func (ro *idAssignment) AssignValue(r *Resource, ctx context.Context) error {
	uuid, err := uuid.NewV4()
	if err != nil {
		return err
	}
	r.Complex["id"] = uuid.String()
	return nil
}

// Generates new meta value for resource if meta is vacant
// Otherwise, update meta value for resource
type metaAssignment struct {
	PropertySource
	resourceType string
}

func (ro *metaAssignment) AssignValue(r *Resource, ctx context.Context) error {
	id := r.Complex["id"].(string)
	if len(id) == 0 {
		return Error.Text("Cannot assign value to meta: no id")
	}

	now := ro.timestamp()
	if meta, ok := r.Complex["meta"].(map[string]interface{}); !ok {
		propertyKey := fmt.Sprintf("scim.resources.%s.locationBase", strings.ToLower(ro.resourceType))
		locationTemplate := strings.TrimSuffix(ro.GetString(propertyKey), "/")
		if len(locationTemplate) == 0 {
			return Error.Text("Cannot assign value to meta: no resource location template configured with key %s", propertyKey)
		}
		meta := map[string]interface{}{
			"created":      now,
			"lastModified": now,
			"version":      ro.generateVersion(id, now),
			"resourceType": ro.resourceType,
			"location":     fmt.Sprintf("%s/%s", locationTemplate, id),
		}
		r.Complex["meta"] = meta
	} else {
		now := ro.timestamp()
		meta["lastModified"] = now
		meta["version"] = ro.generateVersion(id, now)
		r.Complex["meta"] = meta
	}
	return nil
}

func (ro *metaAssignment) timestamp() string {
	return time.Now().Format("2006-01-02T15:04:05Z")
}

func (ro *metaAssignment) generateVersion(args ...string) string {
	hash := sha1.New()
	for _, arg := range args {
		hash.Write([]byte(arg))
	}
	return fmt.Sprintf("W/\"%s\"", base64.StdEncoding.EncodeToString(hash.Sum(nil)))
}

type groupAssignment struct {
	groupRepo Repository
}

func (ro *groupAssignment) AssignValue(r *Resource, ctx context.Context) error {
	allGroups := make([]interface{}, 0)
	addToAllGroups := func(dp DataProvider, direct bool) {
		g := map[string]interface{}{
			"value":   dp.GetId(),
			"$ref":    dp.GetData()["meta"].(map[string]interface{})["location"],
			"display": dp.GetData()["displayName"],
		}
		if direct {
			g["type"] = "direct"
		} else {
			g["type"] = "indirect"
		}
		allGroups = append(allGroups, g)
	}

	memberIdsToSearch := NewQueueWithoutLimit()
	memberIdsToSearch.Offer(r.GetId())

	searched := make(map[string]bool)

	for memberIdsToSearch.Size() > 0 {
		id := memberIdsToSearch.Poll().(string)
		groups, err := ro.searchGroups(id)
		if err != nil {
			return err
		}

		searched[id] = true
		for _, group := range groups {
			if len(searched) == 1 {
				addToAllGroups(group, true)
			} else {
				addToAllGroups(group, false)
			}

			if _, ok := searched[group.GetId()]; !ok {
				memberIdsToSearch.Offer(group.GetId())
			}
		}
	}

	r.Complex["groups"] = allGroups
	return nil
}

func (ro *groupAssignment) searchGroups(memberId string) ([]DataProvider, error) {
	list, err := ro.groupRepo.Search(SearchRequest{
		Filter:     fmt.Sprintf("members.value eq \"%s\"", memberId),
		Count:      math.MaxInt32,
		StartIndex: 1,
	})
	if err != nil {
		return nil, Error.Text("Failed to calculate group: %s", err.Error())
	} else {
		return list.Resources, nil
	}
}
