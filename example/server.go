package main

import (
	"context"
	"fmt"
	web "github.com/davidiamyou/go-scim/handlers"
	"github.com/davidiamyou/go-scim/mongo"
	scim "github.com/davidiamyou/go-scim/shared"
	"github.com/go-zoo/bone"
	"net/http"
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
			"mongo.url":                                "mongodb://localhost:32794/scim_example?maxPoolSize=100",
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

func (ss *simpleServer) Property() scim.PropertySource { return ss.propertySource }
func (ss *simpleServer) Logger() scim.Logger           { return ss.logger }
func (ss *simpleServer) UrlParam(name string, req *http.Request) string {
	return extractUrlParameter(name, req)
}
func (ss *simpleServer) Schema(id string) *scim.Schema {
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
func (ss *simpleServer) InternalSchema(id string) *scim.Schema {
	switch id {
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

// url parameter extractor
func extractUrlParameter(name string, req *http.Request) string { return bone.GetValue(req, name) }

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
