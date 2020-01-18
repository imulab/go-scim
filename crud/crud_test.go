package crud

import "testing"

func TestAdd(t *testing.T) {

}

func TestReplace(t *testing.T) {

}

func TestDelete(t *testing.T) {

}

const (
	testCoreSchema = `
{
  "id": "core",
  "name": "core",
  "attributes": [
    {
      "id": "schemas",
      "name": "schemas",
      "type": "string",
      "multiValued": true,
      "_index": 0,
      "_path": "schemas",
      "_annotations": {
        "@AutoCompact": {}
      }
    },
    {
      "id": "id",
      "name": "id",
      "type": "string",
      "_index": 1,
      "_path": "id"
    },
    {
      "id": "meta",
      "name": "meta",
      "type": "complex",
      "_index": 2,
      "_path": "meta",
      "_annotations": {
        "@StateSummary": {}
      },
      "subAttributes": [
        {
          "id": "meta.version",
          "name": "version",
          "type": "string",
          "_index": 0,
          "_path": "meta.version"
        },
        {
          "id": "meta.location",
          "name": "location",
          "type": "reference",
          "_index": 1,
          "_path": "meta.location"
        }
      ]
    }
  ]
}
`
	testMainSchema = `
{
  "id": "main",
  "name": "main",
  "attributes": [
    {
      "id": "emails",
      "name": "emails",
      "type": "complex",
      "multiValued": true,
      "_index": 100,
      "_path": "emails",
      "_annotations": {
        "@ExclusivePrimary": {},
        "@AutoCompact": {},
        "@ElementAnnotations": {
          "@StateSummary": {}
        }
      },
      "subAttributes": [
        {
          "id": "emails.value",
          "name": "value",
          "type": "string",
          "_index": 0,
          "_path": "emails.value",
          "_annotations": {
            "@Identity": {}
          }
        },
        {
          "id": "emails.primary",
          "name": "primary",
          "type": "boolean",
          "_index": 1,
          "_path": "emails.primary",
          "_annotations": {
            "@Primary": {}
          }
        }
      ]
    }
  ]
}
`
	testResourceType = `
{
  "id": "Test",
  "name": "Test",
  "schema": "main"
}
`
)
