# Coding Conventions

**Analysis Date:** 2026-05-07

## Naming Patterns

**Files:**
- Singular names for most files: `property.go`, `subscriber.go`, `navigator.go`, `event.go`
- Test files use `_test.go` suffix: `property_test.go`, `integer_test.go`
- Implementation structures are unexported (lowercase): `complexProperty`, `binaryProperty`, `multiValuedProperty`
- Public API uses exported names: `Property`, `Navigator`, `Subscriber`

**Functions:**
- Constructor functions use `New` prefix: `NewProperty()`, `NewString()`, `NewComplex()`, `NewBinary()`
- Constructor functions with initial values use `Of` suffix: `NewStringOf()`, `NewComplexOf()`, `NewIntegerOf()`
- Factory functions also use `New` pattern: `SubscriberFactory()` returns factory singleton, factory methods use `Create()`
- Service constructors use pattern `{Name}Service()`: `CreateService()`, `DeleteService()`, `GetService()`, `QueryService()`
- Private/helper functions use snake_case or camelCase: `ensureSingularComplexType()`, `parseResource()`
- Test methods follow Go convention: `TestXxx()`, `SetupSuite()`, `SetupTest()`, `TeardownTest()`

**Variables:**
- Interface types are PascalCase (singular): `Property`, `Navigator`, `Subscriber`, `Visitor`
- Capability interfaces append `Capable`: `EqCapable`, `SwCapable`, `EwCapable`, `CoCapable`, `GtCapable`, `PrCapable`
- Receiver variables use `s` for suite, `p` for property, `attr` for attribute
- Loop variables: `index`, `child` for iteration functions; `i` for indices in some contexts

**Types:**
- Interfaces are exported and minimal: `Property`, `Navigator`, `Visitor`, `Subscriber`
- Unexported structs that implement interfaces: `complexProperty`, `stringProperty`, `binaryProperty`, `defaultNavigator`
- Public type groups use `(...)` blocks for related types: `type ( Create interface {...}, CreateRequest struct {...}, CreateResponse struct {...} )`
- Enums are `const` with `iota`: `EventType` with `EventAssigned`, `EventDeleted`, `EventCompacted`

## Code Style

**Formatting:**
- Standard Go formatting (implicit via gofmt)
- 4-space indentation (Go standard)
- Line wrapping at reasonable width

**Linting:**
- Code follows Go conventions and idioms
- Error handling is explicit and pervasive
- No empty interface{} parameters unless absolutely necessary

## Import Organization

**Order:**
1. Standard library imports: `context`, `fmt`, `encoding/json`, `io`, `strings`, `hash/fnv`
2. External packages: `github.com/stretchr/testify/...`, `go.mongodb.org/mongo-driver`, `github.com/ory/dockertest`
3. Internal packages: `github.com/imulab/go-scim/pkg/v2/...`

**Path Aliases:**
- Import paths use full canonical paths
- No path aliases observed
- Relative imports never used

## Error Handling

**Patterns:**
- Custom error types defined in `pkg/v2/spec/error.go` as error prototypes
- Error variables (singletons): `ErrInvalidFilter`, `ErrTooMany`, `ErrUniqueness`, `ErrMutability`, `ErrInvalidSyntax`, `ErrInvalidPath`, `ErrNoTarget`, `ErrInvalidValue`, `ErrNotFound`, `ErrSensitive`, `ErrConflict`, `ErrInternal`
- Error type structure: `Error` with `Status` (HTTP code) and `Type` (error name string)
- Error creation: Use predefined error prototypes, wrap with `fmt.Errorf("%w", err)` for context
- Panic used for programming errors (invalid attribute types): `panic("invalid attribute for complex property")`
- Error checking is explicit: `if err != nil { return ... }`
- Navigator pattern defers error checking: stateful navigation returns error only via `Error()` method

**Error File Location:**
- `pkg/v2/spec/error.go` contains all SCIM error definitions and `Error` type

## Logging

**Framework:** No explicit logging framework detected; code uses no structured logging
- Design allows consumer to implement logging via interfaces/filters

## Comments

**When to Comment:**
- Package-level comments use `package` directive format: `// This package implements...` in `doc.go` files
- Method receivers documented inline when non-obvious
- Comments explain SCIM-specific behavior and constraints
- Comments explain unusual or important invariants (e.g., "element attribute derived from multiValued")

