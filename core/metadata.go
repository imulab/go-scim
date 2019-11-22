package core

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

const (
	DefaultMetadataId = "urn:imulab:metadata:scim:default"
)

var (
	// Entry point to access metadata.
	Meta = &metadataRepository{
		data: make(map[string]map[string]Metadata, 0),
	}
)

type (
	// A unified but flexible API to allow extension of metadata from different user packages.
	Metadata interface {
		// Returns the attribute id that can be used to lookup SCIM attribute. The attribute id is
		// treated case insensitive.
		AttributeId() string
	}

	// A provider that manages a number of metadata
	MetadataProvider interface {
		// Returns a globally id to identify the metadata's provider. The provider id is treated
		// case insensitive.
		ProviderId() string

		// Returns all metadata that this provider manages.
		Metadata() []Metadata
	}

	// Metadata that is used by the core package
	DefaultMetadata struct {
		// Attribute id that links this metadata to the attribute that it describes.
		Id string `json:"id"`

		// path is the name that can be used to address
		// the property from the root of the resource. (i.e. userName, emails.value)
		Path string `json:"path"`

		// True if the property shall participate in identity comparison
		// Two complex properties shall be regarded as equal when all their
		// respective identity sub properties equal. For instance, when
		// comparing emails, two emails are equals when their value and type
		// equal.
		IsIdentity bool `json:"is_identity"`

		// True if the property is an exclusive property. Exclusive properties
		// are boolean properties within a multi-valued complex property that
		// has true value. Among all the values in this multi-valued complex
		// property, at most one true value can exist at the same time. This
		// field controls this invariant, as to ensure when new value with a
		// true exclusive property is added, the existing true exclusive property
		// is turned off (set to false).
		IsExclusive bool `json:"is_exclusive"`

		// An integer indicating the order of the traversal when the property having
		// this attribute is visited. The order is only effective when traversing sub
		// properties in the context of a complex property. The order is in ascending
		// order, smaller order will be visited first.
		VisitOrder int `json:"visit_order"`

		// List of strings that mark the attribute for special processing. These
		// annotations are meaningless in the context of this package, but may be
		// picked up by other packages.
		Annotations []string `json:"annotations"`
	}

	// Default metadata provider
	DefaultMetadataProvider struct {
		Id           string             `json:"id"`
		MetadataList []*DefaultMetadata `json:"metadata"`
	}

	// Central repository to store metadata for attribute's from different providers. The repository is not designed
	// to withstand concurrent writes because the usage scenario is expected to have minimum contention and no writes
	// after the initial setup period.
	metadataRepository struct {
		// data indexed by attribute id (to lower case) and by provider id (to lower case)
		data map[string]map[string]Metadata
	}
)

// Parse default metadata from file content. Panics if there's an error.
func MustLoadDefaultMetadataProvider(filePath string) MetadataProvider {
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	mdp := new(DefaultMetadataProvider)
	err = json.Unmarshal(raw, mdp)
	if err != nil {
		panic(err)
	}

	return mdp
}

func (m *DefaultMetadata) AttributeId() string {
	return m.Id
}

func (p *DefaultMetadataProvider) ProviderId() string {
	return p.Id
}

func (p *DefaultMetadataProvider) Metadata() []Metadata {
	cast := make([]Metadata, 0)
	for _, m := range p.MetadataList {
		cast = append(cast, m)
	}
	return cast
}

// Save all metadata from the given provider into the repository.
// The provider must have non-empty provider id and metadata must have non-empty attribute id.
func (r *metadataRepository) Add(provider MetadataProvider) {
	keyProvider := strings.ToLower(provider.ProviderId())
	if len(keyProvider) == 0 {
		panic("zero length provider id in metadata")
	}

	for _, metadata := range provider.Metadata() {
		keyAttr := strings.ToLower(metadata.AttributeId())
		if len(keyAttr) == 0 {
			panic("zero length attribute id in metadata")
		}

		if _, ok := r.data[keyAttr]; !ok {
			r.data[keyAttr] = make(map[string]Metadata)
		}

		r.data[keyAttr][keyProvider] = metadata
	}
}

// Save a single piece of metadata to the repository. This method is intended to provide convenience to tests so
// it does not have to reinvent the whole metadata provider.
func (r *metadataRepository) AddSingle(providerId string, metadata Metadata) {
	keyProvider := strings.ToLower(providerId)
	if len(keyProvider) == 0 {
		panic("zero length provider id in metadata")
	}

	keyAttr := strings.ToLower(metadata.AttributeId())
	if len(keyAttr) == 0 {
		panic("zero length attribute id in metadata")
	}

	if _, ok := r.data[keyAttr]; !ok {
		r.data[keyAttr] = make(map[string]Metadata)
	}

	r.data[keyAttr][keyProvider] = metadata
}

// Get the metadata corresponding to the attribute and provider. If no metadata, returns nil. The method does not return
// ok or error information in order to maintain a single result API to improve code readability. It is expected in most
// cases that the caller knows what it is doing and the metadata actually exists.
func (r *metadataRepository) Get(attributeId, providerId string) Metadata {
	v0, ok := r.data[strings.ToLower(attributeId)]
	if !ok {
		return nil
	}

	v1, ok := v0[strings.ToLower(providerId)]
	if !ok {
		return nil
	}

	return v1
}

// Load the metadata provider, save it into the repository, or panic.
func (r *metadataRepository) MustLoad(filePath string, metadataProvider interface{}) MetadataProvider {
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

	r.Add(metadataProvider.(MetadataProvider))
	return metadataProvider.(MetadataProvider)
}
