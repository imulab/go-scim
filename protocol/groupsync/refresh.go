package groupsync

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/protocol/crud"
	"github.com/imulab/go-scim/protocol/db"
)

// Create a new refresher in order to refresh the group property of the user resource.
func Refresher(groupDB db.DB) *refresher {
	return &refresher{
		groupDB: groupDB,
		tasks:   []*refreshTask{},
		visited: make(map[string]struct{}),
	}
}

// Stateful refresher for group membership.
type refresher struct {
	groupDB db.DB
	tasks   []*refreshTask
	visited map[string]struct{}
}

// Performs a complete refresh on the user's group property. A recursive query will be performed on the group
// resource database to find out all direct and indirect groups that this user is a member of. The resulting data will
// be formulated and overwrite the group property of the user resource.
func (r *refresher) Refresh(ctx context.Context, user *prop.Resource) error {
	gm := &groupModifier{user: user}
	if err := gm.clearGroups(); err != nil {
		return err
	}

	r.submitTask(&refreshTask{
		memberID: user.ID(),
		direct:   true,
	})

	for {
		task := r.popTask()
		if task == nil {
			return nil
		}

		groups, err := r.getGroupsForMember(ctx, task.memberID)
		if err != nil {
			return err
		}

		for _, g := range groups {
			if err := gm.addGroup(g, task.direct); err != nil {
				return err
			}
			if r.markVisited(g) {
				r.submitTask(&refreshTask{
					memberID: g.ID(),
					direct:   false,
				})
			}
		}
	}
}

func (r *refresher) getGroupsForMember(ctx context.Context, memberID string) ([]*prop.Resource, error) {
	filter := fmt.Sprintf("members.value eq \"%s\"", memberID)
	projection := &crud.Projection{
		Attributes: []string{"id", "meta.location", "displayName"},
	}
	return r.groupDB.Query(ctx, filter, nil, nil, projection)
}

// Mark the group as visited. Return false if it has already been marked visited, otherwise, return true.
func (r *refresher) markVisited(group *prop.Resource) bool {
	id := group.ID()
	if _, ok := r.visited[id]; ok {
		return false
	} else {
		r.visited[id] = struct{}{}
		return true
	}
}

func (r *refresher) submitTask(task *refreshTask) {
	r.tasks = append(r.tasks, task)
}

func (r *refresher) popTask() *refreshTask {
	if len(r.tasks) == 0 {
		return nil
	}
	top := r.tasks[0]
	r.tasks = r.tasks[1:]
	return top
}

type refreshTask struct {
	memberID string
	direct   bool
}

type groupModifier struct {
	user   *prop.Resource
	groups prop.Container
}

func (gm *groupModifier) cacheGroupsProperty() error {
	if nav := gm.user.NewFluentNavigator().FocusName("groups"); nav.Error() != nil {
		return nav.Error()
	} else {
		gm.groups = nav.CurrentAsContainer()
	}
	return nil
}

func (gm *groupModifier) clearGroups() error {
	if gm.groups == nil {
		if err := gm.cacheGroupsProperty(); err != nil {
			return err
		}
	}
	return gm.groups.Delete()
}

func (gm *groupModifier) addGroup(group *prop.Resource, direct bool) error {
	v := make(map[string]interface{})
	{
		v["value"] = group.ID()
		if direct {
			v["type"] = "direct"
		} else {
			v["type"] = "indirect"
		}
		v["$ref"] = group.Location()
		if nav := group.NewFluentNavigator().FocusName("displayName"); nav.Error() == nil {
			v["display"] = nav.Current().Raw()
		}
	}
	if gm.groups == nil {
		if err := gm.cacheGroupsProperty(); err != nil {
			return err
		}
	}
	return gm.groups.Add(v)
}
