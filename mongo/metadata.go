package mongo

import "encoding/json"

// a local cache of all metadata
var metadataHub map[string]*Metadata

func init() {
	metadataHub = make(map[string]*Metadata)
}

// Read metadata and add all metadata to the hub
func ReadMetadata(raw []byte) error {
	p := new(struct {
		Metadata []*Metadata `json:"metadata"`
	})
	if err := json.Unmarshal(raw, p); err != nil {
		return err
	}
	for _, md := range p.Metadata {
		metadataHub[md.Id] = md
	}
	return nil
}

// Mongo package extension to spec.Attribute. Here we define a MongoDB property alias
// to override the attribute name when saving to or reading from MongoDB.
type Metadata struct {
	Id        string `json:"id"`
	MongoName string `json:"mongoName"`
	MongoPath string `json:"mongoPath"`
}
