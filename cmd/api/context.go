package api

import (
	"context"
	scimmongo "github.com/imulab/go-scim/v2/mongo"
	"github.com/imulab/go-scim/v2/pkg/db"
	"github.com/imulab/go-scim/v2/pkg/service"
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

	userCreateService   service.Create
	groupCreateService  service.Create
	userReplaceService  service.Replace
	groupReplaceService service.Replace
	userPatchService    service.Patch
	groupPatchService   service.Patch
	userDeleteService   service.Delete
	groupDeleteService  service.Delete
	userGetService      service.Get
	groupGetService     service.Get
	userQueryService    service.Query
	groupQueryService   service.Query
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

func (ctx *applicationContext) UserCreateService() service.Create {
	if ctx.userCreateService == nil {
		ctx.userCreateService = service.CreateService(ctx.UserResourceType(), ctx.UserDatabase(), []filter.ByResource{
			filter.ByPropertyToByResource(
				filter.ReadOnlyFilter(),
				filter.UUIDFilter(),
				filter.BCryptFilter(),
			),
			filter.MetaFilter(),
			filter.ByPropertyToByResource(filter.ValidationFilter(ctx.UserDatabase())),
		})
	}
	return ctx.userCreateService
}

func (ctx *applicationContext) GroupCreateService() service.Create {
	if ctx.groupCreateService == nil {
		ctx.groupCreateService = &groupCreated{
			service: service.CreateService(ctx.GroupResourceType(), ctx.GroupDatabase(), []filter.ByResource{
				filter.ByPropertyToByResource(
					filter.ReadOnlyFilter(),
					filter.UUIDFilter(),
				),
				filter.MetaFilter(),
				filter.ByPropertyToByResource(filter.ValidationFilter(ctx.GroupDatabase())),
			}),
			sender: &groupSyncSender{
				channel: ctx.RabbitMQChannel(),
				logger:  ctx.Logger(),
			},
		}
	}
	return ctx.groupCreateService
}

func (ctx *applicationContext) UserReplaceService() service.Replace {
	if ctx.userReplaceService == nil {
		ctx.userReplaceService = service.ReplaceService(ctx.ServiceProviderConfig(), ctx.UserResourceType(), ctx.UserDatabase(), []filter.ByResource{
			filter.ByPropertyToByResource(
				filter.ReadOnlyFilter(),
				filter.BCryptFilter(),
			),
			filter.ByPropertyToByResource(filter.ValidationFilter(ctx.UserDatabase())),
			filter.MetaFilter(),
		})
	}
	return ctx.userReplaceService
}

func (ctx *applicationContext) GroupReplaceService() service.Replace {
	if ctx.groupReplaceService == nil {
		ctx.groupReplaceService = &groupReplaced{
			service: service.ReplaceService(ctx.ServiceProviderConfig(), ctx.GroupResourceType(), ctx.GroupDatabase(), []filter.ByResource{
				filter.ByPropertyToByResource(
					filter.ReadOnlyFilter(),
				),
				filter.ByPropertyToByResource(filter.ValidationFilter(ctx.UserDatabase())),
				filter.MetaFilter(),
			}),
			sender: &groupSyncSender{
				channel: ctx.RabbitMQChannel(),
				logger:  ctx.Logger(),
			},
		}
	}
	return ctx.groupReplaceService
}

func (ctx *applicationContext) UserPatchService() service.Patch {
	if ctx.userPatchService == nil {
		ctx.userPatchService = service.PatchService(ctx.ServiceProviderConfig(), ctx.UserDatabase(), []filter.ByResource{}, []filter.ByResource{
			filter.ByPropertyToByResource(
				filter.ReadOnlyFilter(),
				filter.BCryptFilter(),
			),
			filter.ByPropertyToByResource(filter.ValidationFilter(ctx.UserDatabase())),
			filter.MetaFilter(),
		})
	}
	return ctx.userPatchService
}

func (ctx *applicationContext) GroupPatchService() service.Patch {
	if ctx.groupPatchService == nil {
		ctx.groupPatchService = &groupPatched{
			service: service.PatchService(ctx.ServiceProviderConfig(), ctx.GroupDatabase(), []filter.ByResource{}, []filter.ByResource{
				filter.ByPropertyToByResource(
					filter.ReadOnlyFilter(),
				),
				filter.ByPropertyToByResource(filter.ValidationFilter(ctx.GroupDatabase())),
				filter.MetaFilter(),
			}),
			sender: &groupSyncSender{
				channel: ctx.RabbitMQChannel(),
				logger:  ctx.Logger(),
			},
		}
	}
	return ctx.groupPatchService
}

func (ctx *applicationContext) UserDeleteService() service.Delete {
	if ctx.userDeleteService == nil {
		ctx.userDeleteService = service.DeleteService(ctx.ServiceProviderConfig(), ctx.UserDatabase())
	}
	return ctx.userDeleteService
}

func (ctx *applicationContext) GroupDeleteService() service.Delete {
	if ctx.groupDeleteService == nil {
		ctx.groupDeleteService = &groupDeleted{
			service: service.DeleteService(ctx.ServiceProviderConfig(), ctx.GroupDatabase()),
			sender: &groupSyncSender{
				channel: ctx.RabbitMQChannel(),
				logger:  ctx.Logger(),
			},
		}
	}
	return ctx.groupDeleteService
}

func (ctx *applicationContext) UserGetService() service.Get {
	if ctx.userGetService == nil {
		ctx.userGetService = service.GetService(ctx.UserDatabase())
	}
	return ctx.userGetService
}

func (ctx *applicationContext) GroupGetService() service.Get {
	if ctx.groupGetService == nil {
		ctx.groupGetService = service.GetService(ctx.GroupDatabase())
	}
	return ctx.groupGetService
}

func (ctx *applicationContext) UserQueryService() service.Query {
	if ctx.userQueryService == nil {
		ctx.userQueryService = service.QueryService(ctx.ServiceProviderConfig(), ctx.UserDatabase())
	}
	return ctx.userQueryService
}

func (ctx *applicationContext) GroupQueryService() service.Query {
	if ctx.groupQueryService == nil {
		ctx.groupQueryService = service.QueryService(ctx.ServiceProviderConfig(), ctx.GroupDatabase())
	}
	return ctx.groupQueryService
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

func (ctx *applicationContext) Close() {
	if ctx.mongoClient != nil {
		_ = ctx.mongoClient.Disconnect(context.Background())
	}
	if ctx.rabbitMqChannel != nil {
		_ = ctx.rabbitMqChannel.Close()
	}
}
