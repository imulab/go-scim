# Codebase Concerns

**Analysis Date:** 2026-05-07

---

## 1. Feasibility: Generic Tree Model Incompatible with SQL Relational Stores

**Issue:** The in-memory tree-like data model (prop.Property hierarchy) adapts well to MongoDB's BSON document model but poses fundamental architectural obstacles for SQL relational databases.

**Files:**
- `pkg/v2/prop/property.go` - Property interface defining generic tree structure
- `pkg/v2/prop/resource.go` - Resource as tree of nested properties
- `pkg/v2/db/db.go` - Database abstraction expecting Resource objects
- `mongo/v2/db.go` - MongoDB adapter that serializes tree to BSON
- `mongo/v2/serialize.go` - Visitor-based BSON serialization

**Why It's Hard:**
- Properties are deeply nested in a tree via `complexProperty.subProps` and `multiValuedProperty.elements`. SQL normalization requires flattening into separate tables with foreign keys.
- SCIM's arbitrary nesting (e.g., `emails[type eq "work"].value`) maps naturally to MongoDB dot notation but requires complex join queries in SQL.
- The `mongoPathFor()` method (`.legacy/mongo/v2/db.go:223`) traverses the tree to resolve SCIM paths to MongoDB persistence paths. SQL would need similar path resolution but with recursive JOIN logic.
- Filter evaluation (`pkg/v2/crud/expr/filter.go`) operates on the tree-structured Resource; SQL would require translating SCIM filters to parameterized SQL WHERE clauses with JOIN chains.
- No SQL adapter exists in the codebase; MongoDB is the only concrete persistence implementation.

**Impact:**
- Porting to PostgreSQL, MySQL, or other SQL stores requires a complete rewrite of the persistence layer.
- The abstraction (`pkg/v2/db/db.go`) is thin enough that implementing a SQL adapter would require either redesigning the Resource model or building an ORM translation layer that adds complexity and overhead.

---

## 2. Efficiency: Generic In-Memory Model Duplication with Brittle Rule Enforcement

**Issue:** Maintaining a generic in-memory tree representation duplicates data and complicates implementation of SCIM-specific constraints (e.g., "only one element of object array can be primary").

**Files:**
- `pkg/v2/prop/subscriber.go` - ExclusivePrimarySubscriber (lines 106-180) enforces "only one primary" rule
- `pkg/v2/annotation/annotation.go` - @Primary and @ExclusivePrimary annotations define the rule
- `pkg/v2/prop/multi.go` - multiValuedProperty.Raw() marked "slow operation" (line 53)
- `pkg/v2/prop/multi.go` - Hash() marked "expensive operation" (line 73), uses insertion sort per element (lines 91-94)
- `pkg/v2/json/deserialize.go` - Visitor-based deserialization traverses full tree

**Constraint Implementation Smell:**
The ExclusivePrimarySubscriber (`.legacy/pkg/v2/prop/subscriber.go:106`) enforces SCIM's "only one primary" rule reactively:
- When a primary boolean is set to true, the subscriber walks all siblings to turn off previous primary values.
- This navigation and modification happens during event propagation, adding O(n) work per primary assignment.

**Raw() Performance:**
The Raw() method on multiValuedProperty allocates and appends to a new slice on every call:
```go
func (p *multiValuedProperty) Raw() interface{} {
    if len(p.elements) == 0 { return nil }
    var values []interface{}
    for _, elem := range p.elements {
        values = append(values, elem.Raw())
    }
    return values
}
```
This is inefficient when called repeatedly (e.g., during serialization, filtering, or comparison).

**Hash() Cost:**
Computing hashes for multiValued properties involves:
1. Iterating all elements via ForEachChild callback.
2. Sorting hash values using insertion sort (O(n²) in worst case).
3. No caching mechanism despite being called for comparison operations.

**Impact:**
- Complex resources with large arrays (e.g., 100+ emails, addresses) incur high CPU cost for routine operations.
- The tree traversal cost compounds when resources are visited multiple times (deserialization, filtering, validation, serialization).
- SCIM-specific rules (primary, cardinality, uniqueness) are implemented as afterthoughts via subscribers rather than enforced at the schema/model level.

---

## 3. Portability: No Accommodation for IdP Non-Compliance with SCIM Spec

**Issue:** Mainstream identity providers (Microsoft Entra ID, Okta, etc.) violate SCIM spec by sending attribute names with unexpected casing (e.g., "UserName" instead of "userName"). The codebase provides **no mechanism to accommodate** such deviations.

**Files:**
- `pkg/v2/spec/attribute.go:220` - GoesBy() uses case-insensitive matching:
  ```go
  func (attr *Attribute) GoesBy(name string) bool {
      switch strings.ToLower(name) {
      case strings.ToLower(attr.id), strings.ToLower(attr.path), strings.ToLower(attr.name):
          return true
      }
      return false
  }
  ```
