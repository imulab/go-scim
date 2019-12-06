# GoSCIM

> GoSCIM aims to be a fully featured implementation of [SCIM v2](http://www.simplecloud.info/) specifiction. It provides basic building blocks to SCIM functions and a functional out-of-box server. It is also designed with extensibility in mind to make customizations easy.

**Caution** This is the early stage of `v2.0.0` version of go-scim. We are now at `v2.0.0-m1` ([release notes](https://github.com/imulab/go-scim/releases/tag/v2.0.0-m1)). This second major release will introduce drastic changes to the way resources are handled in the system. 

For the currently stable version, checkout tag `v1.0.1`, or go to [here](https://github.com/imulab/go-scim/tree/v1.0.1).

## Features in v2

- Reflection free operations on resources
- Property event system
- Direct serialization and deserialization in JSON and other data exchange formats
- Enhanced attributes model to allow for custom metadata
- Robust SCIM path and filter parsing
- Resource filters to allow for custom resource processing
- Feature provider interfaces to allow 3rd party integration

## Installation and Usage

The project is in the early stage of `v2.0.0`. The current milestone does not have the ready-to-go server implementation yet. This will come in later milestones.

As for now, to check out the functionalities included in the tests:

```
$ dep ensure
$ go test ./pkg/... 
```

## Documentation Index (TBD)

- [Project orientation](#)
- [Extensible attributes](#)
- [Resource model and property structure](#)
- [Path, filter and CRUD](#)
- [Resource filters](#)
- [Feature providers](#)
- [Integration points](#)

## Road Map

While the fundamentals of the functions are delivered in `v2.0.0-m1`, we are still hard at work to deliver the rest. In the coming weeks and months, the rest of functions towrads `v2.0.0` will be released.
In addition to the scheduled functions, tests and documentations will also be added.

- `v2.0.0-m2` to introduce package refactoring, query functionality, HTTP handlers.
- `v2.0.0-m3` to introduce a working server that bootstraps all functions.
- `v2.0.0-m4` to exprimentally tackle resource root query and bulk operations.
- `v2.0.0-m5` to (re-)introduce mongo db persistence
- `v2.0.0-rc1` to complete tests and documentations

As for after the release of `v2.0.0`, more features are being planned. The list includes:
- [SCIM Password Management Extension](https://tools.ietf.org/id/draft-hunt-scim-password-mgmt-00.txt)
- Authentication endpoints
- [SCIM Soft Delete](https://tools.ietf.org/id/draft-ansari-scim-soft-delete-00.txt)
- [SCIM Notify](https://tools.ietf.org/id/draft-hunt-scim-notify-00.txt)
