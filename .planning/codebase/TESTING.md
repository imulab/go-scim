# Testing Patterns

**Analysis Date:** 2026-05-07

## Test Framework

**Runner:**
- `testing` (Go standard library)
- `github.com/stretchr/testify` for assertions and test suites

**Assertion Library:**
- `github.com/stretchr/testify/assert` for assertion helpers
- `github.com/stretchr/testify/require` for assertions that halt on failure
- `github.com/stretchr/testify/suite` for test suite support

**Run Commands:**
```bash
go test ./...              # Run all tests
go test -v ./...           # Run all tests with verbose output
go test -cover ./...       # Run tests with coverage
go test -run TestXyz ./... # Run specific test
```

## Test File Organization

**Location:**
- Co-located with implementation: `property.go` and `property_test.go` in same directory
- Pattern applies across all packages: `pkg/v2/prop/`, `pkg/v2/service/`, `mongo/v2/`, `cmd/api/`

**Naming:**
- Test files: `{name}_test.go`
- Test functions: `TestXxx()` where `Xxx` is descriptive (e.g., `TestIntegerProperty`, `TestNavigator`)
- Test suites: `{Descriptive}TestSuite` struct (e.g., `IntegerPropertyTestSuite`, `DeleteServiceTestSuite`)

**Structure:**
```
property.go              # Implementation
property_test.go         # Tests
│
├── TestXxx(t *testing.T)        # Entry point test function
├── type XxxTestSuite struct     # Suite struct
│   ├── suite.Suite              # Embedded testify suite
│   └── shared/reusable fields
└── func (s *XxxTestSuite) TestYyy()  # Test methods
```

## Test Structure

**Suite Organization:**
```go
type IntegerPropertyTestSuite struct {
	suite.Suite
	PropertyTestSuite    // Embedded for shared test methods
	OperatorTestSuite    // Embedded for operator tests
	standardAttr *spec.Attribute
}

func TestIntegerProperty(t *testing.T) {
	s := new(IntegerPropertyTestSuite)
	s.NewFunc = NewInteger              // Setup constructor
	s.NewOfFunc = func(attr *spec.Attribute, v interface{}) Property {
		return NewIntegerOf(attr, v.(int64))
	}
	suite.Run(t, s)
}
```

**Patterns:**
- Setup: `SetupSuite()` runs once before all tests in suite; `SetupTest()` runs before each test
- Teardown: `TeardownSuite()` runs once after suite; `TeardownTest()` after each test
- Assertions: `assert.Equal(t, expected, actual)`, `require.Nil(t, err)`
- Conditional checks: `assert.False(t, expect)` when interface not implemented

**Test Structure Example:**
```go
func (s *IntegerPropertyTestSuite) TestNew() {
	tests := []struct {
		name        string
		description string
		before      func()              // Setup before test
		attr        *spec.Attribute
		expect      func(t *testing.T, p Property)  // Assertion function
	}{
		{
			name: "new integer of integer attribute",
			attr: s.standardAttr,
			expect: func(t *testing.T, p Property) {
				assert.Equal(t, "...", p.Attribute().ID())
				assert.Equal(t, spec.TypeInteger, p.Attribute().Type())
				assert.True(t, p.IsUnassigned())
			},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			if test.before != nil {
				test.before()
			}
			test.expect(t, s.NewFunc(test.attr))
		})
	}
}
```

## Mocking

**Framework:** No explicit mocking library (gomock, mockery)
- Test doubles created manually
- In-memory implementations where needed: `db.Memory()`

**Patterns:**
```go
// In-memory database for testing
database := db.Memory()
err := database.Insert(context.TODO(), resource)

// Setup with dependency injection
setup: func(t *testing.T) Delete {
	database := db.Memory()
	return DeleteService(&spec.ServiceProviderConfig{}, database)
}

// Subscriber test double
type DummySubscriber struct{}
func (s *DummySubscriber) Notify(p Property, events *Events) error {
	return nil
}
```

**What to Mock:**
- Database: Use `db.Memory()` for in-memory test double
- Services: Inject concrete implementations, no mocks
- Properties: Create real properties with test attributes

**What NOT to Mock:**
- Property system: Always use real `Property` implementations
- Attributes: Always create real `*spec.Attribute` via JSON deserialization
- SCIM types: No mocking of spec types; use real definitions

## Fixtures and Factories

**Test Data:**
```go
// Load attribute from JSON
attr := new(spec.Attribute)
require.Nil(t, json.Unmarshal([]byte(`{
  "id": "...",
  "name": "userName",
  "type": "string",
  "multiValued": false,
  ...
}`), attr))

// Property with value
p := NewStringOf(attr, "value")

// Resource with fields
resource := s.resourceOf(t, map[string]interface{}{
	"id": "user001",
	"userName": "user001",
})
```

