package api

import (
	"context"
	"fmt"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/cli/v2"
	"net/http"
)

// Return a command to start a process of serving SCIM HTTP API.
func Command() *cli.Command {
	ag := new(arguments)
	return &cli.Command{
		Name:        "api",
		Usage:       "Serves API for SCIM (Simple Cloud Identity Management) protocol",
		Description: "Manage state of resources defined in the SCIM (Simple Cloud Identity Management) protocol",
		Flags: ag.Flags(),
		Action: func(cliContext *cli.Context) error {
			var appCtx *appContext
			{
				appCtx = new(appContext)
				if err := appCtx.initialize(ag); err != nil {
					return err
				}
				appCtx.logger.Info("application appContext initialized", log.Args{})
			}
			defer func() {
				_ = appCtx.rabbitChannel.Close()
				_ = appCtx.mongoClient.Disconnect(context.Background())
			}()

			var router = httprouter.New()
			{
				router.GET("/ServiceProviderConfig", routeHandler(appCtx.serviceProviderConfigHandler.Handle))

				router.GET("/Users/:id", routeHandler(appCtx.userGetHandler.Handle))
				router.GET("/Users", routeHandler(appCtx.userQueryHandler.Handle))
				router.POST("/Users", routeHandler(appCtx.userCreateHandler.Handle))
				router.PUT("/Users/:id", routeHandler(appCtx.userReplaceHandler.Handle))
				router.PATCH("/Users/:id", routeHandler(appCtx.userPatchHandler.Handle))
				router.DELETE("/Users/:id", routeHandler(appCtx.userDeleteHandler.Handle))

				router.GET("/Groups/:id", routeHandler(appCtx.groupGetHandler.Handle))
				router.GET("/Groups", routeHandler(appCtx.groupQueryHandler.Handle))
				router.POST("/Groups", routeHandler(appCtx.groupCreateHandler.Handle))
				router.PUT("/Groups/:id", routeHandler(appCtx.groupReplaceHandler.Handle))
				router.PATCH("/Groups/:id", routeHandler(appCtx.groupPatchHandler.Handle))
				router.DELETE("/Groups/:id", routeHandler(appCtx.groupDeleteHandler.Handle))
			}

			appCtx.logger.Info("listening for incoming requests", log.Args{
				"port": ag.httpPort,
			})
			return http.ListenAndServe(fmt.Sprintf(":%d", ag.httpPort), router)
		},
	}
}

