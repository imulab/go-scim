package main

import (
	"fmt"
	"net/http"
	"github.com/go-zoo/bone"
	scim "github.com/davidiamyou/go-scim/shared"
	web "github.com/davidiamyou/go-scim/handlers"
	"github.com/davidiamyou/go-scim/mongo"
	"context"
)

// setup everything
func init() {
	propertySource := &mapPropertySource{
		data: map[string]interface{}{
			// TODO
		},
	}

	// TODO load all schemas
	s, _, err := scim.ParseSchema("root")
	web.ErrorCheck(err)
	rootSchemaInternal = s
	s, _, err = scim.ParseSchema("")
	web.ErrorCheck(err)
	userSchemaInternal = s
	s, _, err = scim.ParseSchema("")
	web.ErrorCheck(err)
	groupSchemaInternal = s
	s, _, err = scim.ParseSchema("")
	web.ErrorCheck(err)
	userSchema = s
	s, _, err = scim.ParseSchema("")
	web.ErrorCheck(err)
	groupSchema = s

	// TODO create all repositories
	resourceConstructor := func(c scim.Complex) scim.DataProvider { return &scim.Resource{Complex: c} }
	userRepo, err = mongo.NewMongoRepository(nil, "", "", userSchemaInternal, resourceConstructor)
	web.ErrorCheck(err)
	groupRepo, err = mongo.NewMongoRepository(nil, "", "", groupSchemaInternal, resourceConstructor)
	web.ErrorCheck(err)

	exampleServer = &simpleServer{
		logger:&printLogger{},
		propertySource:propertySource,
		idAssignment:scim.NewIdAssignment(),
		userMetaAssignment:scim.NewMetaAssignment(propertySource, scim.UserResourceType),
		groupMetaAssignment:scim.NewMetaAssignment(propertySource, scim.GroupResourceType),
		groupAssignment:scim.NewGroupAssignment(groupRepo),
	}
}

func main() {
	// TODO
}

// Resource schemas
var (
	rootSchemaInternal,
	userSchemaInternal,
	groupSchemaInternal,
	userSchema,
	groupSchema	*scim.Schema
)

// Repositories
var (
	userRepo,
	groupRepo,
	resourceTypeRepo,
	spConfigRepo 	scim.Repository
)

var exampleServer web.ScimServer

// Example server implementation
type simpleServer struct {
	propertySource	 	*mapPropertySource
	logger 			*printLogger
	idAssignment		scim.ReadOnlyAssignment
	userMetaAssignment 	scim.ReadOnlyAssignment
	groupMetaAssignment 	scim.ReadOnlyAssignment
	groupAssignment 	scim.ReadOnlyAssignment
}
func (ss *simpleServer) Property() scim.PropertySource { return ss.propertySource }
func (ss *simpleServer) Logger() scim.Logger { return ss.logger }
func (ss *simpleServer) UrlParam(name string, req *http.Request) string { return extractUrlParameter(name, req) }
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
		return userSchema
	case scim.GroupUrn:
		return groupSchema
	default:
		panic(scim.Error.Text("unknown schema id %s", id))
	}
}
func (ss *simpleServer) ValidateType(subj *scim.Resource, sch *scim.Schema) error { return scim.ValidateType(subj, sch) }
func (ss *simpleServer) ValidateRequired(subj *scim.Resource, sch *scim.Schema) error { return scim.ValidateRequired(subj, sch) }
func (ss *simpleServer) ValidateMutability(subj *scim.Resource, ref *scim.Resource, sch *scim.Schema) error { return scim.ValidateMutability(subj, ref, sch) }
func (ss *simpleServer) ValidateUniqueness(subj *scim.Resource, sch *scim.Schema, repo scim.Repository) error { return scim.ValidateUniqueness(subj, sch, repo) }
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
func (ss *simpleServer) MarshalJSON(v interface{}, sch *scim.Schema, attributes []string, excludedAttributes []string) ([]byte, error) { return scim.MarshalJSON(v, sch, attributes, excludedAttributes) }
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
type mapPropertySource struct { data map[string]interface{} }
func (mps *mapPropertySource) Get(key string) interface{} { return mps.data[key] }
func (mps *mapPropertySource) GetString(key string) string { return mps.Get(key).(string) }
func (mps *mapPropertySource) GetInt(key string) int  { return mps.Get(key).(int) }
func (mps *mapPropertySource) GetBool(key string) bool  { return mps.Get(key).(bool) }

// simple logger that writes to terminal
type printLogger struct {}
func (pl *printLogger) Info(template string, args ...interface{}) { fmt.Println("[INFO] " + template, args) }
func (pl *printLogger) Debug(template string, args ...interface{}) { fmt.Println("[DEBUG] " + template, args) }
func (pl *printLogger) Error(template string, args ...interface{}) { fmt.Println("[ERROR] " + template, args) }