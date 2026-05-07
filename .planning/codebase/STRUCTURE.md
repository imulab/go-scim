# Codebase Structure

**Analysis Date:** 2026-05-07

## Directory Layout

```
.legacy/
├── bootstrap.go                 # Entry point: CLI app initialization (api, group-sync commands)
├── go.mod                       # Go module declaration
├── go.sum                       # Dependency checksums
├── Dockerfile                   # Container image definition
├── docker-compose.yml           # Local MongoDB + RabbitMQ stack
├── Makefile                     # Build targets
├── README.md                    # Project documentation
├── cmd/                         # Command implementations (HTTP API, async consumer)
│   ├── api/                     # SCIM HTTP API server
│   │   ├── cmd.go              # CLI command definition, router setup, handler attachment
│   │   ├── handler.go          # HTTP request handlers (Create, Get, Put, Patch, Delete, Search, Health)
│   │   ├── handler_test.go     # Handler integration tests
│   │   ├── service.go          # Service initialization and dependency injection
│   │   ├── context.go          # Application context: lazy-initialized components (logger, DB, services)
│   │   └── args.go             # CLI flag parsing for API server
│   ├── groupsync/              # Group membership sync consumer
│   │   ├── cmd.go              # CLI command, RabbitMQ consumer startup
│   │   ├── args.go             # CLI flag parsing
│   │   ├── context.go          # Application context for consumer
│   │   └── consumer.go         # Message consumption and processing
│   └── internal/               # Shared internal CLI utilities (not exported)
│       ├── args/               # Reusable CLI argument packages
│       │   ├── logger.go       # Logger initialization from flags
│       │   ├── database.go     # Memory and MongoDB connection args
│       │   ├── scim.go         # Schema file loading and registration
│       │   └── rabbit.go       # RabbitMQ connection args
│       └── groupsync/          # Group sync shared utilities
│           ├── message.go      # Message type definitions
│           ├── rabbit.go       # RabbitMQ queue/exchange declarations
├── mongo/                       # MongoDB adapter implementation
│   └── v2/                      # SCIM v2 MongoDB persistence layer
│       ├── db.go               # db.DB interface implementation for MongoDB
│       ├── serialize.go        # Resource → BSON document transformation
│       ├── deserialize.go      # BSON document → Resource parsing
│       ├── filter.go           # SCIM filter → MongoDB query operator translation
│       ├── filter_test.go      # Filter translation test cases
│       ├── metadata.go         # Attribute name aliasing (SCIM path → MongoDB field name mapping)
│       ├── index.go            # Index creation for unique/indexed attributes
│       ├── context.go          # MongoDB client initialization
│       ├── doc.go              # Package documentation
│       ├── serialize_test.go   # Serialization tests
│       ├── deserialize_test.go # Deserialization tests
│       └── db_test.go          # Database operation tests
├── pkg/v2/                      # Core SCIM v2 library (main business logic)
│   ├── annotation/              # Metadata annotation system
│   │   └── annotation.go        # Annotation constant definitions (@ReadOnly, @Identity, @MongoIndex, etc.)
│   ├── spec/                    # SCIM specification definitions
│   │   ├── attribute.go        # Attribute struct and accessor methods
│   │   ├── schema.go           # Schema struct (collection of attributes)
│   │   ├── resource_type.go    # ResourceType (schema + SCIM metadata)
│   │   ├── type.go             # Type enum (string, integer, boolean, decimal, datetime, reference, binary, complex)
│   │   ├── mutability.go       # Mutability enum (readOnly, readWrite, writeOnly, immutable)
│   │   ├── returned.go         # Returned enum (always, never, default, request)
│   │   ├── uniqueness.go       # Uniqueness enum (none, server, global)
│   │   ├── constant.go         # SCIM constant URNs and resource type names
│   │   ├── error.go            # Error type definitions (ErrNotFound, ErrConflict, ErrInvalidPath, etc.)
│   │   ├── config.go           # ServiceProviderConfig parsing
│   │   ├── meta.go             # Meta attribute definitions
│   │   ├── internal/           # Internal spec adapters (not exported)
│   │   │   ├── attribute.go    # JSON → Attribute deserialization
│   │   │   └── schema.go       # JSON → Schema deserialization
│   │   ├── attribute_test.go   # Attribute tests
│   │   ├── schema_test.go      # Schema tests
│   │   ├── resource_type_test.go # ResourceType tests
│   │   └── doc.go              # Package documentation
│   ├── prop/                    # Property tree model (in-memory resource representation)
│   │   ├── property.go         # Property interface (core abstraction)
│   │   ├── resource.go         # Resource wrapper around root property
│   │   ├── navigator.go        # Navigator for traversal with event propagation
│   │   ├── factory.go          # Property factory function (NewString, NewComplex, etc.)
│   │   ├── string.go           # String property implementation
│   │   ├── integer.go          # Integer property implementation
│   │   ├── boolean.go          # Boolean property implementation
│   │   ├── datetime.go         # DateTime property implementation
│   │   ├── decimal.go          # Decimal property implementation
│   │   ├── reference.go        # Reference (URI) property implementation
│   │   ├── binary.go           # Binary property implementation
│   │   ├── complex.go          # Complex (object) property implementation
│   │   ├── multi.go            # MultiValued (array) property implementation
│   │   ├── event.go            # Event type definitions and Events container
│   │   ├── subscriber.go       # Subscriber interface and built-in subscribers (AutoCompact)
│   │   ├── op.go               # Operation result wrapper (not commonly used)
│   │   ├── visit.go            # Visitor pattern for tree traversal (DFS)
│   │   ├── *_test.go           # Property type implementation tests (string_test.go, integer_test.go, etc.)
│   │   └── doc.go              # Package documentation
│   ├── crud/                    # CRUD operations and query expressions
│   │   ├── crud.go             # Add, Replace, Delete functions (high-level mutations)
│   │   ├── eval.go             # Filter expression evaluation against property tree
│   │   ├── query.go            # Query parameters (Filter, Sort, Pagination, Projection)
│   │   ├── traverse.go         # Property tree traversal engine
│   │   ├── register.go         # Resource type registration for path compilation caching
│   │   ├── sort_by.go          # Sort criteria evaluation
│   │   ├── expr/               # Expression compilation and evaluation
│   │   │   ├── expression.go   # Expression tree node definition
│   │   │   ├── filter.go       # Filter expression parser and evaluator
│   │   │   ├── path.go         # SCIM path expression parser
│   │   │   ├── constant.go     # Expression constants
│   │   │   ├── urn.go          # URN parsing utilities
│   │   │   ├── filter_test.go  # Filter parsing tests
│   │   │   └── path_test.go    # Path parsing tests
│   │   ├── *_test.go           # CRUD operation tests
│   │   └── doc.go              # Package documentation
│   ├── db/                      # Database adapter interface
│   │   ├── db.go               # DB interface (Insert, Get, Replace, Delete, Query, Count)
│   │   ├── memory.go           # In-memory implementation (for testing/demo)
│   │   ├── noop.go             # No-op implementation (for testing)
│   │   └── doc.go              # Package documentation
│   ├── json/                    # JSON serialization/deserialization
│   │   ├── deserialize.go      # JSON bytes → Resource (with schema validation)
│   │   ├── serialize.go        # Resource → JSON bytes (with projection filtering)
│   │   ├── adapt.go            # JSON adapter utilities
│   │   ├── options.go          # Serialization options (Include, Exclude attributes)
│   │   ├── scanner.go          # JSON scanner state machine (custom parser)
│   │   ├── safe.go             # Safe serialization (ensures returned=never not leaked)
│   │   ├── internal/           # Internal JSON adapters (not exported)
│   │   │   ├── resource_type.go # ResourceType JSON marshaling
│   │   │   └── schema.go       # Schema JSON marshaling
│   │   ├── *_test.go           # JSON serialization tests
│   │   └── doc.go              # Package documentation
│   ├── service/                 # SCIM RFC7644 endpoint implementation
│   │   ├── create.go           # Create resource service (POST)
│   │   ├── get.go              # Get resource service (GET)
│   │   ├── replace.go          # Replace resource service (PUT)
│   │   ├── patch.go            # Patch resource service (PATCH)
│   │   ├── delete.go           # Delete resource service (DELETE)
│   │   ├── query.go            # Search/query resources service (GET with filter/sort)
│   │   ├── filter/             # Service request filters (pipeline stage implementations)
│   │   │   ├── filter.go       # Filter interface and composition
│   │   │   ├── readonly.go     # ReadOnlyFilter: reject mutations on read-only attributes
│   │   │   ├── uuid.go         # UUIDFilter: auto-generate id if missing
│   │   │   ├── bcrypt.go       # BCryptFilter: hash password attributes
│   │   │   ├── meta.go         # MetaFilter: set created/lastModified timestamps
│   │   │   ├── validation.go   # ValidationFilter: check uniqueness constraints
│   │   │   ├── visit.go        # Filter evaluation helpers
│   │   │   ├── *_test.go       # Filter tests
│   │   │   └── doc.go          # Package documentation
│   │   ├── *_test.go           # Service operation tests
│   │   └── doc.go              # Package documentation
│   ├── handlerutil/             # HTTP handler utilities
│   │   ├── request.go          # Parse HTTP requests to service requests
│   │   ├── response.go         # Serialize service responses to HTTP
│   │   ├── *_test.go           # Handler utility tests
│   │   └── doc.go              # Package documentation
│   ├── facade/                  # Struct ↔ Resource conversion
│   │   ├── export.go           # Convert struct → Resource (domain object mapping)
│   │   ├── import.go           # Convert Resource → struct (domain object deserialization)
│   │   ├── support.go          # Type conversion and validation utilities
│   │   ├── internal/           # Internal facade helpers (not exported)
│   │   │   ├── reflect.go      # Reflection utilities for struct inspection
│   │   │   └── slice.go        # Slice type utilities
│   │   ├── facade_test.go      # Facade tests
│   │   └── doc.go              # Package documentation
│   ├── groupsync/              # Group membership synchronization
│   │   ├── sync.go             # Synchronization logic (apply diff to group members)
│   │   ├── diff.go             # Group membership diff detection
│   │   ├── *_test.go           # GroupSync tests
│   │   └── doc.go              # Package documentation
│   └── annotation/             # (Note: already listed above)
├── asset/                       # Static assets
│   └── schema/                  # SCIM schema JSON files
│       ├── ...                  # Core schemas, extension schemas
└── public/                      # Public assets (Swagger UI, etc.)
```

