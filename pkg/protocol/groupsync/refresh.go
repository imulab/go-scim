package groupsync

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/protocol/crud"
	"github.com/imulab/go-scim/pkg/protocol/db"
)

const (
	fieldGroups      = "groups"
	fieldMeta        = "meta"
	fieldLocation    = "location"
	fieldDisplayName = "displayName"
	fieldRef         = "$ref"
	fieldDisplay     = "display"
	fieldType        = "type"

	groupTypeDirect   = "direct"
	groupTypeIndirect = "indirect"
)

// Create a new refresher in order to refresh the group property of the user resource.
func Refresher(groupDB db.DB) *refresher {
	return &refresher{
		groupDB: groupDB,
		visited: make(map[string]struct{}),
	}
}

type refresher struct {
	groupDB db.DB
	visited map[string]struct{}
}

// Performs a complete refresh on the user's group property. This method assumes the caller holds the pessimistic lock
// on the user and therefore do not attempt to further acquire the lock. A recursive query will be perform on the group
// resource database to find out all direct and indirect groups that this user is a member of. The resulting data will
// be formulated and overwrite the group property of the user resource.
func (r *refresher) Refresh(ctx context.Context, user *prop.Resource) error {
	addGroup, err := r.groupAddHandle(user)
	if err != nil {
		return err
	}

	type refreshTask struct {
		memberID  string
		groupType string
	}
	var tasks = []*refreshTask{
		{
			memberID:  user.ID(),
			groupType: groupTypeDirect,
		},
	}

	for len(tasks) > 0 {
		t := tasks[0]
		tasks = tasks[1:]

		groups, err := r.groupDB.Query(ctx, fmt.Sprintf(
			"members.value eq \"%s\"", t.memberID),
			nil,
			nil,
			&crud.Projection{Attributes: []string{"id", "meta.location", "displayName"}})
		if err != nil {
			return err
		}

		for _, g := range groups {
			if v, err := r.makeGroup(g, t.groupType); err != nil {
				return err
			} else if err := addGroup(v); err != nil {
				return err
			}

			// check with visited record to avoid cyclic path
			gid := g.ID()
			if _, ok := r.visited[gid]; !ok {
				tasks = append(tasks, &refreshTask{
					memberID:  gid,
					groupType: groupTypeIndirect,
				})
				r.visited[gid] = struct{}{}
			}
		}
	}

	return nil
}

func (r *refresher) groupAddHandle(user *prop.Resource) (groupAddFunc func(value map[string]interface{}) error, err error) {
	nav := user.NewNavigator()
	if _, err = nav.FocusName(fieldGroups); err != nil {
		return
	}
	if err = nav.Current().Delete(); err != nil {
		return
	}
	groupAddFunc = func(value map[string]interface{}) error {
		return nav.Current().Add(value)
	}
	return
}

func (r *refresher) makeGroup(g *prop.Resource, groupType string) (map[string]interface{}, error) {
	var v = make(map[string]interface{})
	v[fieldValue] = g.ID()
	v[fieldType] = groupType
	if nav := g.NewFluentNavigator().FocusName(fieldMeta).FocusName(fieldLocation); nav.Error() != nil {
		return nil, nav.Error()
	} else {
		v[fieldRef] = nav.Current().Raw()
	}
	if nav := g.NewFluentNavigator().FocusName(fieldDisplayName); nav.Error() != nil {
		return nil, nav.Error()
	} else {
		v[fieldDisplay] = nav.Current().Raw()
	}
	return v, nil
}
