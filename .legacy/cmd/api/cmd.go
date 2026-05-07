package api

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/cli/v2"
	"net/http"
)

// Command returns a cli.Command that starts an HTTP router to serve the SCIM API.
func Command() *cli.Command {
	args := newArgs()
	return &cli.Command{
		Name:        "api",
		Description: "Manage state of resources defined in the SCIM (Simple Cloud Identity Management) protocol",
		Flags:       args.Flags(),
		Action: func(_ *cli.Context) error {
			app := args.Initialize()
			defer app.Close()

			app.ensureSchemaRegistered()

			var router = httprouter.New()
			{
				router.GET("/ServiceProviderConfig", ServiceProviderConfigHandler(app.ServiceProviderConfig()))
				router.GET("/Schemas", SchemasHandler())
				router.GET("/Schemas/:id", SchemaByIdHandler())
				router.GET("/ResourceTypes", ResourceTypesHandler(app.UserResourceType(), app.GroupResourceType()))
				router.GET("/ResourceTypes/:id", ResourceTypeByIdHandler(app.userResourceType, app.GroupResourceType()))

				router.GET("/Users/:id", GetHandler(app.UserGetService(), app.Logger()))
				router.GET("/Users", SearchHandler(app.UserQueryService(), app.Logger()))
				router.POST("/Users", CreateHandler(app.UserCreateService(), app.Logger()))
				router.PUT("/Users/:id", ReplaceHandler(app.UserReplaceService(), app.Logger()))
				router.PATCH("/Users/:id", PatchHandler(app.UserPatchService(), app.Logger()))
				router.DELETE("/Users/:id", DeleteHandler(app.UserDeleteService(), app.Logger()))

				router.GET("/Groups/:id", GetHandler(app.GroupGetService(), app.Logger()))
				router.GET("/Groups", SearchHandler(app.GroupQueryService(), app.Logger()))
				router.POST("/Groups", CreateHandler(app.GroupCreateService(), app.Logger()))
				router.PUT("/Groups/:id", ReplaceHandler(app.GroupReplaceService(), app.Logger()))
				router.PATCH("/Groups/:id", PatchHandler(app.GroupPatchService(), app.Logger()))
				router.DELETE("/Groups/:id", DeleteHandler(app.GroupDeleteService(), app.Logger()))

				router.GET("/health", HealthHandler(app.MongoClient(), app.RabbitMQConnection()))
			}

			app.Logger().Info().Fields(map[string]interface{}{
				"port": args.httpPort,
			}).Msg("Listening for incoming requests.")

			return http.ListenAndServe(fmt.Sprintf(":%d", args.httpPort), router)
		},
	}
}
