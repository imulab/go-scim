package groupsync

import (
	"context"
	gs "github.com/imulab/go-scim/cmd/internal/groupsync"
	scimmongo "github.com/imulab/go-scim/mongo/v2"
	"github.com/imulab/go-scim/pkg/v2/db"
	"github.com/imulab/go-scim/pkg/v2/groupsync"
	"github.com/imulab/go-scim/pkg/v2/service/filter"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

type applicationContext struct {
	args                      *arguments
	logger                    *zerolog.Logger
	serviceProviderConfig     *spec.ServiceProviderConfig
	registerSchemaOnce        sync.Once
	userResourceType          *spec.ResourceType
	groupResourceType         *spec.ResourceType
	userDatabase              db.DB
	groupDatabase             db.DB
	mongoClient               *mongo.Client
	registerMongoMetadataOnce sync.Once
	rabbitMqConn              *amqp.Connection
	rabbitMqChannel           *amqp.Channel
	userSyncService           *groupsync.SyncService
	messageConsumer           *consumer
}

func (ctx *applicationContext) Logger() *zerolog.Logger {
	if ctx.logger == nil {
		ctx.logger = ctx.args.Logger()
		ctx.logger.Info().Msg("logger initialized")
	}
	return ctx.logger
}

func (ctx *applicationContext) ServiceProviderConfig() *spec.ServiceProviderConfig {
	if ctx.serviceProviderConfig == nil {
		spc, err := ctx.args.ParseServiceProviderConfig()
		if err != nil {
			ctx.logInitFailure("service provider config", err)
			panic(err)
		}
		ctx.serviceProviderConfig = spc
		ctx.logInitialized("service provider config")
	}
	return ctx.serviceProviderConfig
}

func (ctx *applicationContext) UserResourceType() *spec.ResourceType {
	ctx.ensureSchemaRegistered()
	if ctx.userResourceType == nil {
		u, err := ctx.args.ParseUserResourceType()
		if err != nil {
			ctx.logInitFailure("user resource type", err)
			panic(err)
		}
		ctx.userResourceType = u
		ctx.logInitialized("user resource type")
	}
	return ctx.userResourceType
}

func (ctx *applicationContext) GroupResourceType() *spec.ResourceType {
	ctx.ensureSchemaRegistered()
	if ctx.groupResourceType == nil {
		g, err := ctx.args.ParseGroupResourceType()
		if err != nil {
			ctx.logInitFailure("group resource type", err)
			panic(err)
		}
		ctx.groupResourceType = g
		ctx.logInitialized("group resource type")
	}
	return ctx.groupResourceType
}

func (ctx *applicationContext) ensureSchemaRegistered() {
	ctx.registerSchemaOnce.Do(func() {
		if err := ctx.args.RegisterSchemas(); err != nil {
			ctx.logInitFailure("schema", err)
			panic(err)
		}
		ctx.logInitialized("schema")
	})
}

func (ctx *applicationContext) MongoClient() *mongo.Client {
	if ctx.mongoClient == nil {
		connectCtx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelFunc()

		c, err := ctx.args.MongoDB.Connect(connectCtx)
		if err != nil {
			ctx.logInitFailure("mongo client", err)
			panic(err)
		}

		ctx.mongoClient = c
		ctx.logInitialized("mongo client")
	}
	return ctx.mongoClient
}

func (ctx *applicationContext) UserDatabase() db.DB {
	if ctx.userDatabase == nil {
		if ctx.args.UseMemoryDB {
			ctx.userDatabase = db.Memory()
			ctx.logInitialized("in-memory user database")
		} else {
			ctx.ensureMongoMetadata()
			resourceType := ctx.UserResourceType()
			collection := ctx.MongoClient().
				Database(ctx.args.MongoDB.Database, options.Database()).
				Collection(resourceType.Name(), options.Collection())
			ctx.userDatabase = scimmongo.DB(resourceType, collection, scimmongo.Options().IgnoreProjection())
			ctx.logInitialized("mongo user database")
		}
	}
	return ctx.userDatabase
}

func (ctx *applicationContext) GroupDatabase() db.DB {
	if ctx.groupDatabase == nil {
		if ctx.args.UseMemoryDB {
			ctx.groupDatabase = db.Memory()
			ctx.logInitialized("in-memory group database")
		} else {
			ctx.ensureMongoMetadata()
			resourceType := ctx.GroupResourceType()
			collection := ctx.MongoClient().
				Database(ctx.args.MongoDB.Database, options.Database()).
				Collection(resourceType.Name(), options.Collection())
			ctx.groupDatabase = scimmongo.DB(resourceType, collection, scimmongo.Options().IgnoreProjection())
			ctx.logInitialized("mongo group database")
		}
	}
	return ctx.groupDatabase
}

func (ctx *applicationContext) ensureMongoMetadata() {
	ctx.registerMongoMetadataOnce.Do(func() {
		if err := ctx.args.MongoDB.RegisterMetadata(); err != nil {
			ctx.logInitFailure("mongo metadata", err)
			panic(err)
		}
		ctx.logInitialized("mongo metadata")
	})
}

func (ctx *applicationContext) RabbitMQConnection() *amqp.Connection {
	if ctx.rabbitMqConn == nil {
		connectCtx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelFunc()

		c, err := ctx.args.RabbitMQ.Connect(connectCtx)
		if err != nil {
			ctx.logInitFailure("rabbit connection", err)
			panic(err)
		}
		ctx.rabbitMqConn = c
		ctx.logInitialized("rabbit connection")
	}
	return ctx.rabbitMqConn
}

func (ctx *applicationContext) RabbitMQChannel() *amqp.Channel {
	if ctx.rabbitMqChannel == nil {
		c, err := ctx.RabbitMQConnection().Channel()
		if err != nil {
			ctx.logInitFailure("rabbit channel", err)
			panic(err)
		}
		if err := gs.DeclareQueue(c); err != nil {
			ctx.logInitFailure("rabbit channel", err)
			panic(err)
		}
		ctx.rabbitMqChannel = c
		ctx.logInitialized("rabbit channel")
	}
	return ctx.rabbitMqChannel
}

func (ctx *applicationContext) UserSyncService() *groupsync.SyncService {
	if ctx.userSyncService == nil {
		ctx.userSyncService = groupsync.NewSyncService(ctx.GroupDatabase())
		ctx.logInitialized("user sync service")
	}
	return ctx.userSyncService
}

func (ctx *applicationContext) MessageConsumer() *consumer {
	if ctx.messageConsumer == nil {
		ctx.messageConsumer = &consumer{
			rabbitCh:        ctx.RabbitMQChannel(),
			userDatabase:    ctx.UserDatabase(),
			groupDatabase:   ctx.GroupDatabase(),
			userSyncService: ctx.UserSyncService(),
			metaFilter:      filter.MetaFilter(),
			logger:          ctx.Logger(),
			trialLimit:      ctx.args.requeueLimit,
		}
		ctx.logInitialized("message consumer")
	}
	return ctx.messageConsumer
}

func (ctx *applicationContext) Close() {
	if ctx.mongoClient != nil {
		_ = ctx.mongoClient.Disconnect(context.Background())
	}
	if ctx.rabbitMqChannel != nil {
		_ = ctx.rabbitMqChannel.Close()
	}
}

func (ctx *applicationContext) logInitialized(resourceName string) {
	ctx.Logger().
		Info().
		Fields(map[string]interface{}{
			"component": resourceName,
			"status":    "initialized",
		}).
		Msg("component initialized")
}

func (ctx *applicationContext) logInitFailure(resourceName string, err error) {
	ctx.Logger().
		Fatal().
		Err(err).
		Fields(map[string]interface{}{
			"component": resourceName,
			"status":    "initialization_failed",
		}).
		Msg("component failed to initialize")
}
