package core

import (
	"encoding/json"
	"io/ioutil"
)

// This file contains the loader mechanism to
// load necessary resources like schema, resourceType, metadata...

type (
	// Loader for schema and resourceType
	Loader struct {}
)

func (l Loader) MustSchema(filePath string) *Schema {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	schema, err := ParseSchema(raw)
	if err != nil {
		panic(err)
	}

	Schemas.Add(schema)
	return schema
}

func (l Loader) MustResourceType(filePath string) *ResourceType {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	resourceType, err := ParseResourceType(raw)
	if err != nil {
		panic(err)
	}

	return resourceType
}

func (l Loader) MustMetadataProvider(filePath string, metadataProvider interface{}) MetadataProvider {
	if _, ok := metadataProvider.(MetadataProvider); !ok {
		panic("argument metadataProvider must be of type MetadataProvider")
	}

	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(raw, metadataProvider)
	if err != nil {
		panic(err)
	}

	return metadataProvider.(MetadataProvider)
}