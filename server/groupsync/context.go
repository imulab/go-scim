package groupsync

import (
	"github.com/imulab/go-scim/protocol/db"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/imulab/go-scim/server/logger"
	"github.com/nats-io/nats.go"
	"time"
)

type appContext struct {
	logger  log.Logger
	userDB  db.DB
	groupDB db.DB
	natConn *nats.Conn
}

func (c *appContext) initialize(args *args) error {
	if err := c.loadLogger(); err != nil {
		return err
	}
	if err := c.loadUserDatabase(args); err != nil {
		return err
	}
	if err := c.loadGroupDatabase(args); err != nil {
		return err
	}
	if err := c.loadNatsConnection(args); err != nil {
		return err
	}
	return nil
}

func (c *appContext) loadLogger() error {
	c.logger = logger.Zero()
	return nil
}

func (c *appContext) loadUserDatabase(args *args) error {
	c.userDB = db.Memory()
	return nil
}

func (c *appContext) loadGroupDatabase(args *args) error {
	c.groupDB = db.Memory()
	return nil
}

func (c *appContext) loadNatsConnection(args *args) (err error) {
	c.natConn, err = nats.Connect(args.natsServers, nats.Timeout(10*time.Second), nats.PingInterval(10*time.Second))
	return
}
