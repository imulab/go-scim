package mongo

import (
	"github.com/imulab/go-scim/pkg/core"
)

const (
	MetadataId = "urn:imulab:metadata:scim:mongo"
)

type (
	// Metadata for attributes with respect to MongoDB.
	Metadata struct {
		Id      string `json:"id"`
		DbAlias string `json:"db_alias"`
	}

	// Provider of MongoDB related metadata
	MetadataProvider struct {
		Id           string      `json:"id"`
		MetadataList []*Metadata `json:"metadata"`
	}
)

func (m *Metadata) AttributeId() string {
	return m.Id
}

func (p *MetadataProvider) ProviderId() string {
	return p.Id
}

func (p *MetadataProvider) Metadata() []core.Metadata {
	cast := make([]core.Metadata, 0)
	for _, m := range p.MetadataList {
		cast = append(cast, m)
	}
	return cast
}
