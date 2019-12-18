module github.com/imulab/go-scim/server

require (
	github.com/imulab/go-scim/core v0.0.0
	github.com/imulab/go-scim/protocol v0.0.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/rs/zerolog v1.17.2
	github.com/satori/go.uuid v1.2.0
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/urfave/cli/v2 v2.0.0
)

replace github.com/imulab/go-scim/core => ../core

replace github.com/imulab/go-scim/protocol => ../protocol
