package groupsync

import (
	"context"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/protocol/crud"
	"github.com/imulab/go-scim/protocol/db"
	"github.com/imulab/go-scim/protocol/event"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/imulab/go-scim/protocol/services/filter"
)

const (
	joined = "joined"
	left   = "left"
)

// Create a event listener for the group resource to create GroupSync tasks from the differences
// of the before and after state of the same group resource. The created GroupSync task will be
// saved into database and read by another process to sync to the group property of the User resource.
// This listener runs asynchronously and dumps all errors to logs.
func Listener(groupSyncDB db.DB, logger log.Logger) event.Publisher {
	return &membershipListener{
		groupSyncDB: groupSyncDB,
		log:         logger,
		filters: []filter.ForResource{
			filter.ID(),
			filter.Meta(),
		},
	}
}

type membershipListener struct {
	groupSyncDB db.DB
	filters     []filter.ForResource
	log         log.Logger
}

func (l *membershipListener) ResourceCreated(_ context.Context, created *prop.Resource) {
	go l.run(nil, created)
}

func (l *membershipListener) ResourceUpdated(_ context.Context, updated *prop.Resource, original *prop.Resource) {
	go l.run(original, updated)
}

func (l *membershipListener) ResourceDeleted(_ context.Context, deleted *prop.Resource) {
	go l.run(deleted, nil)
}

func (l *membershipListener) run(before *prop.Resource, after *prop.Resource) {
	diff := Compare(before, after)

	var (
		gs  *prop.Resource
		err error
	)
	{
		if before != nil {
			gs, err = l.makeGroupSync(before, diff)
		} else {
			gs, err = l.makeGroupSync(after, diff)
		}

		if err != nil {
			l.dumpErrorToLog(before, after, err)
			return
		}

		if p, err := gs.NewNavigator().FocusName("diff"); err != nil {
			l.dumpErrorToLog(before, after, err)
			return
		} else if p.(prop.Container).CountChildren() == 0 {
			return // no need to create group sync task if members haven't changed
		}
	}

	err = l.groupSyncDB.Insert(context.Background(), gs)
	if err != nil {
		l.dumpErrorToLog(before, after, err)
	}
}

func (l *membershipListener) makeGroupSync(group *prop.Resource, diff *Diff) (*prop.Resource, error) {
	gs := prop.NewResource(ResourceType())

	for _, f := range l.filters {
		if err := f.Filter(context.Background(), gs); err != nil {
			return nil, err
		}
	}

	if err := l.fillGroup(gs, group); err != nil {
		return nil, err
	}

	var diffErr error
	{
		diff.ForEachJoined(func(id string) {
			if err := l.addDiff(gs, id, joined); err != nil {
				diffErr = err
			}
		})
		diff.ForEachLeft(func(id string) {
			if err := l.addDiff(gs, id, left); err != nil {
				diffErr = err
			}
		})
	}
	if diffErr != nil {
		return nil, diffErr
	}

	return gs, nil
}

func (l *membershipListener) addDiff(gs *prop.Resource, id string, status string) error {
	return crud.Add(gs, "diff", map[string]interface{}{
		"id":     id,
		"type":   "unknown",
		"status": status,
	})
}

func (l *membershipListener) dumpErrorToLog(before *prop.Resource, after *prop.Resource, err error) {
	var groupID string
	{
		if before != nil {
			groupID = before.ID()
		} else {
			groupID = after.ID()
		}
	}
	l.log.Error("failed to create group sync job [id=%s]: %s", groupID, err.Error())
}

func (l *membershipListener) fillGroup(gs *prop.Resource, group *prop.Resource) error {
	if err := crud.Replace(gs, "group.id", group.ID()); err != nil {
		return err
	}
	if err := crud.Replace(gs, "group.location", group.Location()); err != nil {
		return err
	}
	var displayName interface{}
	{
		if p, err := group.NewNavigator().FocusName("displayName"); err != nil {
			return err
		} else {
			displayName = p.Raw()
		}
	}
	if err := crud.Replace(gs, "group.display", displayName); err != nil {
		return err
	}
	return nil
}