## Directory Purposes

**`.legacy/cmd/api/`:**
- Purpose: HTTP API server command entry point
- Contains: CLI flag parsing, application context factory, HTTP route handlers, service wiring
- Key files: `cmd.go` (command def + router), `context.go` (lazy-init components), `handler.go` (request handlers)

**`.legacy/cmd/groupsync/`:**
- Purpose: Async group membership sync consumer
- Contains: RabbitMQ consumer setup, message processing, group sync service calls
- Key files: `cmd.go` (command def), `context.go` (MongoDB/RabbitMQ setup), `consumer.go` (message loop)

**`.legacy/cmd/internal/args/`:**
- Purpose: Reusable CLI argument packages shared between api and groupsync commands
- Contains: Logger, MongoDB, RabbitMQ, SCIM schema file argument handlers
- Key files: `logger.go` (zerolog setup), `database.go` (MongoDB URI parsing), `scim.go` (schema loading)

**`.legacy/mongo/v2/`:**
- Purpose: Concrete db.DB implementation for MongoDB collections
- Contains: BSON serialization, filter translation, index management
- Key files: `db.go` (DB interface impl), `serialize.go` (Resource → BSON), `filter.go` (SCIM filter → Mongo query)

**`.legacy/pkg/v2/spec/`:**
- Purpose: SCIM specification type system (schemas, attributes, resource types)
- Contains: Read-only data structures that define all SCIM contracts
- Key files: `attribute.go` (Attribute), `schema.go` (Schema), `resource_type.go` (ResourceType), errors, mutability/returned/uniqueness enums

