package api

import (
	"github.com/imulab/go-scim/core/expr"
	"github.com/imulab/go-scim/core/spec"
	scimmongo "github.com/imulab/go-scim/mongo"
	"github.com/imulab/go-scim/protocol/db"
	"github.com/imulab/go-scim/protocol/event"
	"github.com/imulab/go-scim/protocol/handler"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/imulab/go-scim/protocol/services"
	"github.com/imulab/go-scim/protocol/services/filter"
	"github.com/imulab/go-scim/server/groupsync"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type appContext struct {
	serviceProviderConfig        *spec.ServiceProviderConfig
	serviceProviderConfigHandler *handler.ServiceProviderConfig
	rabbitChannel                *amqp.Channel
	mongoClient                  *mongo.Client
	logger                       log.Logger

	// user related
	userResourceType   *spec.ResourceType
	userDatabase       db.DB
	userCreateService  *services.CreateService
	userReplaceService *services.ReplaceService
	userPatchService   *services.PatchService
	userDeleteService  *services.DeleteService
	userGetService     *services.GetService
	userQueryService   *services.QueryService
	userCreateHandler  *handler.Create
	userReplaceHandler *handler.Replace
	userPatchHandler   *handler.Patch
	userDeleteHandler  *handler.Delete
	userGetHandler     *handler.Get
	userQueryHandler   *handler.Query

	// group related
	groupResourceType   *spec.ResourceType
	groupDatabase       db.DB
	groupCreateService  *services.CreateService
	groupReplaceService *services.ReplaceService
	groupPatchService   *services.PatchService
	groupDeleteService  *services.DeleteService
	groupGetService     *services.GetService
	groupQueryService   *services.QueryService
	groupCreateHandler  *handler.Create
	groupReplaceHandler *handler.Replace
	groupPatchHandler   *handler.Patch
	groupDeleteHandler  *handler.Delete
	groupGetHandler     *handler.Get
	groupQueryHandler   *handler.Query

	// todo root related: bulk, root query
}

func (c *appContext) initialize(args *arguments) (err error) {
	// logger
	c.logger = args.Logger()

	// service provider config
	if c.serviceProviderConfig, err = args.ParseServiceProviderConfig(); err != nil {
		return
	} else {
		c.serviceProviderConfigHandler = &handler.ServiceProviderConfig{
			Log: c.logger,
			SPC: c.serviceProviderConfig,
		}
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

	// user related
	c.loadUserDatabase(args)
	c.loadUserServices()
	c.loadUserHandlers()

	// group related
	c.loadGroupDatabase(args)
	if err = c.loadGroupServices(); err != nil {
		return
	}
	c.loadGroupHandlers()

	return
}

func (c *appContext) loadUserHandlers() {
	c.userCreateHandler = &handler.Create{
		Log:          c.logger,
		Service:      c.userCreateService,
		ResourceType: c.userResourceType,
	}
	c.userReplaceHandler = &handler.Replace{
		Log:                 c.logger,
		Service:             c.userReplaceService,
		ResourceIDPathParam: "id",
		ResourceType:        c.userResourceType,
	}
	c.userPatchHandler = &handler.Patch{
		Log:                 c.logger,
		Service:             c.userPatchService,
		ResourceIDPathParam: "id",
	}
	c.userDeleteHandler = &handler.Delete{
		Log:                 c.logger,
		Service:             c.userDeleteService,
		ResourceIDPathParam: "id",
	}
	c.userGetHandler = &handler.Get{
		Log:                 c.logger,
		Service:             c.userGetService,
		ResourceIDPathParam: "id",
	}
	c.userQueryHandler = &handler.Query{
		Log:     c.logger,
		Service: c.userQueryService,
	}
}

func (c *appContext) loadGroupHandlers() {
	c.groupCreateHandler = &handler.Create{
		Log:          c.logger,
		Service:      c.groupCreateService,
		ResourceType: c.groupResourceType,
	}
	c.groupReplaceHandler = &handler.Replace{
		Log:                 c.logger,
		Service:             c.groupReplaceService,
		ResourceIDPathParam: "id",
		ResourceType:        c.groupResourceType,
	}
	c.groupPatchHandler = &handler.Patch{
		Log:                 c.logger,
		Service:             c.groupPatchService,
		ResourceIDPathParam: "id",
	}
	c.groupDeleteHandler = &handler.Delete{
		Log:                 c.logger,
		Service:             c.groupDeleteService,
		ResourceIDPathParam: "id",
	}
	c.groupGetHandler = &handler.Get{
		Log:                 c.logger,
		Service:             c.groupGetService,
		ResourceIDPathParam: "id",
	}
	c.groupQueryHandler = &handler.Query{
		Log:     c.logger,
		Service: c.groupQueryService,
	}
}

func (c *appContext) loadUserServices() {
	c.userCreateService = &services.CreateService{
		Logger:   c.logger,
		Database: c.userDatabase,
		Filters: []filter.ForResource{
			filter.ClearReadOnly(),
			filter.ID(),
			filter.Password(10),
			filter.Meta(),
			filter.Validation(c.userDatabase),
		},
	}
	c.userReplaceService = &services.ReplaceService{
		Logger:                c.logger,
		Database:              c.userDatabase,
		ServiceProviderConfig: c.serviceProviderConfig,
		Filters: []filter.ForResource{
			filter.ClearReadOnly(),
			filter.CopyReadOnly(),
			filter.Password(10),
			filter.Validation(c.userDatabase),
			filter.Meta(),
		},
	}
	c.userPatchService = &services.PatchService{
		Logger:                c.logger,
		Database:              c.userDatabase,
		ServiceProviderConfig: c.serviceProviderConfig,
		PrePatchFilters:       []filter.ForResource{},
		PostPatchFilters: []filter.ForResource{
			filter.CopyReadOnly(),
			filter.Password(10),
			filter.Validation(c.userDatabase),
			filter.Meta(),
		},
	}
	c.userDeleteService = &services.DeleteService{
		Logger:                c.logger,
		Database:              c.userDatabase,
		ServiceProviderConfig: c.serviceProviderConfig,
	}
	c.userGetService = &services.GetService{
		Logger:   c.logger,
		Database: c.userDatabase,
	}
	c.userQueryService = &services.QueryService{
		Logger:                c.logger,
		Database:              c.userDatabase,
		ServiceProviderConfig: c.serviceProviderConfig,
	}
}

func (c *appContext) loadGroupServices() error {
	syncSender, err := groupsync.Sender(c.rabbitChannel, c.logger)
	if err != nil {
		return err
	}
	c.groupCreateService = &services.CreateService{
		Logger:   c.logger,
		Database: c.groupDatabase,
		Filters: []filter.ForResource{
			filter.ClearReadOnly(),
			filter.ID(),
			filter.Meta(),
			filter.Validation(c.groupDatabase),
		},
		Event: event.Of(syncSender),
	}
	c.groupReplaceService = &services.ReplaceService{
		Logger:                c.logger,
		Database:              c.groupDatabase,
		ServiceProviderConfig: c.serviceProviderConfig,
		Filters: []filter.ForResource{
			filter.ClearReadOnly(),
			filter.CopyReadOnly(),
			filter.Validation(c.groupDatabase),
			filter.Meta(),
		},
		Event: event.Of(syncSender),
	}
	c.groupPatchService = &services.PatchService{
		Logger:                c.logger,
		Database:              c.groupDatabase,
		ServiceProviderConfig: c.serviceProviderConfig,
		PrePatchFilters:       []filter.ForResource{},
		PostPatchFilters: []filter.ForResource{
			filter.CopyReadOnly(),
			filter.Validation(c.groupDatabase),
			filter.Meta(),
		},
		Event: event.Of(syncSender),
	}
	c.groupDeleteService = &services.DeleteService{
		Logger:                c.logger,
		Database:              c.groupDatabase,
		ServiceProviderConfig: c.serviceProviderConfig,
		Event:                 syncSender,
	}
	c.groupGetService = &services.GetService{
		Logger:   c.logger,
		Database: c.groupDatabase,
	}
	c.groupQueryService = &services.QueryService{
		Logger:                c.logger,
		Database:              c.groupDatabase,
		ServiceProviderConfig: c.serviceProviderConfig,
	}
	return nil
}

func (c *appContext) loadUserDatabase(args *arguments) {
	if args.MemoryDB {
		c.userDatabase = db.Memory()
	} else {
		coll := c.mongoClient.Database(args.Mongo.Database, options.Database()).Collection(c.userResourceType.Name(), options.Collection())
		c.userDatabase = scimmongo.DB(c.userResourceType, c.logger, coll, scimmongo.Options())
	}
}

func (c *appContext) loadGroupDatabase(args *arguments) {
	if args.MemoryDB {
		c.groupDatabase = db.Memory()
	} else {
		coll := c.mongoClient.Database(args.Mongo.Database, options.Database()).Collection(c.groupResourceType.Name(), options.Collection())
		c.groupDatabase = scimmongo.DB(c.groupResourceType, c.logger, coll, scimmongo.Options())
	}
}