- `pkg/v2/prop/navigator.go:106` - Dot() delegates to ChildAtIndex(name) which uses SubAttributeForName
- `pkg/v2/spec/attribute.go:205` - SubAttributeForName iterates attributes and checks GoesBy()
- `pkg/v2/json/deserialize.go:163` - Deserialization calls `navigator.Dot(attrName)` directly from JSON field names

**Current Behavior:**
- Case-insensitive matching is performed at attribute lookup time (GoesBy), so "UserName", "username", "USERNAME" will all match "userName".
- However, there is **no configuration mechanism** to:
  - Define case-sensitivity policies per attribute.
  - Remap incoming field names to canonical names.
  - Log or warn about non-compliant inputs.
  - Support strict vs. lenient parsing modes.

**Gaps:**
- No middleware or filter to normalize incoming JSON before deserialization.
- No annotation-based configuration to mark attributes as "accept any case" vs. "case-sensitive".
- No error reporting mechanism to distinguish between malformed vs. non-compliant inputs.

**Impact:**
- Codebase is somewhat resilient (case-insensitive matching helps), but fragile if IdPs send structurally different attribute hierarchies.
- No way to enforce strict SCIM compliance or debug compliance issues with partners.
- Future rewrite should include explicit portability layer with configurable normalization rules.

---

## 4. Known Bugs

### 4.1 Projection Bug in MongoDB Adapter

**Issue:** ExcludedAttributes are projected using the wrong source array.

**File:** `mongo/v2/db.go:302-310`

**Bug:**
```go
if len(projection.ExcludedAttributes) > 0 {
    exclude := bson.D{}
    for _, p := range projection.Attributes {  // WRONG: should be projection.ExcludedAttributes
        if mp := d.mongoPathFor(p); len(mp) > 0 {
            exclude = append(exclude, bson.E{Key: mp, Value: 0})
        }
    }
    return exclude
}
```

**Trigger:** Query with exclusion projection (e.g., `attributes=meta` in exclude mode).

**Impact:** Excluded attributes are ignored; all attributes are returned regardless of the exclusion request.

**Workaround:** Currently mitigated by `IgnoreProjection()` option (used in `.legacy/cmd/api/context.go:136, 154`), which bypasses projection entirely.

---

### 4.2 Deprecated io.ioutil Usage

**File:** `pkg/v2/prop/property_test.go:9` imports `io/ioutil`

**Issue:** io.ioutil has been deprecated since Go 1.16; functions should be called directly from io or os packages.

---

## 5. Security Considerations

### 5.1 BCrypt Implementation Timing Vulnerability

**Files:** `pkg/v2/service/filter/bcrypt.go`

**Implementation:**
```go
hashed, err := bcrypt.GenerateFromPassword(raw, cost)
if err != nil {
    return fmt.Errorf("%w: failed to perform bCrypt on attribute '%s'", spec.ErrInternal, attr.Path())
}
_, err = nav.Current().Replace(replacement)
```

**Concern:** While bcrypt.GenerateFromPassword() is cryptographically sound, the error path logs the attribute name. If passwords are hashed, detailed error messages could leak information about which password fields failed.

**Mitigation:** Error is already generic ("failed to perform bCrypt on attribute..."), but the attribute name could be removed for passwords.

---

### 5.2 No Input Validation Before Deserialization

**Files:** `pkg/v2/json/deserialize.go`, `pkg/v2/service/patch.go`

**Concern:** Deserialization does not validate field counts, array sizes, or nesting depth before processing. A malicious client could send:
- Deeply nested JSON to cause stack overflow.
- Large arrays to cause memory exhaustion.
- Invalid UTF-8 sequences in string fields.

**Impact:** Moderate - Go's JSON parser and the manual scanner provide some protection, but resource limits are not enforced.

---

## 6. Performance Bottlenecks

### 6.1 Visitor Traversal Cost

**Files:** `pkg/v2/prop/visit.go`

**Pattern:** Full-resource DFS traversal happens during:
- JSON deserialization (`.legacy/pkg/v2/json/deserialize.go` via manual recursion)
- Serialization to BSON (`.legacy/mongo/v2/serialize.go` visitor)
- Validation (`.legacy/pkg/v2/service/filter/navigate.go`)
- Filter evaluation (`.legacy/pkg/v2/crud/eval.go`)

**Cost:** For a resource with N attributes and M multiValued elements, traversal is O(N + M).

**Cumulative Impact:** If a PATCH operation deserializes, validates, filters, and serializes the same resource, it traverses the tree 4+ times.

---

### 6.2 Slow Array Operations

**File:** `pkg/v2/prop/multi.go:53, 73`

**Code Comments Acknowledge Issues:**
- Raw() marked "slow operation" - rebuilds slice on each call
- Hash() marked "expensive operation" - iterates all elements and sorts hashes

**Impact:** Resources with large multiValued arrays (100+ emails, groups, etc.) pay O(n) cost per Raw() call and O(n²) per Hash() call.

