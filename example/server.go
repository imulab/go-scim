package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-zoo/bone"
	web "github.com/parsable/go-scim/handlers"
	"github.com/parsable/go-scim/mongo"
	scim "github.com/parsable/go-scim/shared"
)

// setup everything
func initConfiguration() {
	propertySource := &mapPropertySource{
		data: map[string]interface{}{
			"scim.resources.user.locationBase":         "http://localhost:8080/v2/Users",
			"scim.resources.group.locationBase":        "http://localhost:8080/v2/Groups",
			"scim.resources.schema.internalRoot.path":  "../resources/schemas/root_internal.json",
			"scim.resources.schema.internalUser.path":  "../resources/schemas/user_internal.json",
			"scim.resources.schema.internalGroup.path": "../resources/schemas/group_internal.json",
			"scim.resources.schema.user.path":          "../resources/schemas/user.json",
			"scim.resources.schema.group.path":         "../resources/schemas/group.json",
			"scim.resources.resourceType.user":         "../resources/resource_types/user.json",
			"scim.resources.resourceType.group":        "../resources/resource_types/group.json",
			"scim.resources.spConfig":                  "../resources/sp_config/sp_config.json",
			"scim.protocol.itemsPerPage":               10,
			"scim.protocol.uri.user":                   "/Users",
			"scim.protocol.uri.group":                  "/Groups",
			"mongo.url":                                "mongodb://localhost:32768/scim_example?maxPoolSize=100",
			"mongo.db":                                 "scim_example",
			"mongo.collection.user":                    "users",
			"mongo.collection.group":                   "groups",
		},
	}

	s, _, err := scim.ParseSchema(propertySource.GetString("scim.resources.schema.internalRoot.path"))
	web.ErrorCheck(err)
	rootSchemaInternal = s
	s, _, err = scim.ParseSchema(propertySource.GetString("scim.resources.schema.internalUser.path"))
	web.ErrorCheck(err)
	userSchemaInternal = s
	s, _, err = scim.ParseSchema(propertySource.GetString("scim.resources.schema.internalGroup.path"))
	web.ErrorCheck(err)
	groupSchemaInternal = s
	s, _, err = scim.ParseSchema(propertySource.GetString("scim.resources.schema.user.path"))
	web.ErrorCheck(err)
	userSchema = s
	s, _, err = scim.ParseSchema(propertySource.GetString("scim.resources.schema.group.path"))
	web.ErrorCheck(err)
	groupSchema = s

	userResourceType, _, err := scim.ParseResource(propertySource.GetString("scim.resources.resourceType.user"))
	web.ErrorCheck(err)
	groupResourceType, _, err := scim.ParseResource(propertySource.GetString("scim.resources.resourceType.group"))
	web.ErrorCheck(err)

	spConfig, _, err := scim.ParseResource(propertySource.GetString("scim.resources.spConfig"))
	web.ErrorCheck(err)

	resourceConstructor := func(c scim.Complex) scim.DataProvider { return &scim.Resource{Complex: c} }
	userRepo, err = mongo.NewMongoRepositoryWithUrl(
		propertySource.GetString("mongo.url"),
		propertySource.GetString("mongo.db"),
		propertySource.GetString("mongo.collection.user"),
		userSchemaInternal,
		resourceConstructor)
	web.ErrorCheck(err)
	groupRepo, err = mongo.NewMongoRepositoryWithUrl(
		propertySource.GetString("mongo.url"),
		propertySource.GetString("mongo.db"),
		propertySource.GetString("mongo.collection.group"),
		groupSchemaInternal,
		resourceConstructor)
	web.ErrorCheck(err)
	rootQueryRepo = &mongoRootQueryRepository{
		repos: []scim.Repository{
			userRepo,
			groupRepo,
		},
	}
	resourceTypeRepo = scim.NewMapRepository(map[string]scim.DataProvider{
		userResourceType.GetId():  userResourceType,
		groupResourceType.GetId(): groupResourceType,
	})
	spConfigRepo = scim.NewMapRepository(map[string]scim.DataProvider{
		"": spConfig,
	})

	exampleServer = &simpleServer{
		logger:              &printLogger{},
		propertySource:      propertySource,
		idAssignment:        scim.NewIdAssignment(),
		userMetaAssignment:  scim.NewMetaAssignment(propertySource, scim.UserResourceType),
		groupMetaAssignment: scim.NewMetaAssignment(propertySource, scim.GroupResourceType),
		groupAssignment:     scim.NewGroupAssignment(groupRepo),
	}
}

