# Technology Stack

**Analysis Date:** 2026-05-07

## Languages

**Primary:**
- Go 1.13 - Core SCIM implementation and all application code
  - Defined in `go.mod` at root of legacy codebase
  - Multi-module structure with `pkg/v2` and `mongo/v2` as separate modules

## Runtime

**Environment:**
- Go 1.13 runtime for local development
- Debian buster-slim for production Docker images
- Linux/amd64 as primary deployment target (referenced in Dockerfile and Makefile)

**Package Manager:**
- Go modules (go.mod/go.sum)
- Multiple module files:
  - `.legacy/go.mod` - Root module
  - `.legacy/pkg/v2/go.mod` - SCIM protocol library
  - `.legacy/mongo/v2/go.mod` - MongoDB adapter

## Frameworks

**Core HTTP:**
- `github.com/julienschmidt/httprouter` v1.3.0 - Lightweight HTTP router for REST endpoints
  - Used in `cmd/api/cmd.go` for routing SCIM API endpoints (GET, POST, PUT, PATCH, DELETE)

**CLI Framework:**
- `github.com/urfave/cli/v2` v2.1.1 - Command-line interface framework
  - Bootstraps two commands: `api` and `group-sync`
  - Entry point: `bootstrap.go` (package main)
  - Command files: `cmd/api/cmd.go`, `cmd/groupsync/cmd.go`

**Logging:**
- `github.com/rs/zerolog` v1.17.2 - Structured JSON logging
  - Configuration: `cmd/internal/args/logger.go`
  - Supports log levels: INFO, ERROR, DEBUG, WARN, FATAL
  - Outputs to stderr with timestamps

**Testing:**
- `github.com/stretchr/testify` v1.4.0 - Testing assertions and mocking
  - Used across `pkg/v2` and `mongo/v2` modules

## Key Dependencies

**Critical Infrastructure:**
- `go.mongodb.org/mongo-driver` v1.2.1 - MongoDB client driver
  - Used by `mongo/v2` adapter for persistence
  - Connection management in `cmd/internal/args/database.go`

- `github.com/streadway/amqp` v0.0.0-20200108173154-1c71cc93ed71 - RabbitMQ AMQP client
  - Used by `group-sync` command for async group membership updates
  - Queue operations in `cmd/internal/groupsync/rabbit.go`

**Resilience & Retry:**
- `github.com/cenkalti/backoff` v2.2.1 - Basic backoff (legacy)
- `github.com/cenkalti/backoff/v4` v4.0.0 - Exponential backoff for retries
  - Used for MongoDB connection retry in `cmd/internal/args/database.go`
  - Used for RabbitMQ connection retry in `cmd/internal/args/rabbit.go`

**Utilities:**
- `github.com/satori/go.uuid` v1.2.0 - UUID generation for resource IDs
- `golang.org/x/sync` v0.0.0-20190911185100-cd5d95a43a6e - Synchronization primitives

**Testing/Development:**
- `github.com/ory/dockertest` v3.3.5 - Docker-based integration testing setup
  - Enables ephemeral test database/queue containers

## Configuration

**Environment Variables:**
Primary configuration via environment variables (also supported as CLI flags):

MongoDB:
- `MONGO_HOST` (default: localhost)
- `MONGO_PORT` (default: 27017)
- `MONGO_USERNAME`
- `MONGO_PASSWORD`
- `MONGO_DATABASE`
- `MONGO_OPT` (connection options, e.g., authMechanism=SCRAM-SHA-1)
- `MONGO_METADATA_DIR` - Path to MongoDB metadata JSON files

RabbitMQ:
- `RABBIT_HOST` (default: localhost)
- `RABBIT_PORT` (default: 5672)
- `RABBIT_USERNAME`
- `RABBIT_PASSWORD`
- `RABBIT_VHOST`
- `RABBIT_OPT` (connection options)

Application:
- `HTTP_PORT` (default: 5000 for api command)
- `LOG_LEVEL` (default: INFO)
- `REQUEUE_LIMIT` (for group-sync command)

SCIM Configuration:
- `SERVICE_PROVIDER_CONFIG` - Path to service provider config JSON
- `SCHEMAS_DIR` - Path to SCIM schemas directory
- `USER_RESOURCE_TYPE` - Path to user resource type definition JSON
- `GROUP_RESOURCE_TYPE` - Path to group resource type definition JSON
- `MONGO_METADATA_DIR` - Path to MongoDB-specific metadata

**Static Configuration Files:**
- `.legacy/public/service_provider_config.json` - SCIM ServiceProviderConfig
- `.legacy/public/schemas/` - SCIM schema definitions
- `.legacy/public/resource_types/` - Resource type definitions
- `.legacy/public/mongo_metadata/` - MongoDB collection metadata

## Build Tooling

**Build System:**
- GNU Make - Makefile at `.legacy/Makefile`
  - `make build` - Compile binary for current GOOS/GOARCH
  - `make deps` - Download Go dependencies
  - `make test` - Run test suite with race detector
  - `make docker` - Build Docker image
  - `make compose` - Start docker-compose stack
  - Cross-compilation support for GOOS/GOARCH

**Docker:**
- Multi-stage Dockerfile: `.legacy/Dockerfile`
  - Builder stage: golang:1.13-buster
  - Final stage: debian:buster-slim
  - Binary output: `/usr/bin/scim`
  - Copies public assets to `/usr/share/scim/public`

**Docker Compose:**
- `.legacy/docker-compose.yml` (v3)
- Services:
  - `mediainit` - Alpine container for volume setup
  - `db` - bitnami/mongodb (latest)
  - `rabbit` - bitnami/rabbitmq (latest)
  - `group` - scim:latest (group-sync command)
  - `server` - scim:latest (api command)

## Platform Requirements

**Development:**
- Go 1.13+
- GNU Make
- Docker and Docker Compose (for local stack)

**Production:**
- Docker runtime (containerized Debian buster-slim)
- MongoDB 1.2.x+ (via mongo-driver v1.2.1)
- RabbitMQ (AMQP protocol)

---

*Stack analysis: 2026-05-07*
