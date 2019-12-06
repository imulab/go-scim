# GoSCIM

> GoSCIM aims to be a fully featured implementation of [SCIM v2](http://www.simplecloud.info/) specifiction. It provides basic building blocks to SCIM functions and a functional out-of-box server. It is also designed with extensibility in mind to make customizations easy.

**Caution** This is the early stage of `v2.0.0` version of go-scim. We are now at `v2.0.0-mc1` ([release notes](https://github.com/imulab/go-scim/releases/tag/v2.0.0-mc1)). This second major release will introduce drastic changes to the way resources are handled in the system. For the currently stable version, checkout tag `v1.0.1`

## Features in v2

- Reflection free operations on resources
- Property event system
- Direct serialization and deserialization in JSON and other data exchange formats
- Enhanced attributes model to allow for custom metadata
- Robust SCIM path and filter parsing
- Resource filters to allow for custom resource processing
- Feature provider interfaces to allow 3rd party integration

## Installation and Usage

This section will be 