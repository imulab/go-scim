package groupsync

import (
	"context"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"syscall"
)

// Command returns a cli.Command that starts a process to synchronize group membership of user resources.
func Command() *cli.Command {
	args := newArgs()
	return &cli.Command{
		Name:        "group-sync",
		Aliases:     []string{"gs", "sync"},
		Description: "Asynchronously refresh user resource for group membership changes",
		Flags:       args.Flags(),
		Action: func(_ *cli.Context) error {
			app := args.Initialize()
			defer app.Close()

			ctx, cancelFunc := context.WithCancel(context.Background())
			safeExit, err := app.MessageConsumer().Start(ctx)
			if err != nil {
				return err
			}

			term := make(chan os.Signal)
			signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)
			<-term
			app.Logger().Info().Msg("received terminate signal, waiting to abort")
			cancelFunc()
			<-safeExit

			return nil
		},
	}
}