func main() {
	initConfiguration()
	wrap := func(handler web.EndpointHandler, requestType int) http.HandlerFunc {
		return web.Endpoint(web.InjectRequestScope(web.ErrorRecovery(handler), requestType), exampleServer)
	}

	mux := bone.New()
	mux.Prefix("/v2")

	mux.GetFunc("/Users/:resourceId", wrap(web.GetUserByIdHandler, scim.GetUserById))
	mux.PostFunc("/Users", wrap(web.CreateUserHandler, scim.CreateUser))
	mux.DeleteFunc("/Users/:resourceId", wrap(web.DeleteUserByIdHandler, scim.DeleteUser))
	mux.GetFunc("/Users", wrap(web.QueryUserHandler, scim.QueryUser))
	mux.PostFunc("/Users/.search", wrap(web.QueryUserHandler, scim.QueryUser))
	mux.PutFunc("/Users/:resourceId", wrap(web.ReplaceUserHandler, scim.ReplaceUser))
	mux.PatchFunc("/Users/:resourceId", wrap(web.PatchUserHandler, scim.PatchUser))

	mux.GetFunc("/Groups/:resourceId", wrap(web.GetGroupByIdHandler, scim.GetGroupById))
	mux.PostFunc("/Groups", wrap(web.CreateGroupHandler, scim.CreateGroup))
	mux.DeleteFunc("/Groups/:resourceId", wrap(web.DeleteGroupByIdHandler, scim.DeleteGroup))
	mux.GetFunc("/Groups", wrap(web.QueryGroupHandler, scim.QueryGroup))
	mux.PostFunc("/Groups/.search", wrap(web.QueryGroupHandler, scim.QueryGroup))
	mux.PutFunc("/Groups/:resourceId", wrap(web.ReplaceGroupHandler, scim.ReplaceGroup))
	mux.PatchFunc("/Groups/:resourceId", wrap(web.PatchGroupHandler, scim.PatchGroup))

	mux.PostFunc("/Bulk", wrap(web.BulkHandler, scim.BulkOp))

	mux.GetFunc("/", wrap(web.RootQueryHandler, scim.RootQuery))
	mux.PostFunc("/.search", wrap(web.RootQueryHandler, scim.RootQuery))

	mux.GetFunc("/Schemas/:resourceId", wrap(web.GetSchemaByIdHandler, scim.GetSchemaById))
	mux.GetFunc("/Schemas", wrap(web.GetAllSchemaHandler, scim.GetAllSchema))

	mux.GetFunc("/ResourceTypes", wrap(web.GetAllResourceTypeHandler, scim.GetAllResourceType))

	mux.GetFunc("/ServiceProviderConfig", wrap(web.GetServiceProviderConfigHandler, scim.GetSPConfig))

	http.ListenAndServe(":8080", mux)
}

// Resource schemas
var (
	rootSchemaInternal,
	userSchemaInternal,
	groupSchemaInternal,
	userSchema,
	groupSchema *scim.Schema
)

// Repositories
var (
	userRepo,
	groupRepo,
	rootQueryRepo,
	resourceTypeRepo,
	spConfigRepo scim.Repository
)

var exampleServer web.ScimServer

// Example server implementation
type simpleServer struct {
	propertySource      *mapPropertySource
	logger              *printLogger
	idAssignment        scim.ReadOnlyAssignment
	userMetaAssignment  scim.ReadOnlyAssignment
	groupMetaAssignment scim.ReadOnlyAssignment
	groupAssignment     scim.ReadOnlyAssignment
}

