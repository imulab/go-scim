package groupsync

import (
	"context"
	"github.com/imulab/go-scim/protocol/log"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"syscall"
)

// Return a command that starts a process to synchronize group membership of user resources.
func Command() *cli.Command {
	ag := new(arguments)
	return &cli.Command{
		Name:        "group-sync",
		Aliases:     []string{"gs", "sync"},
		Description: "Asynchronously refresh user resource for group membership changes",
		Flags: ag.Flags(),
		Action: func(cliContext *cli.Context) error {
			appCtx := new(appContext)
			if err := appCtx.initialize(ag); err != nil {
				return err
			}
			defer func() {
				_ = appCtx.rabbitChannel.Close()
				_ = appCtx.mongoClient.Disconnect(context.Background())
			}()

			ctx, cancelFunc := context.WithCancel(context.Background())
			safeExit, err := StartConsumer(ctx, appCtx, ag)
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