**`.legacy/pkg/v2/prop/`:**
- Purpose: In-memory tree-based property model (each node is schema-aware)
- Contains: Property interface + 10 implementations, Navigator, Events, Subscribers, Resource wrapper
- Key files: `property.go` (interface), `resource.go` (Resource), `navigator.go` (tree traversal), all type files (string.go, complex.go, etc.)

**`.legacy/pkg/v2/crud/`:**
- Purpose: Path-based mutations and query expression compilation
- Contains: Add/Replace/Delete functions, filter/sort/projection evaluation, expression compiler
- Key files: `crud.go` (Add/Replace/Delete), `eval.go` (filter evaluation), `expr/filter.go` (filter parsing), `expr/path.go` (path parsing)

**`.legacy/pkg/v2/db/`:**
- Purpose: Persistence layer abstraction
- Contains: DB interface defining contract, in-memory impl for testing
- Key file: `db.go` (DB interface)

**`.legacy/pkg/v2/json/`:**
- Purpose: JSON ↔ Resource conversion with schema validation
- Contains: Custom JSON parser (scanner.go), deserialization (deserialize.go), serialization (serialize.go)
- Key files: `deserialize.go` (JSON → Resource), `serialize.go` (Resource → JSON)

**`.legacy/pkg/v2/service/`:**
- Purpose: RFC7644 SCIM endpoint implementation
- Contains: Create/Get/Replace/Patch/Delete/Query services, filter pipeline
- Key files: Service per operation (create.go, get.go, patch.go), filter/validation.go (constraint checking)

