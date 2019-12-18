module github.com/imulab/go-scim/mongo

require (
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/imulab/go-scim/core v0.0.0
	github.com/stretchr/testify v1.4.0
	github.com/tidwall/pretty v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.2.0
)

replace github.com/imulab/go-scim/core => ../core

replace github.com/imulab/go-scim/protocol => ../protocol
