package api

import (
	"fmt"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/cli/v2"
	"net/http"
)

type args struct {
	serviceProviderConfigPath string
	userResourceTypePath      string
	groupResourceTypePath     string
	schemasFolderPath         string
	memoryDB                  bool
	httpPort                  int
	natsServers               string
}

func Command() *cli.Command {
	args := new(args)
	return &cli.Command{
		Name:        "api",
		Usage:       "Serves API for SCIM (Simple Cloud Identity Management) protocol",
		Description: "Manage state of resources defined in the SCIM (Simple Cloud Identity Management) protocol",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "user",
				Aliases:     []string{"u"},
				Usage:       "Absolute file path to User resource type JSON definition",
				EnvVars:     []string{"USER_RESOURCE_TYPE"},
				Required:    true,
				Destination: &args.userResourceTypePath,
			},
			&cli.StringFlag{
				Name:        "group",
				Aliases:     []string{"g"},
				Usage:       "Absolute file path to Group resource type JSON definition",
				EnvVars:     []string{"GROUP_RESOURCE_TYPE"},
				Required:    true,
				Destination: &args.groupResourceTypePath,
			},
			&cli.StringFlag{
				Name:        "schemas",
				Aliases:     []string{"s"},
				Usage:       "Absolute path to the folder containing all schema JSON definitions",
				EnvVars:     []string{"SCHEMAS"},
				Required:    true,
				Destination: &args.schemasFolderPath,
			},
			&cli.StringFlag{
				Name:        "spc",
				Aliases:     []string{"c"},
				Usage:       "Absolute path to Service Provider Config JSON definition",
				EnvVars:     []string{"SERVICE_PROVIDER_CONFIG"},
				Required:    true,
				Destination: &args.serviceProviderConfigPath,
			},
			&cli.BoolFlag{
				Name:        "memory",
				Usage:       "Use in memory database",
				Value:       false,
				Destination: &args.memoryDB,
			},
			&cli.StringFlag{
				Name:        "nats-url",
				Aliases:     []string{"n"},
				Usage:       "comma delimited URLs to the NATS servers",
				EnvVars:     []string{"NATS_SERVERS"},
				Required:    true,
				Value:       "nats://localhost:4222",
				Destination: &args.natsServers,
			},
			&cli.IntFlag{
				Name:        "port",
				Aliases:     []string{"p"},
				Usage:       "HTTP port that the server listens on",
				EnvVars:     []string{"HTTP_PORT"},
				Value:       8080,
				Destination: &args.httpPort,
			},
		},
		Action: func(context *cli.Context) error {
			var appCtx *appContext
			{
				appCtx = new(appContext)
				if err := appCtx.initialize(args); err != nil {
					return err
				}
				appCtx.logger.Info("application appContext initialized", log.Args{})
			}

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
				"port": args.httpPort,
			})
			return http.ListenAndServe(fmt.Sprintf(":%d", args.httpPort), router)
		},
	}
}