**`.legacy/pkg/v2/facade/`:**
- Purpose: Struct-based domain object ↔ Resource mapping
- Contains: Reflection-based field → SCIM path mapper
- Key files: `export.go` (struct → Resource), `import.go` (Resource → struct)

## Key File Locations

**Entry Points:**
- `/.legacy/bootstrap.go`: Main binary entry point (CLI app routing)
- `.legacy/cmd/api/cmd.go`: HTTP API server initialization
- `.legacy/cmd/groupsync/cmd.go`: Group sync consumer initialization

**Configuration & Setup:**
- `.legacy/cmd/internal/args/scim.go`: Schema file loading and spec registration
- `.legacy/cmd/api/context.go`: Application context (lazy-init of logger, DB, services)
- `.legacy/mongo/v2/metadata.go`: MongoDB field name aliases and index annotations

**Core Logic:**
- `.legacy/pkg/v2/spec/attribute.go`: SCIM Attribute (schema contract)
- `.legacy/pkg/v2/prop/property.go`: Property interface (data node in tree)
- `.legacy/pkg/v2/prop/resource.go`: Resource (wrapper around root property)
- `.legacy/pkg/v2/prop/navigator.go`: Navigator (tree traversal with event propagation)
- `.legacy/pkg/v2/crud/crud.go`: Add/Replace/Delete (high-level mutations)
- `.legacy/pkg/v2/crud/expr/filter.go`: Filter expression parsing
- `.legacy/pkg/v2/crud/expr/path.go`: SCIM path parsing
- `.legacy/pkg/v2/service/create.go`: Create service implementation
- `.legacy/pkg/v2/json/deserialize.go`: JSON → Resource deserialization
- `.legacy/pkg/v2/json/serialize.go`: Resource → JSON serialization

**JSON Serialization:**
- `.legacy/pkg/v2/json/deserialize.go`: Parse JSON bytes into property tree
- `.legacy/pkg/v2/json/serialize.go`: Render property tree to JSON bytes
- `.legacy/pkg/v2/json/scanner.go`: Custom JSON parser state machine

**Database Adapter:**
- `.legacy/pkg/v2/db/db.go`: DB interface contract
- `.legacy/mongo/v2/db.go`: MongoDB collection-based implementation
- `.legacy/mongo/v2/serialize.go`: Resource → BSON document
- `.legacy/mongo/v2/deserialize.go`: BSON document → Resource
- `.legacy/mongo/v2/filter.go`: SCIM filter → MongoDB query translation
- `.legacy/mongo/v2/metadata.go`: Field name mapping (SCIM path → MongoDB field)

**Testing:**
- `.legacy/pkg/v2/prop/property_test.go`: Property test base suite
- `.legacy/pkg/v2/prop/string_test.go`: String property tests (pattern for other types)
- `.legacy/pkg/v2/service/create_test.go`: Create service integration tests
- `.legacy/mongo/v2/db_test.go`: MongoDB adapter tests

## Naming Conventions

**Files:**
- Pattern: `lowercase_with_underscores.go` for source, `lowercase_with_underscores_test.go` for tests
- Examples: `property.go`, `property_test.go`, `deserialize.go`, `deserialize_test.go`
- Convention: Test file pairs (one-to-one with source files in same package)

**Packages:**
- Doc file: `doc.go` in each package root provides package-level godoc
- Pattern: `.legacy/pkg/v2/{spec,prop,crud,json,service,db,facade}/doc.go` present in all major packages
- Internal packages: `.legacy/pkg/v2/{spec,json,facade}/internal/` hide non-public adapters

