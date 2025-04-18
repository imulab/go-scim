module github.com/imulab/go-scim/mongo/v2

require (
	github.com/imulab/go-scim/pkg/v2 v2.0.0
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/stretchr/testify v1.10.0
	go.mongodb.org/mongo-driver v1.17.3
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/containerd/continuity v0.4.5 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/gotestyourself/gotestyourself v1.3.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/lib/pq v0.0.0-20180327071824-d34b9ff171c2 // indirect
	github.com/moby/sys/user v0.3.0 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/opencontainers/runc v1.2.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/imulab/go-scim/pkg/v2 => ../../pkg/v2

go 1.23.0

toolchain go1.24.2
