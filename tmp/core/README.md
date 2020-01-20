# Core Module

The core module provides features described in [RFC 7643 - SCIM: Core Schema](https://tools.ietf.org/html/rfc7643), as well as the foundation for features described in [RFC 7644 - SCIM: Protocol](https://tools.ietf.org/html/rfc7644).

## Attributes

At the core of the SCIM specification is the attribute, which describes data type and constraints. The core module uses an extended version of the attributes from the specification. In addition to defined properties such as `type`, `required`, `multiValued`, etc, we added four more internal properties:

- `id`: The unique id of an attribute. This id can be used to identity the attribute and correlate other metadata to the attribute, thus serves as the major extension point.
- `_path`: The complete path from the resources's root. The path saves processing complexity so that the traversal mechanism do not have to remember its trail in order to point out the full path of an attribute.
- `_index`: The index specifies a relative ascending order for an attribute among its fellow attributes. Although SCIM does not specify attribute orders and it is okay to scramble them, it is just so much nicer to be able to return attributes in a determined order.
- `_annotations`: The string based annotations applied to the attribute. Other functions may detect annotations on an attribute and carry out extra logic. This is another major extension point. The module comes with several baked in annotations in `annotations/annotations.go`.

The added internal properties are never leaked to the API. This is ensured by the `json.Marshaler` implementation.

## Schema and ResourceType

The core module lets user define schemas and then assemble them into resource types. The schema definition completely follows the specification, with the addition of the internal properties described above in its attributes. It is important to cache them into the `SchemaHub` after parsing them.

```go
schema := new(spec.Schema)
if err := json.Unmarshal(rawJsonBytes, schema); err != nil {
	return err
}
spec.SchemaHub.Put(schema)
```

After all schemas are parsed, we can then parse the resource types which reference their id in schemas and schema extensions.

```go
resourceType := new(spec.ResourceType)
if err := json.Unmarshal(rawJsonBytes, resourceType); err != nil {
	return err
}
```

Check out the `internal` folder for schema and resource type definitions in test cases.

## Property and Container

`Property` references an attribute that describes its data requirement and holds a piece of resource data that conforms to that attribute. `Property` is the reason why the project can traverse dynamic data without resorting to reflection.

All SCIM attribute types have their corresponding property:

- `stringProperty`: holds and represents data of Go's `string` type, or `nil` if unassigned.
- `integerProperty`: holds and represents data of Go's `int64` type, or `nil` if unassigned.
- `decimalProperty`: holds and represents data of Go's `float64` type, or `nil` if unassigned.
- `booleanProperty`: holds and represents data of Go's `bool` type, or `nil` if unassigned.
- `referenceProperty`: holds and represents data of Go's `string` type, or `nil` if unassigned.
- `dateTimeProperty`: holds data of Go's `time.Time` type, represents data of Go's `string` type in ISO8601 format, or `nil` if unassigned.
- `binaryProperty`: holds data of Go's `[]byte` type, represents data of Go's `string` type in base64 encoded format, or `nil` if unassigned.

As with `Property` holds data, `Container` holds `Property`. Two containers are present:

- `complexProperty`: holds a list of sub properties, corresponds to SCIM's `complex` type
- `multiValuedProperty`: holds a list of member properties, corresponds to SCIM attribute where `multiValued=true`.

The `multiValuedProperty` is special because this type was not explicitly modelled in the SCIM specification. The modelling of this type of property makes traversing resource structure much easier, although it adds one complexity: the attribute of its member properties will be derived. They are derived by setting `multiValued=false` and appending `$elem` to the container attribute. For instance:

```
# emails attribute (incomplete for brevity)
{
	"id": "urn:ietf:params:scim:schemas:core:2.0:User",
	"name": "emails",
	"type": "complex",
	"multiValued": true
}

# derived emails member attribute (incomplete for brevity)
{
	"id": "urn:ietf:params:scim:schemas:core:2.0:User$elem",
	"name": "emails",
	"type": "complex",
	"multiValued": false
}
```

### Traversal

The core module introduces two ways to traverse the resource data structure: `Visitor` and `Navigator`, along with its fluent API derivation `FluentNavigator`. These two mechanism supports different use cases.

`Visitor` is an interface to be implemented by the intended party to visit the resource in depth-first-search order. The visiting party may chose to skip or visit a certain property through callback defined in the interface. However, the main control lies with the DFS traversal on the resource side. This is a passive traversal mechanism. It is usually used in serialization scenario.

`Navigator` is a structure that exposes methods to focus on sub properties addressable by different types of handle (i.e. by name, by index, by criteria). When the caller is done with the property, it can go back to the last state by calling `Retract`. This is an active traversal mechanism. It is usually used in deserialization scenario.

### Event System

The property design also features an event system, which allows different parts of the resource to react to changes in other non-local properties.

When the property value is modified, an event will be generated describing the modification. This event will be propagated up the resource tree structure and eventually arrive at the root of the resource. Along the way during the propagation, `Subscriber`s can subscribe to a property for such modification events and react to it.

Subscribers are added to the property using the attribute annotation. Some subscribers are already baked in. For instance, `prop.NewExclusivePrimarySubscriber` produces a subscriber loads itself onto properties annotated with `@ExclusivePrimary` (usually on multiValued complex property with a primary boolean sub property). It reacts to boolean property changes from its sub properties and maintain at most one primary property that has `true` value. This implements the SCIM requirement:

> The primary attribute value "true" MUST appear no more than once.

To add custom subscribers, use `prop.AddEventFactory` method.

## Expression

The `expr` package maintains the `CompilePath` and `CompileFilter` method to compile SCIM path and SCIM filters into a hybrid abstract syntax tree-list. The resulting data structure can then be used to guide the traversal of the resource.

`expr.Expression` is the main data structure, it represents a meaningful token in the path or filter expression. For instance, for path `emails[value eq "foo@bar.com"].primary`, compilation turns it into a structure where each component is a separate `*expr.Expression`.

```
emails -> eq -> primary
        /   \
     value  "foo@bar.com"
```

One pecularity to watch out for is that SCIM paths can be prepended by namespaces. For instance, `emails` are equivalent to `urn:ietf:params:scim:schemas:core:2.0:emails`. Notice the `2.0` part in the namespace violates the SCIM attribute name syntax by introducing a dot that does not indicate path separation. To properly compile these paths, the compiler needs to learn all the expected namespaces of likes so it can distinguish between path separation and namespace dot when seeing a dot. To do so:

```go
expr.Register(resourceType)
```

This will register the ids of all schema and schema extensions of the resource type with the expression compiler.