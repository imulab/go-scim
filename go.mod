module github.com/imulab/go-scim

go 1.13

require (
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cenkalti/backoff/v4 v4.2.0
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/imulab/go-scim/mongo/v2 v2.0.0
	github.com/imulab/go-scim/pkg/v2 v2.0.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/klauspost/compress v1.16.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/montanaflynn/stats v0.7.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.1.4 // indirect
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/rs/zerolog v1.29.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.8.2
	github.com/urfave/cli/v2 v2.25.0
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	go.mongodb.org/mongo-driver v1.11.2
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/sync v0.1.0
	golang.org/x/tools v0.7.0 // indirect
)

replace github.com/imulab/go-scim/mongo/v2 => ./mongo/v2

replace github.com/imulab/go-scim/pkg/v2 => ./pkg/v2
