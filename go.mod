module github.com/imulab/go-scim

go 1.13

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cenkalti/backoff/v4 v4.0.0
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/google/uuid v1.3.1
	github.com/imulab/go-scim/mongo/v2 v2.0.0
	github.com/imulab/go-scim/pkg/v2 v2.0.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/opencontainers/runc v1.0.0-rc9 // indirect
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/rs/zerolog v1.17.2
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/urfave/cli/v2 v2.1.1
	go.mongodb.org/mongo-driver v1.2.1
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200121082415-34d275377bf9 // indirect
	gopkg.in/yaml.v2 v2.2.7 // indirect
)

replace github.com/imulab/go-scim/mongo/v2 => ./mongo/v2

replace github.com/imulab/go-scim/pkg/v2 => ./pkg/v2
