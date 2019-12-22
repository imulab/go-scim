package args

import (
	"fmt"
	"github.com/streadway/amqp"
	"github.com/urfave/cli/v2"
)

type Rabbit struct {
	Username string
	Password string
	Host     string
	Port     int
	VHost    string
	Options  string
}

func (arg *Rabbit) Url() string {
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

func (arg *Rabbit) Connect() (*amqp.Channel, error) {
	conn, err := amqp.Dial(arg.Url())
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (arg *Rabbit) Flags() []cli.Flag {
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
