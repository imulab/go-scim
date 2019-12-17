package groupsync

import (
	"context"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/protocol/event"
	"github.com/imulab/go-scim/protocol/groupsync"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/nats-io/nats.go"
)

func Sender(nc *nats.Conn, logger log.Logger) (pub event.Publisher, closer func(), err error) {
	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		return
	}

	pub = &sender{
		ec:     ec,
		logger: logger,
	}
	closer = func() {
		ec.Close()
	}
	return
}

type sender struct {
	ec     *nats.EncodedConn
	logger log.Logger
}

func (l *sender) ResourceCreated(_ context.Context, created *prop.Resource) {
	go l.notify(created, groupsync.Compare(nil, created))
}

func (l *sender) ResourceUpdated(_ context.Context, updated *prop.Resource, original *prop.Resource) {
	go l.notify(updated, groupsync.Compare(original, updated))
}

func (l *sender) ResourceDeleted(_ context.Context, deleted *prop.Resource) {
	go l.notify(deleted, groupsync.Compare(deleted, nil))
}

func (l *sender) notify(group *prop.Resource, diff *groupsync.Diff) {
	if diff.CountJoined()+diff.CountLeft() == 0 {
		return
	}
	groupId := group.ID()
	diff.ForEachLeft(func(id string) {
		l.sendMessage(groupId, id)
	})
	diff.ForEachJoined(func(id string) {
		l.sendMessage(groupId, id)
	})
}

func (l *sender) sendMessage(groupId string, memberId string) {
	msg := &message{
		GroupID:  groupId,
		MemberID: memberId,
	}
	if err := l.ec.Publish(subject, msg); err != nil {
		msg.logFailed(l.logger)
	} else {
		msg.logSent(l.logger)
	}
}