**Functions:**
- Pattern: CamelCase (NewProperty, Deserialize, Add, Replace, Delete)
- Factory functions: `New{Type}` (NewString, NewComplex, NewMulti, NewResource)
- Service factories: `{Operation}Service` (CreateService, GetService, QueryService)
- Interface implementations: Usually unexported concrete structs (e.g., `stringProperty` impl of `Property`)

**Types:**
- Pattern: Exported interfaces as PascalCase (Property, Navigator, DB)
- Pattern: Unexported concrete types as camelCase with suffix (stringProperty, complexProperty, mongoDB)
- Pattern: Request/Response types as PascalCase with suffix (CreateRequest, CreateResponse)

**Test Patterns:**
- Test suite naming: `{Type}TestSuite` (StringPropertyTestSuite, BinaryPropertyTestSuite)
- Test function naming: `Test{Feature}` (TestNew, TestAdd, TestReplace)
- Testify pattern: Uses `suite.Suite` for shared setup/teardown, `assert.Equal`, `assert.Error`
- Example: `.legacy/pkg/v2/prop/string_test.go` embeds `suite.Suite` and `PropertyTestSuite`

**Directories:**
- Pattern: Lowercase, no underscores (cmd, api, groupsync, mongo, pkg, v2, spec, prop, crud, db, json, service)
- Command packages: `cmd/{command-name}/` (api, groupsync)
- Internal packages: `internal/` for non-exported packages
- Subpackages: `.legacy/pkg/v2/{spec,prop,crud,service}/internal/` for internal adapters
- Version packages: `v2/` for SCIM v2 (future-proof for v3+)

## Where to Add New Code

**New SCIM Attribute Type:**
- Primary code: `.legacy/pkg/v2/prop/{typename}.go` (e.g., uuid.go for UUID type)
  - Implement Property interface
  - Provide NewTypeName and NewTypeNameOf factory functions
  - Provide Type() method returning spec.Type
- Tests: `.legacy/pkg/v2/prop/{typename}_test.go`
  - Create Test{TypeName} function
  - Embed PropertyTestSuite for common test patterns
  - Test Add/Replace/Delete, Raw() values, Matches(), Clone()

**New Service Filter:**
- Implementation: `.legacy/pkg/v2/service/filter/{filter_name}.go`
  - Implement `ByResource` interface
  - Add to filter pipeline in `.legacy/cmd/api/context.go` (UserCreateService, etc.)
- Tests: `.legacy/pkg/v2/service/filter/{filter_name}_test.go`

**New Database Adapter (e.g., PostgreSQL):**
- Implementation: `.legacy/sql/v2/db.go`
  - Implement db.DB interface (Insert, Get, Replace, Delete, Query, Count)
  - Handle Filter → SQL translation (parse crud.expr.Filter to WHERE clause)
  - Handle resource serialization/deserialization
  - Handle projection (column selection)
- Tests: `.legacy/sql/v2/db_test.go`
- Metadata: `.legacy/sql/v2/metadata.go` (field name mapping if needed)
- Integration: Update `.legacy/cmd/api/context.go` to initialize SQL adapter

**New CLI Command:**
- Structure: `.legacy/cmd/{command_name}/`
  - `cmd.go`: CLI command definition via urfave/cli
  - `context.go`: Application context and lazy initialization
  - `args.go`: CLI flag parsing
  - Supporting files as needed
- Integration: Register command in `.legacy/bootstrap.go` via cli.App.Commands

**New Endpoint:**
- Handler: Add to `.legacy/cmd/api/handler.go` (new handler function)
- Route: Register in `.legacy/cmd/api/cmd.go` (router.{METHOD})
- Service: Add service interface + implementation in `.legacy/pkg/v2/service/` if not existing

**Utility Helpers:**
- Shared logic: `.legacy/pkg/v2/{category}/internal/{file}.go`
- Examples: `.legacy/pkg/v2/spec/internal/attribute.go` (JSON adapter), `.legacy/pkg/v2/json/internal/resource_type.go`

## Special Directories

**`.legacy/asset/schema/`:**
- Purpose: SCIM schema JSON definitions (Core User, Core Group, Enterprise User extension, etc.)
- Generated: No (manually maintained)
- Committed: Yes
- Usage: Loaded by `.legacy/cmd/internal/args/scim.go` on startup

**`.legacy/public/`:**
- Purpose: Swagger/OpenAPI UI and documentation
- Generated: No (static files)
- Committed: Yes
- Usage: Served by HTTP API if enabled

---

*Structure analysis: 2026-05-07*