func (ss *simpleServer) Property() scim.PropertySource              { return ss.propertySource }
func (ss *simpleServer) Logger() scim.Logger                        { return ss.logger }
func (ss *simpleServer) WebRequest(r *http.Request) scim.WebRequest { return HttpWebRequest{Req: r} }
func (ss *simpleServer) Schema(id string) *scim.Schema {
	switch id {
	case scim.UserUrn:
		return userSchema
	case scim.GroupUrn:
		return groupSchema
	default:
		panic(scim.Error.Text("unknown schema id %s", id))
	}
}
func (ss *simpleServer) InternalSchema(id string) *scim.Schema {
	switch id {
	case "":
		return rootSchemaInternal
	case scim.UserUrn:
		return userSchemaInternal
	case scim.GroupUrn:
		return groupSchemaInternal
	default:
		panic(scim.Error.Text("unknown schema id %s", id))
	}
}
func (ss *simpleServer) CorrectCase(subj *scim.Resource, sch *scim.Schema, ctx context.Context) error {
	return scim.CorrectCase(subj, sch, ctx)
}
func (ss *simpleServer) ApplyPatch(patch scim.Patch, subj *scim.Resource, sch *scim.Schema, ctx context.Context) error {
	return scim.ApplyPatch(patch, subj, sch, ctx)
}
func (ss *simpleServer) ValidateType(subj *scim.Resource, sch *scim.Schema, ctx context.Context) error {
	return scim.ValidateType(subj, sch, ctx)
}
func (ss *simpleServer) ValidateRequired(subj *scim.Resource, sch *scim.Schema, ctx context.Context) error {
	return scim.ValidateRequired(subj, sch, ctx)
}
func (ss *simpleServer) ValidateMutability(subj *scim.Resource, ref *scim.Resource, sch *scim.Schema, ctx context.Context) error {
	return scim.ValidateMutability(subj, ref, sch, ctx)
}
func (ss *simpleServer) ValidateUniqueness(subj *scim.Resource, sch *scim.Schema, repo scim.Repository, ctx context.Context) error {
	return scim.ValidateUniqueness(subj, sch, repo, ctx)
}
func (ss *simpleServer) AssignReadOnlyValue(r *scim.Resource, ctx context.Context) (err error) {
	requestType := ctx.Value(scim.RequestType{}).(int)
	switch requestType {
	case scim.CreateUser:
		err = ss.idAssignment.AssignValue(r, ctx)
		web.ErrorCheck(err)
		err = ss.userMetaAssignment.AssignValue(r, ctx)
		web.ErrorCheck(err)
		err = ss.groupAssignment.AssignValue(r, ctx)
		web.ErrorCheck(err)
	case scim.ReplaceUser, scim.PatchUser:
		err = ss.userMetaAssignment.AssignValue(r, ctx)
		web.ErrorCheck(err)
		err = ss.groupAssignment.AssignValue(r, ctx)
		web.ErrorCheck(err)
	case scim.CreateGroup:
		err = ss.idAssignment.AssignValue(r, ctx)
		web.ErrorCheck(err)
		err = ss.groupMetaAssignment.AssignValue(r, ctx)
		web.ErrorCheck(err)
	case scim.ReplaceGroup, scim.PatchGroup:
		err = ss.groupMetaAssignment.AssignValue(r, ctx)
		web.ErrorCheck(err)
	}
	return
}
func (ss *simpleServer) MarshalJSON(v interface{}, sch *scim.Schema, attributes []string, excludedAttributes []string) ([]byte, error) {
	return scim.MarshalJSON(v, sch, attributes, excludedAttributes)
}
func (ss *simpleServer) Repository(identifier string) scim.Repository {
	switch identifier {
	case "":
		return rootQueryRepo
	case scim.UserResourceType:
		return userRepo
	case scim.GroupResourceType:
		return groupRepo
	case scim.ResourceTypeResourceType:
		return resourceTypeRepo
	case scim.ServiceProviderConfigResourceType:
		return spConfigRepo
	default:
		panic(scim.Error.Text("no repo matches identifier %s", identifier))
	}
}

// simple map based property source
type mapPropertySource struct{ data map[string]interface{} }

func (mps *mapPropertySource) Get(key string) interface{}  { return mps.data[key] }
func (mps *mapPropertySource) GetString(key string) string { return mps.Get(key).(string) }
func (mps *mapPropertySource) GetInt(key string) int       { return mps.Get(key).(int) }
func (mps *mapPropertySource) GetBool(key string) bool     { return mps.Get(key).(bool) }

// simple logger that writes to terminal
type printLogger struct{}

func (pl *printLogger) Info(template string, args ...interface{}) {
	fmt.Println("[INFO] "+template, args)
}
func (pl *printLogger) Debug(template string, args ...interface{}) {
	fmt.Println("[DEBUG] "+template, args)
}
func (pl *printLogger) Error(template string, args ...interface{}) {
	fmt.Println("[ERROR] "+template, args)
}

// http request source
type HttpWebRequest struct{ Req *http.Request }

func (hwr HttpWebRequest) Target() string            { return hwr.Req.RequestURI }
func (hwr HttpWebRequest) Method() string            { return hwr.Req.Method }
func (hwr HttpWebRequest) Header(name string) string { return hwr.Req.Header.Get(name) }
func (hwr HttpWebRequest) Body() ([]byte, error)     { return ioutil.ReadAll(hwr.Req.Body) }
func (hwr HttpWebRequest) Param(name string) string {
	if v := hwr.Req.URL.Query().Get(name); len(v) > 0 {
		return v
	} else if v := bone.GetValue(hwr.Req, name); len(v) > 0 {
		return v
	} else {
		return ""
	}
}

// mongo root query repository
type mongoRootQueryRepository struct {
	repos []scim.Repository
}

func (m *mongoRootQueryRepository) Create(provider scim.DataProvider) error { panic("not implemented") }
func (m *mongoRootQueryRepository) Get(id, version string) (scim.DataProvider, error) {
	panic("not implemented")
}
func (m *mongoRootQueryRepository) GetAll() ([]scim.Complex, error) { panic("not implemented") }
func (m *mongoRootQueryRepository) Count(query string) (int, error) { panic("not implemented") }
func (m *mongoRootQueryRepository) Update(id, version string, provider scim.DataProvider) error {
	panic("not implemented")
}
func (m *mongoRootQueryRepository) Delete(id, version string) error { panic("not implemented") }
func (m *mongoRootQueryRepository) Search(payload scim.SearchRequest) (*scim.ListResponse, error) {
	return scim.CompositeSearchFunc(m.repos...)(payload)
}
