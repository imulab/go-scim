module github.com/imulab/go-scim/server

require (
	github.com/imulab/go-scim/core v0.0.0
	github.com/imulab/go-scim/protocol v0.0.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/nats-io/nats-server/v2 v2.1.2 // indirect
	github.com/nats-io/nats.go v1.9.1
	github.com/rs/zerolog v1.17.2
	github.com/urfave/cli/v2 v2.0.0
)

replace github.com/imulab/go-scim/core => ../core

replace github.com/imulab/go-scim/protocol => ../protocol