**JSDoc/TSDoc:**
- Go uses doc comments starting with name: `// Property holds a piece of data...`
- Interface methods are documented: `// Attribute always returns a non-nil attribute...`
- Complex behavior documented: Navigator explanation spans multiple comment lines with detailed semantics

## Function Design

**Size:**
- Median function size is 20-40 lines
- Large functions in test suites (setup/fixtures can be longer)
- Factory/constructor functions are short (5-15 lines)

**Parameters:**
- Constructor methods typically take `*spec.Attribute` and context parameters
- Service methods take `context.Context` as first parameter (Go standard)
- Request/response structs group related parameters
- Callbacks/functions use concrete types, not `interface{}`

**Return Values:**
- Public APIs return interfaces: `Navigate(p Property) Navigator`, `NewProperty() Property`
- Error as last return value: `(resp *CreateResponse, err error)`
- Single-value returns for factories when error semantics clear from context

## Module Design

**Exports:**
- Public interfaces exported: `Property`, `Navigator`, `Subscriber`, `Visitor`
- Public functions exported: `NewProperty()`, `Navigate()`, `Visit()`, `SubscriberFactory()`
- Implementation details unexported: `complexProperty` struct not exported
- Service constructors exported: `CreateService()`, `DeleteService()`
- Public request/response types: `CreateRequest`, `CreateResponse`, `DeleteRequest`

**Barrel Files:**
- No barrel files (`index.go`) observed
- Each package focused on specific domain (prop, spec, service, etc.)
- `doc.go` files used for package documentation

**Package Layering:**
- `pkg/v2/spec/` - SCIM specifications, attributes, resource types, error definitions
- `pkg/v2/prop/` - Property implementations, navigator, visitor pattern, events
- `pkg/v2/service/` - Business logic services (CRUD, patch, query)
- `pkg/v2/json/` - JSON serialization/deserialization
- `pkg/v2/crud/` - CRUD abstractions (sort, filter, pagination)
- `pkg/v2/db/` - Database abstraction interface

## Interface Design Patterns

**Core Patterns:**
- **Factory Pattern**: `SubscriberFactory()` singleton with `Register()` and `Create()` methods
- **Visitor Pattern**: `Visitor` interface with `ShouldVisit()`, `Visit()`, `BeginChildren()`, `EndChildren()`
- **Observer/Subscriber Pattern**: Properties notify subscribers of changes via `Subscriber.Notify()`
- **Capability Interfaces**: Fine-grained operation capabilities: `EqCapable`, `SwCapable`, `CoCapable`, etc.
- **Navigator Pattern**: Stateful traversal with fluent API returning `Navigator` for chaining

**Property System:**
- `Property` interface is the core abstraction
- Properties hold SCIM attribute metadata via `Attribute()` method
- Child access via `CountChildren()`, `ForEachChild()`, `FindChild()`, `ChildAtIndex()`
- Modification tracked via `Dirty()` and `Dirty` events
- Events propagated upstream via `Navigator` stateful tracking

**Service Pattern:**
- Interface per service: `Create`, `Get`, `Delete`, `Replace`, `Patch`, `Query`
- Request/Response structs for each service
- Service constructor injects dependencies: `CreateService(resourceType, database, filters)`
- Closed interface implementation pattern (private structs implementing public interfaces)

## Annotations System

**Usage:**
- Attributes carry annotations: `map[string]map[string]interface{}`
- Accessed via `Attribute.ForEachAnnotation(callback)`
- Annotations trigger automatic behavior: `@AutoCompact`, `@Identity`, `@Test`
- `SubscriberFactory` loads subscribers based on annotations during property creation
- Extensible mechanism for custom validators and processors

## Schema and JSON Loading

**Pattern:**
- Attributes loaded from JSON via `json.Unmarshal()` into `Attribute` struct
- JSON fields map to public fields with special adapter handling
- Schemas and resource types loaded from JSON file paths (seen in tests)
- Example path: `public/schemas`, `public/resource_types`
- Deserialization uses `json.Unmarshal()` into `*spec.ResourceType` or `*spec.Schema`

---

*Convention analysis: 2026-05-07*
