package groupsync

import (
	"github.com/imulab/go-scim/core/expr"
	"github.com/imulab/go-scim/core/spec"
	scimmongo "github.com/imulab/go-scim/mongo"
	"github.com/imulab/go-scim/protocol/db"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/imulab/go-scim/server/logger"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type appContext struct {
	serviceProviderConfig *spec.ServiceProviderConfig
	rabbitChannel         *amqp.Channel
	mongoClient           *mongo.Client
	logger                log.Logger
	userResourceType      *spec.ResourceType
	groupResourceType     *spec.ResourceType
	userDB                db.DB
	groupDB               db.DB
}

func (c *appContext) initialize(args *arguments) (err error) {
	// logger
	c.logger = args.Logger()

	// service provider config
	if c.serviceProviderConfig, err = args.ParseServiceProviderConfig(); err != nil {
		return
	}

	// schemas
	if schemas, err := args.ParseSchemas(); err != nil {
		return err
	} else {
		for _, sch := range schemas {
			spec.SchemaHub.Put(sch)
		}
	}

	// resource type
	if c.userResourceType, err = args.ParseUserResourceType(); err != nil {
		return
	} else if c.groupResourceType, err = args.ParseGroupResourceType(); err != nil {
		return
	} else {
		expr.Register(c.userResourceType)
		expr.Register(c.groupResourceType)
	}

	// rabbit
	if c.rabbitChannel, err = args.Rabbit.Connect(); err != nil {
		return
	}

	// mongo
	if !args.MemoryDB {
		if c.mongoClient, err = args.Mongo.Connect(); err != nil {
			return
		}
		if mdBytes, err := args.Mongo.ReadMetadataBytes(); err != nil {
			return err
		} else {
			for _, md := range mdBytes {
				if err := scimmongo.ReadMetadata(md); err != nil {
					return err
				}
			}
		}
	}

	c.loadUserDatabase(args)
	c.loadGroupDatabase(args)

	return nil
}

func (c *appContext) loadLogger() error {
	c.logger = logger.Zero()
	return nil
}

func (c *appContext) loadUserDatabase(args *arguments) {
	if args.MemoryDB {
		c.userDB = db.Memory()
	} else {
		coll := c.mongoClient.Database(args.Mongo.Database, options.Database()).Collection(c.userResourceType.Name(), options.Collection())
		c.userDB = scimmongo.DB(c.userResourceType, c.logger, coll, scimmongo.Options())
	}
}

func (c *appContext) loadGroupDatabase(args *arguments) {
	if args.MemoryDB {
		c.groupDB = db.Memory()
	} else {
		coll := c.mongoClient.Database(args.Mongo.Database, options.Database()).Collection(c.groupResourceType.Name(), options.Collection())
		c.groupDB = scimmongo.DB(c.groupResourceType, c.logger, coll, scimmongo.Options())
	}
}
