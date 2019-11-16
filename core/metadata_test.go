package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMustLoadDefaultMetadataProvider(t *testing.T) {
	mdp := MustLoadDefaultMetadataProvider("../resource/metadata/default_metadata.json")
	assert.NotNil(t, mdp)

	Meta.Add(mdp)
	md := Meta.Get("urn:ietf:params:scim:schemas:core:2.0:User:userName", DefaultMetadataId)
	assert.NotNil(t, md)
}
