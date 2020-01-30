<img src="./asset/scim.png" width="200">

> GoSCIM aims to be a fully featured implementation of [SCIM v2](http://www.simplecloud.info/) specification. It 
provides opinion-free and extensible building blocks, as well as an opinionated server implementation.

## TLDR

###### *Requirements:* [Docker](https://docs.docker.com/get-started/) and [Docker Compose](https://docs.docker.com/compose/gettingstarted/)

```bash
# Builds docker image and starts local docker-compose stack
make docker compose
```

## Project structure

Since v1, the project has grown into three independent modules. 
- `github.com/imulab/go-scim/pkg/v2` module evolved from most of the original building blocks. This module provides
customizable, extensible and opinion free implementation of the SCIM specification.
- `github.com/imulab/go-scim/mongo/v2` module evolved from the original mongo package. This module provides persistence
capabilities to MongoDB.
- `github.com/imulab/go-scim` module evolved from the original example server implementation. It is now a __opinionated__
personal server implementation that depends on the above two modules.

Documentation for the [pkg](https://github.com/imulab/go-scim/tree/master/pkg/v2) and [mongo](https://github.com/imulab/go-scim/tree/master/mongo/v2) module can be viewed in their respective directories.

## End of v1

Due to limited time and resource and a drastic new design in v2, the building blocks and mongo package from v1 will no 
longer be maintained. The `github.com/imulab/go-scim` will go on as an opinionated personal implementation, which may or
may not resonate with everyone's use case.

However, other modules will continue to be maintained and accept changes to allow reasonable use cases, while remaining
true to the specification itself.
