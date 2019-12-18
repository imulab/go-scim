package groupsync

import (
	"github.com/imulab/go-scim/protocol/db"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/imulab/go-scim/server/logger"
	"github.com/streadway/amqp"
)

type appContext struct {
	logger   log.Logger
	userDB   db.DB
	groupDB  db.DB
	rabbitCh *amqp.Channel
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
	if err := c.loadRabbitMqChannel(args); err != nil {
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

func (c *appContext) loadRabbitMqChannel(args *args) error {
	conn, err := amqp.Dial(args.rabbitMqAddress)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	c.rabbitCh = ch
	return nil
}
