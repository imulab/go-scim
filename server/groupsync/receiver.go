package groupsync

import (
	"context"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/protocol/db"
	"github.com/imulab/go-scim/protocol/groupsync"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/nats-io/nats.go"
	"sync/atomic"
)

const queue = "sync"

func Receiver(nc *nats.Conn, userDB db.DB, groupDB db.DB, logger log.Logger) (*receiver, error) {
	var r = new(receiver)
	{
		r.logger = logger
		r.userDB = userDB
		r.groupDB = groupDB
		r.maxAttempt = 10
		r.safeExit = make(chan struct{}, 1)
		if ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER); err != nil {
			return nil, err
		} else {
			r.ec = ec
		}
		if sub, err := r.ec.QueueSubscribe(subject, queue, r.handle); err != nil {
			return nil, err
		} else {
			r.sub = sub
		}
	}
	return r, nil
}

type receiver struct {
	ec         *nats.EncodedConn
	sub        *nats.Subscription
	userDB     db.DB
	groupDB    db.DB
	logger     log.Logger
	cancelled  uint32
	safeExit   chan struct{} // channel to signal once all cancellation work is done
	maxAttempt int
}

func (r *receiver) handle(msg *message) {
	remaining, _, _ := r.sub.Pending()
	if r.isCancelled() {
		r.returnMessage(msg)
		if remaining <= 1 {
			r.callerCanExitSafely()
		}
	} else {
		r.syncOrExpand(msg)
	}
}

func (r *receiver) syncOrExpand(msg *message) {
	if user, _ := r.userDB.Get(context.Background(), msg.MemberID, nil); user != nil {
		r.syncUser(user, msg)
		return
	}

	if group, _ := r.groupDB.Get(context.Background(), msg.MemberID, nil); group != nil {
		r.expandGroup(group, msg)
		return
	}

	r.logger.Debug("member is neither group nor user", log.Args{
		"member": msg.MemberID,
	})
}

func (r *receiver) syncUser(user *prop.Resource, msg *message) {
	if err := groupsync.Refresher(r.groupDB).Refresh(context.Background(), user); err != nil {
		r.logErrorAndReturn(err, msg)
		return
	}
	if err := r.userDB.Replace(context.Background(), user); err != nil {
		r.logErrorAndReturn(err, msg)
		return
	}
	r.logger.Debug("synced group for user resource", log.Args{
		"user": msg.MemberID,
	})
}

func (r *receiver) expandGroup(group *prop.Resource, msg *message) {
	if nav := group.NewFluentNavigator().FocusName("members"); nav.Error() != nil {
		r.logErrorAndDiscard(nav.Error(), msg)
		return
	} else {
		_ = nav.CurrentAsContainer().ForEachChild(func(index int, child prop.Property) error {
			if vp := child.(prop.Container).ChildAtIndex("value"); vp != nil && !vp.IsUnassigned() {
				r.sendMessage(&message{
					GroupID:  msg.GroupID,
					MemberID: vp.Raw().(string),
				})
			}
			return nil
		})
		r.logger.Debug("added more sync tasks for group resource", log.Args{
			"group": msg.MemberID,
		})
	}
}

func (r *receiver) logErrorAndReturn(err error, msg *message) {
	r.logger.Error("sync encountered error", log.Args{
		"error": err,
	})
	if msg.Trial >= r.maxAttempt {
		r.logger.Debug("exceeded max attempt, will discard message", log.Args{})
	} else {
		r.logger.Debug("will return message", log.Args{})
		r.returnMessage(msg)
	}
}

func (r *receiver) logErrorAndDiscard(err error, msg *message) {
	r.logger.Error("sync encountered error", log.Args{
		"error": err,
	})
	r.logger.Debug("will discard message", log.Args{})
}

func (r *receiver) returnMessage(msg *message) {
	msg.Trial++
	if err := r.ec.Publish(subject, msg); err != nil {
		msg.logFailed(r.logger)
	} else {
		msg.logReturned(r.logger)
	}
}

func (r *receiver) sendMessage(msg *message) {
	if err := r.ec.Publish(subject, msg); err != nil {
		msg.logFailed(r.logger)
	} else {
		msg.logSent(r.logger)
	}
}

func (r *receiver) isCancelled() bool {
	return atomic.LoadUint32(&r.cancelled) == 1
}

func (r *receiver) callerCanExitSafely() {
	r.safeExit <- struct{}{}
}

func (r *receiver) Close() {
	_ = r.sub.Drain()
	atomic.StoreUint32(&r.cancelled, 1)
	<-r.safeExit
}
