# MongoDB Module

[![GoDoc](https://godoc.org/github.com/imulab/go-scim/mongo/v2?status.svg)](https://godoc.org/github.com/imulab/go-scim/mongo/v2)

This module provides the capability to persist SCIM resources in MongoDB.

## :bulb: Usage

To get this package:

```bash
# make sure Go 1.13
go get github.com/imulab/go-scim/mongo/v2
```

## :floppy_disk: Persistence

This basic `db.DB` implementation in this module assumes one-to-one mapping between a SCIM resource type and a MongoDB 
collection.

### Index

MongoDB indexes are automatically created for attributes whose `uniqueness=server` or `uniqueness=global`, and for
attributes who were annotated with `@MongoIndex`. When `uniqueness` is not `none`, a unique index is created; otherwise,
just the single field index is created. This module does not support the creation of composite index, or geo-spatial
indexes. In addition, index creation failures are logged as warning to the logger, instead of being returned as an
error. A particular failure situation to watch out for is that, after MongoDB `4.2`, creating an already existing index
will actually return an error, in contrast to just returning an implicit success in versions before. Such failure can still
be considered an implicit success in our case as the indexes will be there.

### Metadata

The domain space of SCIM path characters and MongoDB path characters do not completely overlap. Some characters legal
in the SCIM world may not be legal in the MongoDB world. One example is the `$` character, typically seen in the `$ref`
attribute. This module provides a way to correlate MongoDB specific metadata to SCIM attributes so the attribute name and
attribute path can be override.

The module provides stock metadata in the `public` folder which covers `User`, `Group` and Enterprise User extension. A
typical metadata looks like:

```json
{
  "metadata": [
    {
      "id": "urn:ietf:params:scim:schemas:core:2.0:User:groups.$ref",
      "mongoName": "ref",
      "mongoPath": "groups.ref"
    }
  ]
}
```

The `id` field is the unique attribute id. `mongoName` specifies the field name within its containing object. `mongoPath`
specifies the path name from the document root. Use `ReadMetadata` to register bytes of the metadata file.

### Atomicity

The module utilizes the atomicity of MongoDB and safely performs modification operations without explicitly locking the
resource. `Replace` and `Delete` operations would only perform data modification if the `id` and `meta.version` fields
matches the record in MongoDB. If no match was found, a `conflict` error is returned to indicate some current process
must have modified the resource in between.

### Projection

The projection feature of the `Query` method is not completely fool-proof. It does not check for the `returned` property
of the target attribute pointed to by the `attributes` or `excludedAttributes` specified in `crud.Projection` parameter.
This is done intentionally, as `db.DB` is not a client facing component with respect to SCIM API. Any projection
parameters supplied should have been pre-sanitized so that it does not contradict with the `returned` property. 

There are also cases where callers may wish to carry out operations after the query on fields not requested by the client.
In this case, callers can use `Options.IgnoreProjection()` to disable projection altogether so the database always 
return the full version of the resource.

## :black_nib: Serialization

This module provides direct serialization and de-deserialization between SCIM resource and MongoDB BSON format, without
the transition of an intermediary data format.

The serialization of the less significant components like filter, sort, pagination and projection are still being
converted to `bson.D`, before being serialized to BSON and sent to MongoDB.

## :construction: Testing

This module uses [org/dockertest](https://github.com/ory/dockertest) to setup testing docker containers at test run time.
The environment variables to customize the local docker connection are:

```bash
# shows the environment variable and their default value
TEST_DOCKER_ENDPOINT=""
TEST_DOCKER_MONGO_IMAGE="bitnami/mongodb"
TEST_DOCKER_MONGO_TAG=latest
TEST_DOCKER_MONGO_USERNAME=testUser
TEST_DOCKER_MONGO_SECRET=s3cret
TEST_DOCKER_MONGO_DB_NAME=mongo_database_test_suite
```