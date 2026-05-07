# Architecture

**Analysis Date:** 2026-05-07

## Pattern Overview

**Overall:** Tree-based property model with schema-driven CRUD and pluggable persistence layer.

**Key Characteristics:**
- Generic in-memory tree-like data model where every node (Property) has schema attribute references
- Schema-first design: all data requirements defined through spec.Attribute
- CRUD operations performed via path traversal on property tree with event propagation
- Pluggable adapter interface (db.DB) for persistence implementations
- Subscription-based event system for reactive processing during mutations
- HTTP API facade consumed by SCIM v2 endpoints via service layer

## Layers

**Specification Layer (spec):**
- Purpose: Define SCIM resource types, schemas, attributes, and core data requirements
- Location: `.legacy/pkg/v2/spec/`
- Contains: Attribute definitions, ResourceType, Schema, Type system, Uniqueness/Mutability/Returned enums
- Depends on: annotation package for attribute metadata
- Used by: All higher layers (prop, crud, service, db adapters)

**Property/Data Model Layer (prop):**
- Purpose: In-memory tree representation of SCIM resources where each node is schema-aware
- Location: `.legacy/pkg/v2/prop/`
- Contains: Property interface with 10 concrete implementations (String, Integer, Boolean, DateTime, Reference, Binary, Decimal, Complex, MultiValued), Resource wrapper, Navigator for traversal, Event/Subscriber system
- Depends on: spec (for Attribute references)
- Used by: crud, json, db adapters, service filters

**CRUD/Path Expression Layer (crud):**
- Purpose: Modify properties via SCIM paths, compile filter/sort expressions, handle resource evaluation
- Location: `.legacy/pkg/v2/crud/`
- Contains: Add/Replace/Delete operations, expression compiler (Filter, Sort, Path, Projection, Query), attribute registration, traversal engine
- Depends on: prop (Navigator), spec (Attribute paths)
- Used by: service layer, json deserialization, db adapters

**JSON Serialization Layer (json):**
- Purpose: Convert between JSON bytes and Property/Resource trees with schema-aware validation
- Location: `.legacy/pkg/v2/json/`
- Contains: Deserialize (JSON → Resource), Serialize (Resource → JSON), SafeSerialize with projection filtering, Scanner state machine
- Depends on: prop (Property/Resource), crud (projection handling), spec
- Used by: Service (Create, Replace, Patch), Handlers

**Database Adapter Interface (db):**
- Purpose: Abstract persistence layer enabling multiple storage backends
- Location: `.legacy/pkg/v2/db/`
- Contains: DB interface (Insert, Get, Replace, Delete, Query, Count), in-memory implementation, noop implementation
- Depends on: prop (Resource), crud (Projection, Sort, Pagination)
- Used by: Service layer, db implementations

**MongoDB Adapter (mongo/v2):**
- Purpose: Concrete implementation of db.DB for MongoDB persistence
- Location: `.legacy/mongo/v2/`
- Contains: MongoDB collection-based resource storage, BSON transformation, filter-to-mongo-query translation, metadata for field name mapping, index management
- Depends on: db interface, crud (expr, Filter), prop, spec
- Used by: API context initialization

**Service Layer (service):**
- Purpose: Implement RFC7644 SCIM endpoint logic with request/response handling
- Location: `.legacy/pkg/v2/service/`
- Contains: Create, Get, Replace, Patch, Delete, Query services; filter pipeline (ReadOnly, UUID, BCrypt, Meta, Validation filters)
- Depends on: db (DB interface), json (Deserialize), prop (Resource), spec, service/filter
- Used by: HTTP handlers

**HTTP Handler Utilities (handlerutil):**
- Purpose: Parse HTTP requests into service requests, serialize service responses to HTTP
- Location: `.legacy/pkg/v2/handlerutil/`
- Contains: Request parsing (CreateRequest, GetRequest, QueryRequest), response writing, error serialization
- Depends on: service (interfaces), json (Serialize), crud (Projection, Sort, Pagination)
- Used by: API handlers

**API Command (cmd/api):**
- Purpose: HTTP server entry point for SCIM API
- Location: `.legacy/cmd/api/`
- Contains: HTTP route handlers (Create, Get, Put, Patch, Delete, Search), application context factory, service initialization
- Depends on: service, handlerutil, spec, db, mongo
- Used by: bootstrap.go

