module github.com/imulab/go-scim

go 1.13

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cenkalti/backoff/v4 v4.0.0
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/imulab/go-scim/mongo/v2 v2.0.0
	github.com/imulab/go-scim/pkg/v2 v2.0.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/rs/zerolog v1.17.2
	github.com/satori/go.uuid v1.2.0
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/stretchr/testify v1.4.0
	github.com/urfave/cli/v2 v2.1.1
	go.mongodb.org/mongo-driver v1.2.1
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
)

replace github.com/imulab/go-scim/mongo/v2 => ./mongo/v2

replace github.com/imulab/go-scim/pkg/v2 => ./pkg/v2
