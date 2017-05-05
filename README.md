# What is GoSCIM

GoSCIM is the LEGO blocks for building a custom [SCIM v2](http://www.simplecloud.info/) implementation. It already implemented most of, if not all, the functionalities specified in the IETF document.

## What is included?

Functionalities like validation, read-only attributes assignment, query resolution, persistence and JSON serialization are included. In addition, a [sample server](https://github.com/davidiamyou/go-scim/tree/development/example) is included with the project just to demonstrate how easy it is to put together these components into a functioning SCIM API Server.

## How to Run?

GoSCIM does not compile to a runnable artifact. Instead, it enables building one. If you want to quickly have a preview of an example server that has been already built:

```
cd $GOPATH/src/github.com/davidiamyou/go-scim/example
go run server.go
```

## Key Know-Hows

This section explains some of the design decisions. Knowing these may save you some time in figuring out about your own implementations.

### Schema vs. Internal Schema

The SCIM protocol defines resources based on schemas. Naturally GoSCIM uses these schema attribute definitions to perform validation, serialize JSON and so on. However, GoSCIM extended the SCIM defined schema attribute to include some more useful information called `Assists`. Some key assists include:
- JSON name of the field (i.e. `displayName`)
- Path of the field (i.e. `emails.value`, `active`, `name.familyName`)
- Full URN name of the field (i.e. `id`, `urn:ietf:params:scim:schemas:core:2.0:User:name.familyName`)
- Array key field (i.e. the field that could be used to uniquely identify a complex array entry, for instance, `value`)

GoSCIM uses the extended schema internally and renders SCIM defined schema. It is important to know, although one extends another, they are separate entities inside GoSCIM.

### Types

The following table relates SCIM type to Go type:

SCIM Type | Go Type
--- | ---
string | `string`
reference | `string`
binary (base64) | `string`
date time | `string`
boolean | `bool`
integer | `int64`
decimal | `float64`
complex | `map[string]interface{}`
array | `[]interface{}`

### JSON Serialization

The entry point of serializing an object in GoSCIM is `MarshalJSON` in `shared/json.go`. Underneath, it tries to utilize
go's native JSON capabilities whenever possible. However, when serializing resources, it does not rely on tags, rather it
seeks advice from SCIM schema.

### Query Resolution

GoSCIM tries to parse the query text into an abstract syntax tree first. The tree then can be flattened and transformed to whichever query language the database understands.

GoSCIM supports MongoDB. The `mongo` directory contains an example of how the AST can be flattened to MongoDB query. It should work similarly at least with other document based databases.

### Persistence

GoSCIM supports MongoDB, but it does not restrict adopters to it. It provides a `Repository` interface in `shared/persistence.go` which other database choices can implement. The MongoDB implementation is contained in the `mongo` folder.

### Other Interfaces

- `WebRequest`: an abstraction of HTTP request. Useful when delegating mock requests, for instance, during bulk operation.
- `WebResponse`: similar to `WebRequest` in intentions.
- `PropertySource`: abstraction of a property provider. The example server uses a map to implement this. Actual implementations can be projects like `viper`
- `Logger`: abstraction of a logger. The example server implementations just prints to console. Actual logger can be used in real implementations.
- `ReadOnlyAssignment`: logic to assign value to read only fields. GoSCIM already provides `id`, `meta` and `group` assignment, plus copying any read only value from existing resource reference during update. User needs to implement this interface per custom readonly field. 
