<img src="./asset/scim.png" width="200">

> GoSCIM aims to be a fully featured implementation of [SCIM v2](http://www.simplecloud.info/) specification. It 
provides opinion-free and extensible building blocks, as well as an opinionated server implementation.

## :rocket: TLDR

###### *Requirements:* [Docker](https://docs.docker.com/get-started/) and [Docker Compose](https://docs.docker.com/compose/gettingstarted/)

```bash
# Builds docker image and starts local docker-compose stack
make docker compose
```

## :file_folder: Project structure

Since v1, the project has grown into three independent modules. 
- [pkg module](https://github.com/imulab/go-scim/tree/master/pkg/v2) evolved from most of the original building blocks. 
This module provides customizable, extensible and opinion free implementation of the SCIM specification.
- [mongo module](https://github.com/imulab/go-scim/tree/master/mongo/v2) evolved from the original mongo package. 
This module provides persistence capabilities to MongoDB.
- [server module](https://github.com/imulab/go-scim) evolved from the original example server implementation. It is now 
an __opinionated__ personal server implementation that depends on the above two modules.

Documentation for the individual modules can be viewed in their respective directories and godoc badge links.

## :no_entry_sign: End of v1

Due to limited time and resource and a drastic new design in v2, the building blocks and mongo package from v1 will no 
longer be maintained. The `github.com/imulab/go-scim` will go on as an opinionated personal implementation, which may or
may not resonate with everyone's use case.

However, other modules will continue to be maintained and accept changes to allow reasonable use cases, while remaining
true to the specification itself.
