package api

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/cli/v2"
)

// Command returns a cli.Command that starts an HTTP router to serve the SCIM API.
func Command() *cli.Command {
	args := newArgs()
	return &cli.Command{
		Name:        "api",
		Usage:       "Start SCIM API server with optional authentication",
		Description: "Manage state of resources defined in the SCIM protocol. Supports OAuth2 and Bearer token authentication.",
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

			// Configure handler with optional authentication
			var handler http.Handler = router

			// Apply authentication middleware if enabled
			if args.Auth.IsAuthenticationEnabled() {
				handler = AuthMiddleware(args.Auth)(router)

				app.Logger().Info().Fields(map[string]interface{}{
					"port":           args.httpPort,
					"auth_enabled":   true,
					"oauth2_enabled": args.Auth.OAuth2Enabled,
					"bearer_tokens":  len(args.Auth.GetBearerTokens()) > 0,
				}).Msg("Listening for incoming requests with authentication enabled.")
			} else {
				app.Logger().Info().Fields(map[string]interface{}{
					"port":         args.httpPort,
					"auth_enabled": false,
				}).Msg("Listening for incoming requests without authentication.")
			}

			return http.ListenAndServe(fmt.Sprintf(":%d", args.httpPort), handler)
		},
	}
}