**Location:**
- Fixtures embedded in test files as structs or inline JSON
- Reusable test data defined in test suite structs
- Helper methods on suite: `s.mustAttribute(t, reader)`, `s.resourceOf(t, data)`
- JSON test data loaded from files in integration tests (e.g., `../../public/schemas`)

## Coverage

**Requirements:** No explicit coverage requirements enforced
- Code coverage not mentioned in build config
- Tests focus on critical paths and edge cases

**View Coverage:**
```bash
go test -cover ./...                           # Summary
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out               # Visual report
```

## Test Types

**Unit Tests:**
- Scope: Individual property types, services, operations
- Approach: Table-driven tests with multiple cases per test method
- Example: `TestIntegerProperty` covers New, Raw, Unassigned, Dirty, Clone, Add, Replace, Delete, EqualTo, GreaterThan
- Properties tested: Basic behavior, edge cases, state transitions

**Integration Tests:**
- Scope: Database operations, service chains, filter/transform pipeline
- Approach: Docker-based test containers for MongoDB and RabbitMQ
- Example: `MongoDatabaseTestSuite.TestQuery` tests actual database queries
- Example: `CommandTestSuite.TestCommand` tests full API startup with real services

**E2E Tests:**
- Framework: Docker containers for dependencies
- Setup: Uses `dockertest` to spin up MongoDB, RabbitMQ
- Cleanup: Handles resource cleanup after tests
- Example: `cmd/api/cmd_test.go` validates full application startup

## Common Patterns

**Async Testing:**
```go
// Testing goroutines and channels
go func() {
	err := app.Run([]string{"scim", "api", ...})
	assert.Nil(s.T(), err)
}()

// Retry pattern for async operations
err := backoff.Retry(func() error {
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("non-200 status")
	}
	return nil
}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 10))
```

**Error Testing:**
```go
// Check for specific error
expect: func(t *testing.T, err error) {
	assert.NotNil(t, err)
	assert.Equal(t, spec.ErrNotFound, errors.Unwrap(err))
},

// Check error wrapping
if err != nil {
	return nil, fmt.Errorf("%w: additional context", spec.ErrInternal)
}
```

**Visitor Testing:**
```go
// Navigator fluent API tested via chaining
nav := getResource().Navigator()
nav.Dot("emails").At(1).Dot("value")
if err := nav.Error(); err != nil {
	panic(err)
}
println(nav.Current().Raw())
```

**Event Testing:**
```go
// Events emitted and checked during modifications
event, err := property.Replace(newValue)
assert.Nil(t, err)
assert.NotNil(t, event)
assert.Equal(t, EventAssigned, event.Type())
```

## Table-Driven Test Pattern (Core Pattern)

All major test suites use table-driven tests with `[]struct{ name string, ... }`:

```go
tests := []struct {
	name        string      // Test case name
	description string      // Optional detailed description
	before      func()       // Setup function
	input       interface{}  // Input data
	expect      func(t *testing.T, result interface{})  // Assertion callback
}{
	{
		name: "descriptive test case",
		before: func() { /* setup */ },
		input: "test value",
		expect: func(t *testing.T, result interface{}) {
			assert.Equal(t, "expected", result)
		},
	},
}

for _, test := range tests {
	s.T().Run(test.name, func(t *testing.T) {
		if test.before != nil {
			test.before()
		}
		result := operation(test.input)
		test.expect(t, result)
	})
}
```

**Why used:** Allows testing many cases in one test method, clear case names, easy to add new cases, assertion logic isolated in callback

## Docker-based Integration Tests

**Setup pattern:**
```go
func (s *MongoDatabaseTestSuite) SetupSuite() {
	var err error
	s.dockerPool, err = dockertest.NewPool(testDockerEndpoint)
	s.dockerResource, err = s.dockerPool.Run(
		testMongoImageName,
		testMongoImageTag,
		[]string{...},
	)
	// Wait for container to be ready
	backoff.Retry(func() error {
		return s.mongoReadiness()
	}, ...)
}

func (s *MongoDatabaseTestSuite) TeardownSuite() {
	s.dockerPool.Purge(s.dockerResource)
}
```

**Test Count:** 47 `_test.go` files in `.legacy/` codebase

## Key Test Packages

**Major test suites by location:**
- `pkg/v2/prop/` - 14+ test files covering all property types and navigation
- `pkg/v2/service/` - 6+ test files for CRUD services
- `pkg/v2/json/` - 2 test files for serialization
- `pkg/v2/spec/` - 3 test files for schemas and attributes
- `mongo/v2/` - 3 test files for MongoDB database operations
- `cmd/api/` - 1 test file for command/server startup

---

*Testing analysis: 2026-05-07*
