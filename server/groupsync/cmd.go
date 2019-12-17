package groupsync

import (
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"syscall"
)

type args struct {
	//serviceProviderConfigPath string
	//userResourceTypePath      string
	//groupResourceTypePath     string
	//schemasFolderPath         string
	//memoryDB    bool
	natsServers string
}

func Command() *cli.Command {
	args := new(args)
	return &cli.Command{
		Name:        "group-sync",
		Aliases:     []string{"gs", "sync"},
		Description: "Asynchronously refresh user resource for group membership changes",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "nats-urls",
				Aliases:     []string{"n"},
				Usage:       "comma delimited URLs to the NATS servers",
				EnvVars:     []string{"NATS_SERVERS"},
				Destination: &args.natsServers,
				Required:    true,
				Value:       "nats://localhost:4222",
			},
		},
		Action: func(cliContext *cli.Context) error {
			appCtx := new(appContext)
			if err := appCtx.initialize(args); err != nil {
				return err
			}
			defer appCtx.natConn.Close()

			r, err := Receiver(appCtx.natConn, appCtx.userDB, appCtx.groupDB, appCtx.logger)
			if err != nil {
				return err
			}
			defer r.Close()

			term := make(chan os.Signal)
			signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)
			<-term

			return nil
		},
	}
}
