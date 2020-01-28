package groupsync

import (
	"context"
	scimmongo "github.com/imulab/go-scim/v2/mongo"
	"github.com/imulab/go-scim/v2/pkg/db"
	"github.com/imulab/go-scim/v2/pkg/groupsync"
	"github.com/imulab/go-scim/v2/pkg/service/filter"
	"github.com/imulab/go-scim/v2/pkg/spec"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

type applicationContext struct {
	args *arguments

	logger *zerolog.Logger

	serviceProviderConfig *spec.ServiceProviderConfig
	registerSchemaOnce    sync.Once
	userResourceType      *spec.ResourceType
	groupResourceType     *spec.ResourceType

	userDatabase              db.DB
	groupDatabase             db.DB
	mongoClient               *mongo.Client
	registerMongoMetadataOnce sync.Once

	rabbitMqChannel *amqp.Channel

	userSyncService *groupsync.SyncService

	messageConsumer *consumer
}

func (ctx *applicationContext) Logger() *zerolog.Logger {
	if ctx.logger == nil {
		ctx.logger = ctx.args.Logger()
	}
	return ctx.logger
}

func (ctx *applicationContext) ServiceProviderConfig() *spec.ServiceProviderConfig {
	if ctx.serviceProviderConfig == nil {
		spc, err := ctx.args.ParseServiceProviderConfig()
		if err != nil {
			panic(err)
		}
		ctx.serviceProviderConfig = spc
	}
	return ctx.serviceProviderConfig
}

func (ctx *applicationContext) UserResourceType() *spec.ResourceType {
	ctx.ensureSchemaRegistered()
	if ctx.userResourceType == nil {
		u, err := ctx.args.ParseUserResourceType()
		if err != nil {
			panic(err)
		}
		ctx.userResourceType = u
	}
	return ctx.userResourceType
}

func (ctx *applicationContext) GroupResourceType() *spec.ResourceType {
	ctx.ensureSchemaRegistered()
	if ctx.groupResourceType == nil {
		g, err := ctx.args.ParseGroupResourceType()
		if err != nil {
			panic(err)
		}
		ctx.groupResourceType = g
	}
	return ctx.groupResourceType
}

func (ctx *applicationContext) ensureSchemaRegistered() {
	ctx.registerSchemaOnce.Do(func() {
		if err := ctx.args.RegisterSchemas(); err != nil {
			panic(err)
		}
	})
}

func (ctx *applicationContext) MongoClient() *mongo.Client {
	if ctx.mongoClient == nil {
		connectCtx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelFunc()

		c, err := ctx.args.MongoDB.Connect(connectCtx)
		if err != nil {
			panic(err)
		}

		ctx.mongoClient = c
	}
	return ctx.mongoClient
}

func (ctx *applicationContext) UserDatabase() db.DB {
	if ctx.userDatabase == nil {
		if ctx.args.UseMemoryDB {
			ctx.userDatabase = db.Memory()
		} else {
			ctx.ensureMongoMetadata()
			resourceType := ctx.UserResourceType()
			collection := ctx.MongoClient().
				Database(ctx.args.MongoDB.Database, options.Database()).
				Collection(resourceType.Name(), options.Collection())
			ctx.userDatabase = scimmongo.DB(resourceType, collection, scimmongo.Options().IgnoreProjection())
		}
	}
	return ctx.userDatabase
}

func (ctx *applicationContext) GroupDatabase() db.DB {
	if ctx.groupDatabase == nil {
		if ctx.args.UseMemoryDB {
			ctx.groupDatabase = db.Memory()
		} else {
			ctx.ensureMongoMetadata()
			resourceType := ctx.GroupResourceType()
			collection := ctx.MongoClient().
				Database(ctx.args.MongoDB.Database, options.Database()).
				Collection(resourceType.Name(), options.Collection())
			ctx.groupDatabase = scimmongo.DB(resourceType, collection, scimmongo.Options().IgnoreProjection())
		}
	}
	return ctx.groupDatabase
}

func (ctx *applicationContext) ensureMongoMetadata() {
	ctx.registerMongoMetadataOnce.Do(func() {
		if err := ctx.args.MongoDB.RegisterMetadata(); err != nil {
			panic(err)
		}
	})
}

func (ctx *applicationContext) RabbitMQChannel() *amqp.Channel {
	if ctx.rabbitMqChannel == nil {
		connectCtx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelFunc()

		c, err := ctx.args.RabbitMQ.Connect(connectCtx)
		if err != nil {
			panic(err)
		}
		ctx.rabbitMqChannel = c
	}
	return ctx.rabbitMqChannel
}

func (ctx *applicationContext) UserSyncService() *groupsync.SyncService {
	if ctx.userSyncService == nil {
		ctx.userSyncService = groupsync.NewSyncService(ctx.GroupDatabase())
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
