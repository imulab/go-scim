package groupsync

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/pkg/core/errors"
	"github.com/imulab/go-scim/pkg/core/prop"
	"github.com/imulab/go-scim/pkg/protocol/crud"
	"github.com/imulab/go-scim/pkg/protocol/db"
	"github.com/imulab/go-scim/pkg/protocol/lock"
	"github.com/imulab/go-scim/pkg/protocol/log"
	"time"
)

type worker struct {
	userDB      db.DB
	groupDB     db.DB
	groupSyncDB db.DB
	lock        lock.Lock
	log         log.Logger
	errChan     chan error
	doneChan    chan struct{}
}

func (w *worker) Start() chan error {
	go w.start()
	return w.errChan
}

func (w *worker) Stop() {
	w.doneChan <- struct{}{}
}

func (w *worker) withLock(resource *prop.Resource, f func() error) {
	defer w.lock.Unlock(context.Background(), resource)
	if err := w.lock.Lock(context.Background(), resource); err != nil {
		w.errChan <- err
		return
	}
	if err := f(); err != nil {
		w.errChan <- err
	}
}

func (w *worker) start() {
	for {
		select {
		case <-w.doneChan:
			return
		default:
			sync := w.nextSync()
			if sync != nil {
				w.withLock(sync, func() error {
					for w.syncTop(sync) {
					}
					return nil
				})
			}
			w.sleep(5)
		}
	}
}

func (w *worker) syncTop(sync *prop.Resource) bool {
	beforeHash := sync.Hash()

	nav := sync.NewFluentNavigator().FocusName("diff")
	diffs, err := nav.CurrentAsContainer(), nav.Error()
	if err != nil {
		w.errChan <- err
		return false
	}

	// Whatever the processing result, save or delete the group sync resource.
	defer func() {
		if diffs.CountChildren() == 0 {
			if err := w.groupSyncDB.Delete(context.Background(), sync.ID()); err != nil {
				w.errChan <- err
			}
		} else if beforeHash != sync.Hash() {
			if err := w.groupSyncDB.Replace(context.Background(), sync); err != nil {
				w.errChan <- err
			}
		}
	}()

	if diffs.CountChildren() == 0 {
		return false
	}

	var (
		topDiff = diffs.ChildAtIndex(0).(prop.Container)
		topType = topDiff.ChildAtIndex("type")
	)
	if topType.Raw() == "unknown" {
		// the type of the diff is unknown, we will try count user by that id.
		// if > 0, it is a user; if == 0, it is a group
		if n, err := w.userDB.Count(
			context.Background(),
			fmt.Sprintf("id eq \"%s\"", topDiff.ChildAtIndex("id").Raw()),
		); err != nil {
			w.errChan <- err
			return false
		} else {
			if n > 0 {
				_ = topType.Replace("direct")
			} else {
				_ = topType.Replace("indirect")
			}
		}
	}

	switch topType.Raw() {
	case "direct":
		if err := w.syncDirect(sync, topDiff); err != nil {
			w.errChan <- err
			return false
		}
	case "indirect":
		if err := w.expandIndirect(diffs, topDiff); err != nil {
			w.errChan <- err
			return false
		}
	default:
		w.errChan <- errors.Internal("invalid type value for diff")
		return false
	}

	return diffs.CountChildren() > 0
}

func (w *worker) syncDirect(sync *prop.Resource, diff prop.Container) error {
	var (
		userID    = diff.ChildAtIndex("id").Raw().(string)
		status    = diff.ChildAtIndex("status").Raw().(string)
		groupData = make(map[string]interface{})
	)
	{
		nav := sync.NewFluentNavigator().FocusName("group")
		groupData["type"] = "direct"

		nav.FocusName("id")
		if nav.Error() != nil {
			return nav.Error()
		}
		groupData["value"] = nav.Current().Raw()
		nav.Retract()

		nav.FocusName("location")
		if nav.Error() != nil {
			return nav.Error()
		}
		groupData["$ref"] = nav.Current().Raw()
		nav.Retract()

		nav.FocusName("display")
		if nav.Error() != nil {
			return nav.Error()
		}
		groupData["display"] = nav.Current().Raw()
		nav.Retract()
	}

	user, err := w.userDB.Get(context.Background(), userID, nil)
	if err != nil {
		return err
	}

	w.withLock(user, func() error {
		nav := user.NewFluentNavigator().FocusName("groups")
		if nav.Error() != nil {
			return nav.Error()
		}
		switch status {
		case joined:
			if err := nav.Current().Add(groupData); err != nil {
				return err
			}
		case left:
			if err := crud.Delete(user, fmt.Sprintf("groups[value eq \"%s\"]", groupData["id"])); err != nil {
				return err
			}
		}
		return w.userDB.Replace(context.Background(), user)
	})

	return diff.Delete()
}

func (w *worker) expandIndirect(diffs prop.Container, top prop.Container) error {
	groupID := top.ChildAtIndex("id").Raw().(string)

	group, err := w.groupDB.Get(context.Background(), groupID, nil)
	if err != nil {
		return err
	}

	nav := group.NewFluentNavigator().FocusName("members")
	if nav.Error() != nil {
		return nav.Error()
	}
	_ = nav.CurrentAsContainer().ForEachChild(func(_ int, child prop.Property) error {
		memberData := make(map[string]interface{})
		memberData["id"] = prop.NewFluentNavigator(child).FocusName("id").Current().Raw()
		memberData["type"] = "unknown"
		memberData["status"] = top.ChildAtIndex("status").Raw()
		return diffs.Add(memberData)
	})

	return top.Delete()
}

func (w *worker) nextSync() *prop.Resource {
	results, err := w.groupSyncDB.Query(
		context.Background(),
		"id pr",
		&crud.Sort{By: "meta.created", Order: crud.SortAsc},
		&crud.Pagination{StartIndex: 1, Count: 1},
		nil,
	)

	if err != nil {
		w.errChan <- err
		return nil
	}

	if len(results) == 0 {
		return nil
	} else {
		return results[0]
	}
}

func (w *worker) sleep(seconds int) {
	w.log.Info("sleeping for %d seconds before continue.", seconds)
	timer := time.NewTimer(time.Duration(seconds) * time.Second)
	<-timer.C
}
