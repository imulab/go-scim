# External Integrations

**Analysis Date:** 2026-05-07

## APIs & External Services

**SCIM Protocol:**
- SCIM 2.0 specification (https://scim.cloud/)
  - Users and Groups resources
  - Search, filter, and projection support
  - Standard HTTP methods: GET, POST, PUT, PATCH, DELETE
  - Endpoints: `/Users`, `/Groups`, `/Schemas`, `/ResourceTypes`, `/ServiceProviderConfig`

## Data Storage

**Databases:**
- MongoDB 1.2.x+
  - Driver: `go.mongodb.org/mongo-driver` v1.2.1
  - Connection: Environment variables (MONGO_HOST, MONGO_PORT, MONGO_USERNAME, MONGO_PASSWORD, MONGO_DATABASE)
  - Connection string format: `mongodb://[user:pass@]host[:port]/[database][?options]`
  - Adapter module: `.legacy/mongo/v2/`
  - Key files: 
    - `mongo/v2/db.go` - Core database interface
    - `mongo/v2/serialize.go` - SCIM to MongoDB serialization
    - `mongo/v2/deserialize.go` - MongoDB to SCIM deserialization
    - `mongo/v2/filter.go` - SCIM filter to MongoDB query translation
  - Metadata registration: Loads JSON metadata from `MONGO_METADATA_DIR`
  - Authentication: Supports SCRAM-SHA-1 via MONGO_OPT parameter

**File Storage:**
- Local filesystem only
  - Static SCIM configuration files in `.legacy/public/`
  - Service provider config: `.legacy/public/service_provider_config.json`
  - Schema definitions: `.legacy/public/schemas/`
  - Resource types: `.legacy/public/resource_types/`
  - MongoDB metadata: `.legacy/public/mongo_metadata/`

**Caching:**
- None detected (no Redis, Memcached, or similar)

## Message Queue

**Message Broker:**
- RabbitMQ (AMQP 0.9.1)
  - Client: `github.com/streadway/amqp` v0.0.0-20200108173154-1c71cc93ed71
  - Connection: Environment variables (RABBIT_HOST, RABBIT_PORT, RABBIT_USERNAME, RABBIT_PASSWORD, RABBIT_VHOST)
  - Connection string format: `amqp://[user[:pass]@]host[:port][/vhost][?options]`
  - Purpose: Async group membership synchronization when group resources change
  
  **Queue Configuration:**
  - Queue name: `group_sync`
  - Exchange: default (empty string)
  - Durable: true
  - Implementation files:
    - `cmd/internal/groupsync/rabbit.go` - Queue declaration and management
    - `cmd/groupsync/consumer.go` - Message consumer implementation
    - `cmd/api/service.go` - Message producer when groups are created/modified

  **Message Flow:**
  - Producer: API service (`groupCreated`, `groupReplaced`, `groupPatched` wrappers)
  - Consumer: `group-sync` command
  - Payload: Group sync messages with member change information
  - Configuration: `REQUEUE_LIMIT` controls retry behavior

## Authentication & Identity

**Auth Provider:**
- Custom/None - No identity provider detected
- Application appears to manage SCIM users/groups directly
- No OAuth, LDAP, or external IdP integration visible
- SCIM resources are the authoritative identity store

## HTTP Server Framework

**HTTP Listener:**
- Standard Go `net/http` package
- Router: `github.com/julienschmidt/httprouter` v1.3.0
- Listening address: `0.0.0.0:[HTTP_PORT]` (default port 5000)
- SCIM API endpoints defined in `cmd/api/cmd.go`:
  - `/ServiceProviderConfig` - GET
  - `/Schemas` - GET
  - `/Schemas/:id` - GET
  - `/ResourceTypes` - GET
  - `/ResourceTypes/:id` - GET
  - `/Users` - GET (search), POST (create)
  - `/Users/:id` - GET, PUT (replace), PATCH, DELETE
  - `/Groups` - GET (search), POST (create)
  - `/Groups/:id` - GET, PUT (replace), PATCH, DELETE
  - `/health` - GET (health check endpoint)

**Health Checks:**
- Endpoint: `GET /health`
- Status checks: MongoDB connectivity, RabbitMQ connection
- Returns JSON with service, mongodb_connection, rabbitmq_connection status
- HTTP 200 if all up, HTTP 500 if any down

## Monitoring & Observability

**Error Tracking:**
- None detected

**Logs:**
- Structured JSON logging via `github.com/rs/zerolog` v1.17.2
- Output: stderr
- Levels: INFO (default), ERROR, DEBUG, WARN, FATAL
- Configuration: `LOG_LEVEL` environment variable
- Log output includes timestamps and structured fields

## CI/CD & Deployment

**Hosting:**
- Docker containers (deployed via Docker Compose or Kubernetes-style orchestration)
- Base image: debian:buster-slim
- Build image: golang:1.13-buster

**CI Pipeline:**
- None detected (no .github/workflows, .gitlab-ci.yml, or similar)
- Local build via Makefile

**Deployment:**
- Docker Compose file: `.legacy/docker-compose.yml`
- Services: API server, group-sync worker, MongoDB, RabbitMQ
- Network: `app-tier` bridge network
- Volumes: `mongo_data`, `rabbitmq_data`

## Environment Configuration

**Required Environment Variables (API Server):**
```
HTTP_PORT=5000
MONGO_HOST=db
MONGO_PORT=27017
MONGO_USERNAME=user
MONGO_PASSWORD=password123
MONGO_DATABASE=scim_database
MONGO_OPT=authMechanism=SCRAM-SHA-1
RABBIT_HOST=rabbit
RABBIT_PORT=5672
RABBIT_USERNAME=user
RABBIT_PASSWORD=password123
SERVICE_PROVIDER_CONFIG=/usr/share/scim/public/service_provider_config.json
SCHEMAS_DIR=/usr/share/scim/public/schemas
USER_RESOURCE_TYPE=/usr/share/scim/public/resource_types/user_resource_type.json
GROUP_RESOURCE_TYPE=/usr/share/scim/public/resource_types/group_resource_type.json
MONGO_METADATA_DIR=/usr/share/scim/public/mongo_metadata
```

**Required Environment Variables (Group-Sync Worker):**
```
REQUEUE_LIMIT=1
MONGO_HOST=db
MONGO_PORT=27017
MONGO_USERNAME=user
MONGO_PASSWORD=password123
MONGO_DATABASE=scim_database
MONGO_OPT=authMechanism=SCRAM-SHA-1
RABBIT_HOST=rabbit
RABBIT_PORT=5672
RABBIT_USERNAME=user
RABBIT_PASSWORD=password123
SERVICE_PROVIDER_CONFIG=/usr/share/scim/public/service_provider_config.json
SCHEMAS_DIR=/usr/share/scim/public/schemas
USER_RESOURCE_TYPE=/usr/share/scim/public/resource_types/user_resource_type.json
GROUP_RESOURCE_TYPE=/usr/share/scim/public/resource_types/group_resource_type.json
MONGO_METADATA_DIR=/usr/share/scim/public/mongo_metadata
```

**Secrets Location:**
- Environment variables only
- No secrets files, vaults, or encrypted configuration files detected
- Credentials passed via docker-compose environment sections or runtime secrets injection

## Webhooks & Callbacks

**Incoming:**
- None detected (SCIM API is not a webhook receiver)

**Outgoing:**
- Group membership changes trigger RabbitMQ messages to `group_sync` queue
- Internal async callback via message queue, not external webhooks
- Producer: `cmd/api/service.go` wraps group CRUD operations
- Consumer: `cmd/groupsync/consumer.go` processes and applies changes

## Data Format

**Wire Format:**
- JSON (application/scim+json)
- SCIM v2.0 JSON structure
- Content-Type header: `application/scim+json`

**Serialization:**
- SCIM resource trees mapped to MongoDB documents via `.legacy/mongo/v2/serialize.go`
- Properties support complex types: strings, booleans, integers, decimals, binaries, references
- Multi-valued attributes stored as arrays
- Schema validation against loaded schemas

---

*Integration audit: 2026-05-07*