**GroupSync Command (cmd/groupsync):**
- Purpose: Async event consumer for group membership updates
- Location: `.legacy/cmd/groupsync/`
- Contains: RabbitMQ message consumer, group sync service (diff, sync logic)
- Depends on: service, prop, spec, db
- Used by: bootstrap.go

**Facade/Import-Export Layer (facade):**
- Purpose: Convert between custom Go structs and SCIM Resources via struct tags
- Location: `.legacy/pkg/v2/facade/`
- Contains: Export (struct → Resource), Import (Resource → struct), reflect-based field mapping, filtered path support
- Depends on: prop (Resource), crud (path expressions), spec
- Used by: Optional high-level API for struct mapping

## Data Flow

**HTTP Create Request → Persisted Resource:**

1. HTTP POST → `.legacy/cmd/api/CreateHandler()` receives request
2. `handlerutil.CreateRequest()` wraps request body
3. `service.CreateService.Do()` called with request
4. `json.Deserialize()` reads JSON bytes → creates Property tree via Navigator, validates against spec.Attribute
5. Filter pipeline applied: ReadOnlyFilter, UUIDFilter, BCryptFilter, MetaFilter, ValidationFilter
6. `db.DB.Insert()` persists via MongoDB adapter (BSON transform, index checks)
7. Service response returned with Resource
8. `handlerutil.WriteResourceToResponse()` serializes Resource → JSON with projection

**CRUD Path-Based Mutation (e.g., PATCH):**

1. Service receives path expression (e.g., "emails[type eq 'work'].value")
2. `crud.CompilePath()` parses into Expression tree
3. `prop.Navigate()` creates Navigator on root property
4. `crud.Traverse()` walks tree via Navigator using compiled path
5. At target location, `Navigator.Replace()` modifies property
6. Event emitted (EventAssigned/EventUnassigned), propagated up tree
7. Subscribers (e.g., AutoCompactSubscriber) react to events
8. All ancestors notified via propagation, marking dirty state
9. Modified Resource passed to db.DB.Replace()

**Filter/Search Query:**

1. HTTP GET with `filter=emails[type eq 'work'].value ew "@example.com"` and `sort=id asc`
2. Service calls `db.Query()` with filter string, sort, pagination, projection
3. MongoDB adapter translates SCIM filter expression to MongoDB query operators
4. Results deserialized as Resources, projected to requested attributes
5. JSON serialized with projection filtering via `json.SafeSerialize()`

**Schema Registration & Resource Type Setup:**

1. `.legacy/bootstrap.go` → `api.Command()` → `applicationContext.Initialize()`
2. `args.RegisterSchemas()` loads JSON schema definitions from config
3. `spec.Schema` and `spec.ResourceType` parsed and registered
4. `crud.Register()` caches resource type for path compilation
5. Services created with registered resource type reference
6. Database created with resource type to define property structure

## Key Abstractions

**Property:**
- Purpose: Represents a typed data node in the tree matching a spec.Attribute
- Examples: `prop.stringProperty`, `prop.complexProperty`, `prop.multiValuedProperty`
- Pattern: Interface-driven with concrete factory functions (NewString, NewComplex, NewMulti)
- Methods: Raw() (get value), Add/Replace/Delete (mutate with events), Clone, Matches, Navigate children

**Resource:**
- Purpose: Wrapper around root Complex property representing a SCIM resource instance
- Location: `.legacy/pkg/v2/prop/resource.go`
- Pattern: Holds ResourceType reference, provides Navigator entry point, Hash for comparison, Visit for traversal
- Used for: All service operations, pass through create/read/update/delete flows

**Schema & Attribute:**
- Purpose: Define data contracts; Attribute describes single field, Schema describes complex type
- Location: `.legacy/pkg/v2/spec/`
- Pattern: Parsed from JSON (internal/attribute.go, internal/schema.go adapters), read-only accessors for field metadata
- Properties: Type, MultiValued, Required, Mutability, Returned, Uniqueness, SubAttributes, Annotations

**Filter & Expression:**
- Purpose: SCIM filter predicates parsed into tree for evaluation and DB translation
- Location: `.legacy/pkg/v2/crud/expr/`
- Pattern: expression.go defines Expression tree, filter.go evaluates on Property, path.go parses SCIM paths
- Used for: db.Query filters, sort criteria, service-side filtering

