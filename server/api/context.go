package api

import (
	"encoding/json"
	"github.com/imulab/go-scim/core/expr"
	"github.com/imulab/go-scim/core/spec"
	"github.com/imulab/go-scim/protocol/db"
	"github.com/imulab/go-scim/protocol/handler"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/imulab/go-scim/protocol/services"
	"github.com/imulab/go-scim/protocol/services/filter"
	"github.com/imulab/go-scim/server/groupsync"
	"github.com/imulab/go-scim/server/logger"
	"github.com/nats-io/nats.go"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type appContext struct {
	serviceProviderConfig        *spec.ServiceProviderConfig
	serviceProviderConfigHandler *handler.ServiceProviderConfig
	natConn                      *nats.Conn
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

func (c *appContext) initialize(args *args) error {
	if err := c.loadLogger(); err != nil {
		return err
	}
	if err := c.loadServiceProviderConfig(args.serviceProviderConfigPath, c.serviceProviderConfig); err != nil {
		return err
	}
	if err := c.loadSchemasInFolder(args.schemasFolderPath); err != nil {
		return err
	}
	if err := c.loadNatsConnection(args); err != nil {
		return err
	}

	// user related
	if err := c.loadResourceType(args.userResourceTypePath, c.userResourceType); err != nil {
		return err
	}
	if err := c.loadUserDatabase(args); err != nil {
		return err
	}
	if err := c.loadUserServices(); err != nil {
		return err
	}
	if err := c.loadUserHandlers(); err != nil {
		return err
	}

	// group related
	if err := c.loadResourceType(args.groupResourceTypePath, c.groupResourceType); err != nil {
		return err
	}
	if err := c.loadGroupDatabase(args); err != nil {
		return err
	}
	if err := c.loadGroupServices(); err != nil {
		return err
	}
	if err := c.loadGroupHandlers(); err != nil {
		return err
	}

	return nil
}

func (c *appContext) loadUserHandlers() error {
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
	return nil
}

func (c *appContext) loadGroupHandlers() error {
	c.groupCreateHandler = &handler.Create{
		Log:          c.logger,
		Service:      c.groupCreateService,
		ResourceType: c.userResourceType,
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
	return nil
}

func (c *appContext) loadUserServices() error {
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
	return nil
}

func (c *appContext) loadGroupServices() error {
	syncSender, _, err := groupsync.Sender(c.natConn, c.logger)
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
		Event: syncSender,
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
		Event: syncSender,
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
		Event: syncSender,
	}
	c.groupDeleteService = &services.DeleteService{
		Logger:                c.logger,
		Database:              c.groupDatabase,
		ServiceProviderConfig: c.serviceProviderConfig,
		Event: syncSender,
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

func (c *appContext) loadUserDatabase(args *args) error {
	c.userDatabase = db.Memory()
	return nil
}

func (c *appContext) loadGroupDatabase(args *args) error {
	c.groupDatabase = db.Memory()
	return nil
}

func (c *appContext) loadResourceType(path string, dest *spec.ResourceType) error {
	raw, err := c.readFile(path)
	if err != nil {
		return err
	}
	dest = new(spec.ResourceType)
	err = json.Unmarshal(raw, dest)
	if err != nil {
		return err
	}
	expr.Register(dest)
	return nil
}

func (c *appContext) loadSchemasInFolder(folder string) error {
	return filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".json") {
			raw, err := c.readFile(path)
			if err != nil {
				return err
			}
			schema := new(spec.Schema)
			err = json.Unmarshal(raw, schema)
			if err != nil {
				return err
			}
			spec.SchemaHub.Put(schema)
		}
		return nil
	})
}

func (c *appContext) loadLogger() error {
	c.logger = logger.Zero()
	return nil
}

func (c *appContext) loadServiceProviderConfig(path string, dest *spec.ServiceProviderConfig) error {
	raw, err := c.readFile(path)
	if err != nil {
		return err
	}
	dest = new(spec.ServiceProviderConfig)
	if err = json.Unmarshal(raw, dest); err != nil {
		return err
	}
	c.serviceProviderConfigHandler = &handler.ServiceProviderConfig{
		Log: c.logger,
		SPC: dest,
	}
	return nil
}

func (c *appContext) loadNatsConnection(args *args) (err error) {
	c.natConn, err = nats.Connect(args.natsServers, nats.Timeout(10*time.Second), nats.PingInterval(10*time.Second))
	return
}

func (c *appContext) readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	raw, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
