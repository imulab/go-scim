package groupsync

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/pkg/v2/crud"
	"github.com/imulab/go-scim/pkg/v2/db"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"strconv"
)

// NewSyncService returns a new SyncService.
func NewSyncService(groupDB db.DB) *SyncService {
	s := SyncService{groupDB: groupDB}
	return &s
}

// SyncService synchronizes the user resource's "groups" property.
type SyncService struct {
	groupDB db.DB
}

// SyncGroupPropertyForUser updates the user's "groups" property, according to the latest state in Group resources. This
// method does not save or replace the updated resource with the database. It is up to the caller to do so.
//
// Due to nested membership, this method may search the group database multiple times, which may turn out to be a lengthy
// process. The ctx context can be used to set a timeline or cancel the processing, this method will respect that at
// appropriate intervals.
func (s *SyncService) SyncGroupPropertyForUser(ctx context.Context, user *prop.Resource) error {
	groupNav := user.Navigator().Dot("groups")
	if groupNav.HasError() {
		return groupNav.Error()
	}

	// clear the group property
	if groupNav.Delete().HasError() {
		return groupNav.Error()
	}

	// task definition and queue
	type task struct {
		member string
		direct bool
	}
	tasks := []task{
		{member: user.IdOrEmpty(), direct: true},
	}

	// map to record the completed group ids, so we don't fall into cycles
	completed := map[string]struct{}{}

	for len(tasks) > 0 {
		// check if context was closed
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// pop first task
		t := tasks[0]
		tasks = tasks[1:]

		groups, err := s.searchGroupsForMember(ctx, t.member)
		if err != nil {
			return err
		}
		for _, group := range groups {
			// create new group element and modify the value
			if err := func() error {
				index := groupNav.Current().(interface {
					AppendElement() int
				}).AppendElement()

				if groupNav.At(index); groupNav.HasError() {
					return groupNav.Error()
				}
				defer groupNav.Retract()

				return groupNav.Replace(s.formulateGroupElementData(group, t.direct)).Error()
			}(); err != nil {
				return err
			}

			// submit new indirect tasks if not processed before
			groupId := group.IdOrEmpty()
			if _, processed := completed[groupId]; !processed {
				tasks = append(tasks, task{
					member: groupId,
					direct: false,
				})
			}
		}

		// mark this task as completed
		completed[t.member] = struct{}{}
	}

	return nil
}

func (s *SyncService) formulateGroupElementData(group *prop.Resource, direct bool) map[string]interface{} {
	data := map[string]interface{}{
		"value":   group.IdOrEmpty(),
		"$ref":    group.MetaLocationOrEmpty(),
		"display": group.Navigator().Dot("displayName").Current().Raw(),
	}
	if direct {
		data["type"] = "direct"
	} else {
		data["type"] = "indirect"
	}
	return data
}

func (s *SyncService) searchGroupsForMember(ctx context.Context, member string) ([]*prop.Resource, error) {
	filter := fmt.Sprintf("members.value eq %s", strconv.Quote(member))
	return s.groupDB.Query(ctx, filter, nil, nil, &crud.Projection{
		Attributes: []string{"id", "meta.location", "displayName"},
	})
}