**Navigator:**
- Purpose: Maintains call stack during property traversal to enable event propagation
- Location: `.legacy/pkg/v2/prop/navigator.go`
- Pattern: Stack-based (stack: []Property), fluent API (returns Navigator for chaining)
- Key methods: Dot (navigate child), At (index), Where (criteria), Add/Replace/Delete (with propagation), Retract (backtrack)

**Subscriber & Event:**
- Purpose: Reactive processing of property mutations
- Location: `.legacy/pkg/v2/prop/subscriber.go`, `.legacy/pkg/v2/prop/event.go`
- Pattern: SubscriberFactory registers annotation-based handlers, Properties notify subscribers on mutations
- Examples: AutoCompactSubscriber (removes empty multiValued elements), custom validators
- Event types: EventAssigned (value added), EventUnassigned (value deleted)

**Facade:**
- Purpose: Struct-based domain object ↔ Resource conversion
- Location: `.legacy/pkg/v2/facade/`
- Pattern: Struct fields tagged with `scim:"path"`, Export/Import use reflection to map fields to paths
- Support: Nested paths, filtered paths (e.g., `emails[type eq "work"].value`), multi-valued slices

## Entry Points

**HTTP API Server (.legacy/cmd/api):**
- Location: `.legacy/bootstrap.go` → `.legacy/cmd/api/cmd.go` → `Command()`
- Triggers: `go run bootstrap.go api --port 8080 --mongodb-uri=...`
- Responsibilities: 
  - Parse CLI flags, initialize application context (logger, MongoDB, schemas, services)
  - Set up HTTP router with GET/POST/PUT/PATCH/DELETE endpoints for /Users and /Groups
  - Listen on port, delegate requests to handlers

**GroupSync Consumer (.legacy/cmd/groupsync):**
- Location: `.legacy/bootstrap.go` → `.legacy/cmd/groupsync/cmd.go` → `Command()`
- Triggers: `go run bootstrap.go group-sync --mongodb-uri=... --rabbitmq-uri=...`
- Responsibilities:
  - Connect to RabbitMQ, start message consumer
  - Listen for group mutation events, synchronize group membership
  - Update user resources with group references

**Bootstrap Initialization (.legacy/bootstrap.go):**
- Location: Main entry point
- Triggers: Executes when binary runs
- Responsibilities:
  - Creates CLI app with two commands (api, groupsync)
  - Routes to appropriate command handler
  - Handles shutdown

## Error Handling

**Strategy:** Spec-defined error types with HTTP status code mapping

**Patterns:**
- Custom error wrapping: `fmt.Errorf("%w: message", spec.ErrType)` (spec.go defines ErrInternal, ErrNotFound, ErrConflict, ErrInvalidPath, ErrInvalidSyntax, etc.)
- Handler error serialization: `handlerutil.WriteError()` inspects error type and writes HTTP response (400 Bad Request for invalid, 404 Not Found, 409 Conflict, 500 Internal)
- Validation errors: Collected by ValidationFilter in service pipeline, returned in create/replace/patch flows
- Database errors: Propagated from db.DB impl (e.g., MongoDB connection, index creation failures)

## Cross-Cutting Concerns

**Logging:** 
- Package: `github.com/rs/zerolog`
- Entry point: `args.Logger()` in `.legacy/cmd/internal/args/logger.go` creates logger from CLI flags
- Usage: Injected to handlers, services, adapters via context

**Validation:**
- Approach: Two-tier (spec-based + domain-based)
  - Spec-based: Enforced during json.Deserialize (type matching, required fields)
  - Domain-based: service/filter/validation.go ValidationFilter checks uniqueness, format constraints
- Trigger: Create/Replace/Patch service pipelines

**Authentication:**
- Current: Not detected in legacy codebase
- Infrastructure: Assumed handled by reverse proxy in production (not in SCIM library scope)

**Schema Registration:**
- Approach: Config-driven from JSON files
- Location: `.legacy/cmd/internal/args/scim.go` parses schema files and registers with spec package
- Timing: Done once on startup via sync.Once in applicationContext.ensureSchemaRegistered()

---

*Architecture analysis: 2026-05-07*
