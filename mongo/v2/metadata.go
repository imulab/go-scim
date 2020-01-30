package v2

import (
	"bytes"
	"encoding/json"
	"io"
)

// a local cache of all metadata
var metadataHub map[string]*Metadata

func init() {
	metadataHub = make(map[string]*Metadata)
}

// Read metadata and add all metadata to the hub
func ReadMetadata(raw []byte) error {
	return ReadMetadataFromReader(bytes.NewReader(raw))
}

// ReadMetadataFromReader reads and registered the JSON encoded metadata from reader.
func ReadMetadataFromReader(reader io.Reader) error {
	p := new(struct {
		Metadata []*Metadata `json:"metadata"`
	})
	if err := json.NewDecoder(reader).Decode(p); err != nil {
		return err
	}
	for _, md := range p.Metadata {
		metadataHub[md.Id] = md
	}
	return nil
}

// Mongo package extension to spec.Attribute. Here we define a MongoDB property alias
// to override the attribute name when saving to or reading from MongoDB. This is necessary because
// some valid SCIM field names are not valid in MongoDB.
//
// To define metadata to be supplied to ReadMetadataFromReader, compose something similar to:
//	{
//		"metadata": [
//			{
//				"id": "urn:ietf:params:scim:schemas:core:2.0:User:groups.$ref",
//				"mongoName": "ref",
//				"mongoPath": "groups.ref"
//			}
//		]
//	}
type Metadata struct {
	Id        string `json:"id"`
	MongoName string `json:"mongoName"`
	MongoPath string `json:"mongoPath"`
}
