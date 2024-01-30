module github.com/imulab/go-scim

go 1.13

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/google/uuid v1.3.1
	github.com/imulab/go-scim/mongo/v2 v2.0.0
	github.com/imulab/go-scim/pkg/v2 v2.0.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/rs/zerolog v1.29.0
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.8.2
	github.com/urfave/cli/v2 v2.25.7
	go.mongodb.org/mongo-driver v1.11.3
	golang.org/x/sync v0.3.0
)

replace github.com/imulab/go-scim/mongo/v2 => ./mongo/v2

replace github.com/imulab/go-scim/pkg/v2 => ./pkg/v2
