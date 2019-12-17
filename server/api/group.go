package api

import (
	"context"
	"github.com/imulab/go-scim/core/prop"
	"github.com/imulab/go-scim/protocol/groupsync"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/nats-io/nats.go"
)

type groupListener struct {
	nc     *nats.Conn
	logger log.Logger
}

func (l *groupListener) ResourceCreated(ctx context.Context, created *prop.Resource) {
	l.notify(groupsync.Compare(nil, created))
}

func (l *groupListener) ResourceUpdated(ctx context.Context, updated *prop.Resource, original *prop.Resource) {
	l.notify(groupsync.Compare(original, updated))
}

func (l *groupListener) ResourceDeleted(ctx context.Context, deleted *prop.Resource) {
	l.notify(groupsync.Compare(deleted, nil))
}

func (l *groupListener) notify(diff *groupsync.Diff) {
	if diff.CountJoined() + diff.CountLeft() == 0 {
		return
	}
}
