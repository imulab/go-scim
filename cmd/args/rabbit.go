package args

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/urfave/cli/v2"
)

// RabbitMQ is the options related to connecting to RabbitMQ message middleware
type RabbitMQ struct {
	Username string
	Password string
	Host     string
	Port     int
	VHost    string
	Options  string
}

// Url returns the RabbitMQ connection url from the options set.
func (arg RabbitMQ) Url() string {
	url := "amqp://"
	if len(arg.Username) > 0 {
		url += arg.Username
		if len(arg.Password) > 0 {
			url += fmt.Sprintf(":%s", arg.Password)
		}
		url += "@"
	}
	url += arg.Host
	if arg.Port > 0 {
		url += fmt.Sprintf(":%d", arg.Port)
	}
	if len(arg.VHost) > 0 {
		url += fmt.Sprintf("/%s", arg.VHost)
	}
	if len(arg.Options) > 0 {
		url += fmt.Sprintf("?%s", arg.Options)
	}
	return url
}

// Connect returns a connected RabbitMQ AMQP Channel using the options set, or an error. This method respects the
// provided context, and makes the connection process cancellable and timeout-able. The default RabbitMQ connection
// timeouts were kept in place, with the default timeout of 30 seconds.
func (arg RabbitMQ) Connect(ctx context.Context) (*amqp.Channel, error) {
	var (
		amqpChan = make(chan *amqp.Channel, 1)
		errChan  = make(chan error, 1)
	)
	defer close(amqpChan)
	defer close(errChan)

	go func() {
		conn, err := amqp.Dial(arg.Url())
		if err != nil {
			errChan <- err
			return
		}

		ch, err := conn.Channel()
		if err != nil {
			errChan <- err
			return
		}

		amqpChan <- ch
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errChan:
		return nil, err
	case ch := <-amqpChan:
		return ch, nil
	}
}

func (arg RabbitMQ) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "rabbit-host",
			Usage:       "Hostname of RabbitMQ",
			EnvVars:     []string{"RABBIT_HOST"},
			Value:       "localhost",
			Destination: &arg.Host,
		},
		&cli.IntFlag{
			Name:        "rabbit-port",
			Usage:       "Port of RabbitMQ",
			EnvVars:     []string{"RABBIT_PORT"},
			Value:       5672,
			Destination: &arg.Port,
		},
		&cli.StringFlag{
			Name:        "rabbit-username",
			Usage:       "Username for RabbitMQ",
			EnvVars:     []string{"RABBIT_USERNAME"},
			Destination: &arg.Username,
		},
		&cli.StringFlag{
			Name:        "rabbit-password",
			Usage:       "Password for RabbitMQ",
			EnvVars:     []string{"RABBIT_PASSWORD"},
			Destination: &arg.Password,
		},
		&cli.StringFlag{
			Name:        "rabbit-vhost",
			Usage:       "Virtual host for RabbitMQ",
			EnvVars:     []string{"RABBIT_VHOST"},
			Destination: &arg.VHost,
		},
		&cli.StringFlag{
			Name:        "rabbit-options",
			Usage:       "Options for RabbitMQ",
			EnvVars:     []string{"RABBIT_OPT"},
			Destination: &arg.Options,
		},
	}
}