---

### 6.3 Filter Compilation Cost

**File:** `pkg/v2/crud/expr/filter.go:22` (1100 lines)

**Observation:** SCIM filter strings are compiled to ASTs every query. No caching mechanism exists.

**Impact:** Repeated queries with the same filter (e.g., `GET /Users?filter=emails[type eq "work"].value eq "foo"`) recompile the filter AST.

---

## 7. Fragile Areas

### 7.1 JSON Deserializer State Machine

**File:** `pkg/v2/json/deserialize.go` (715 lines)

**Fragility:** Complex state machine with manual offset tracking and scanner state transitions:
- scanNext(), scanWhile() advance offset and opCode
- parseComplexProperty() recursively calls parseFieldName() and parseSingleValuedProperty()
- Error handling relies on state being correct after each step
- Multiple "skip" loops to fast-forward to next object/array boundary

**Risk:** Off-by-one errors, incorrect state transitions, or missing cases in scanner states could silently produce corrupt data.

**Test Coverage:** `json/deserialize_test.go` exists but cannot fully cover state machine complexity.

---

### 7.2 MongoDB Filter Transformation

**File:** `mongo/v2/filter.go` (transformer implementation)

**Concern:** SCIM filter AST must be correctly transformed to BSON filter syntax. A mismatch between SCIM semantics and MongoDB query semantics could produce wrong results.

**Example:** SCIM `filter=emails[primary eq true].value eq "x"` requires:
1. Matching array element with `primary: true`
2. Checking `value` within that element
3. MongoDB `$elemMatch` operator needed for this.

---

### 7.3 Schema Registration Rigidity

**Files:**
- `pkg/v2/spec/schema.go:133-138` - Global schemaRegistry singleton
- `cmd/api/context.go:98-106` - ensureSchemaRegistered() calls panic on failure

**Fragility:**
- Schema registry is a global singleton with no initialization hooks.
- Schemas are registered at startup via args.RegisterSchemas().
- If a schema fails to load, the entire application panics (line 102: `panic(err)`).
- No mechanism to add custom schemas at runtime without modifying startup code.
- No validation that all required schemas are registered before API calls.

**Impact:** Schema changes require full restart; runtime schema updates are impossible.

---

## 8. Test Coverage Gaps

### 8.1 No SQL Adapter Tests

**Gap:** No tests for hypothetical SQL adapter. When rewrite occurs, SQL implementation must be thoroughly tested for:
- JOIN correctness with nested attributes
- Filter translation to WHERE clauses
- Transaction handling for PATCH operations
- Sorting and pagination on calculated fields

---

### 8.2 Limited MongoDB Projection Tests

**Gap:** The projection bug (Section 4.1) suggests insufficient test coverage for Projection operations.

**Files:** `mongo/v2/db_test.go` - No explicit test for ExcludedAttributes projection.

---

### 8.3 Case-Sensitivity Edge Cases

**Gap:** No tests for non-compliant attribute name casing from IdPs.

**Expected Test:** Deserialize JSON with "UserName" (capital U, capital N) and verify it maps to "userName".

---

## 9. Missing Critical Features

### 9.1 No Bulk Operations Support

**Concern:** SCIM v2 spec allows bulk endpoint (/Bulk) for transactional multi-resource operations. Not implemented.

**Impact:** Large-scale migrations or synchronizations must be done via sequential requests, sacrificing performance and atomicity.

---

### 9.2 No Server-Side Filtering Optimization

**Concern:** MongoDB adapter does not use server-side aggregation pipeline for complex filters. All filtering happens in-memory after query.

**Files:** `mongo/v2/db.go:170-211` - Query fetches all matching documents, no aggregation.

---

## 10. Dependencies at Risk

### 10.1 Deprecated io.ioutil in Tests

**Package:** io/ioutil

**Status:** Deprecated since Go 1.16

**Migration:** Replace with os.ReadFile, io.WriteString, etc.

**Files:** `pkg/v2/prop/property_test.go`

---

### 10.2 RabbitMQ Consumer Coupled to Business Logic

**Files:**
- `cmd/internal/groupsync/` - Consumer directly calls service methods
- `cmd/api/context.go:189-200` - groupCreated service wraps create with sender

**Risk:** Tight coupling between messaging and domain logic. Changes to message format require changes to multiple files.

---

## 11. Architectural Debt Summary

| Concern | Severity | Effort to Fix | Blocks Rewrite? |
|---------|----------|---------------|-----------------|
| Generic tree → SQL incompatibility | **Critical** | High | **Yes** - requires new model |
| Efficiency (Raw, Hash, traversal) | High | Medium | Partial |
| No portability accommodation | Medium | Low | No |
| Projection bug | Medium | Low | No |
| Schema registration rigidity | Medium | Medium | No |
| No bulk operations | Low | High | No |
| Visitor traversal overhead | Low | Medium | No |

---

*Concerns audit: 2026-05-07*
