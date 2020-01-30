# SCIM

[![GoDoc](https://godoc.org/github.com/imulab/go-scim/pkg/v2?status.svg)](https://godoc.org/github.com/imulab/go-scim/pkg/v2)

This module implements features described in [RFC 7643 - SCIM: Core Schema](https://tools.ietf.org/html/rfc7643) and
[RFC 7644 - SCIM: Protocol](https://tools.ietf.org/html/rfc7644), along with custom features that address specific 
implementation difficulties.

The goal of this package is to provide __extensible__, __maintainable__ and __easy-to-use__ building blocks for 
implementing a SCIM service with relatively good performance.

It is not the goal of this package to enable every possible use case, or to provide an out-of-box SCIM service.

> This module evolved from [v1](https://github.com/imulab/go-scim/releases/tag/v1.0.1) of 
  [github.com/imulab/go-scim](https://github.com/imulab/go-scim/tree/v1.0.1). 
>
> We had since learnt couple lessons (issue [11](https://github.com/imulab/go-scim/issues/11) and [39](https://github.com/imulab/go-scim/issues/39)) 
  from this initial implementation and had decided to pursue a completely different design. Due to the drastic
  difference between the two major versions, we decided to stop maintaining v1 and devote limited resources and future 
  efforts on maintaining v2. 
>
> We apologize for any inconvenience caused by this decision. The original v1 can still be checked out via tags.

## :bulb: Usage

To get this package:

```bash
# make sure Go 1.13
go get github.com/imulab/go-scim/pkg/v2
```

## :gift: Features in v2

- :free: Reflection free operations on resources
- :mailbox_with_mail: Property subscribers
- :rocket: Direct serialization and deserialization in JSON and other data exchange formats
- :wrench: Enhanced attributes model to allow for custom metadata
- :thumbsup: Robust SCIM path and filter parsing
- :fast_forward: Resource filters to allow for custom resource processing

## :file_folder: Project Structure

Features in this module are separated into different directories:
- `spec` directory implements the foundation of SCIM resource type definition
- `prop` directory implements `Property` which holds pieces of resource data
- `json` directory implements direct serialization and deserialization between SCIM resource and its JSON format
- `crud` directory implements parsing and evaluation capabilities for SCIM path and SCIM filters
- `db` directory introduces a standard `DB` interface and a in-memory implementation
- `annotation` directory documents internally used attribute annotations and their purpose
- `groupsync` directory implements utilities to synchronize change in `Group.members` with `User.groups`
- `service` directory implements CRUD services that carry out most of the protocol work
- `handlerutil` directory implements utilities that help parsing and rendering HTTP, assuming Go's HTTP abstraction

For detailed documentation, please check out README of individual directories, or GoDoc.

## :bullettrain_side: Road Map

After delivering the v2.0.0 which will cover most features, efforts will be directed toward:
- ResourceType(s) and Schema(s) endpoints (see [issue 40](https://github.com/imulab/go-scim/issues/40))
- Root query
- Bulk operations
- SCIM password management extension
- SCIM soft delete extension