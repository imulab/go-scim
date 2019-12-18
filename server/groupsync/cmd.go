package groupsync

import (
	"context"
	"github.com/imulab/go-scim/protocol/log"
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
	requeueLimit int
	rabbitMqAddress string
}

func Command() *cli.Command {
	args := new(args)
	return &cli.Command{
		Name:        "group-sync",
		Aliases:     []string{"gs", "sync"},
		Description: "Asynchronously refresh user resource for group membership changes",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "rabbit-address",
				Usage:       "AMQP connection string to RabbitMQ",
				EnvVars:     []string{"RABBIT_ADDRESS"},
				Destination: &args.rabbitMqAddress,
				Required:    true,
				Value:       "amqp://guest:guest@localhost:5672/",
			},
		},
		Action: func(cliContext *cli.Context) error {
			appCtx := new(appContext)
			if err := appCtx.initialize(args); err != nil {
				return err
			}
			defer func() {
				_ = appCtx.rabbitCh.Close()
			}()

			ctx, cancelFunc := context.WithCancel(context.Background())
			safeExit, err := StartConsumer(ctx, appCtx, args)
			if err != nil {
				return err
			}

			term := make(chan os.Signal)
			signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)
			<-term
			appCtx.logger.Info("received terminate signal, waiting to abort", log.Args{})
			cancelFunc()
			<-safeExit

			return nil
		},
	}
}
