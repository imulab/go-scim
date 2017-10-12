package resource

import (
	"sync"
	"github.com/pkg/errors"
	"strings"
	"path/filepath"
	"os"
	"io/ioutil"
	"encoding/json"
)

var (
	ErrorAttributeNotFound = errors.New("attribute is not found")
	vault 		*mapAttributeVault
	oneVault	sync.Once
)

// Create an attribute vault from a list of schemas. Usually called on
// application initialization. Can only be called once, subsequent calls
// will just return the result of the first call. So this method can be
// used as both a constructor and an accessor.
func GetAttributeVault(schemas ...*Schema) AttributeVault {
	oneVault.Do(func() {
		vault = &mapAttributeVault{data:make(map[string]*Attribute)}
		for _, sch := range schemas {
			indexAttributes(sch.Attributes...)
		}
	})
	return vault
}

func indexAttributes(attributes ...*Attribute) {
	for _, attr := range attributes {
		vault.data[strings.ToLower(attr.Guide.Tag)] = attr
		if attr.SubAttributes != nil {
			indexAttributes(attr.SubAttributes...)
		}
	}
}

// Main entry point of getting an attribute by a corresponding SCIM tag value.
// The lookup process is case insensitive. In case an attribute is not found,
// ErrorAttributeNotFound will be returned.
type AttributeVault interface {
	Get(tag string) (*Attribute, error)
}

type mapAttributeVault struct {
	data 	map[string]*Attribute
}

func (v *mapAttributeVault) Get(tag string) (*Attribute, error) {
	if attr, ok := v.data[strings.ToLower(tag)]; !ok {
		return nil, ErrorAttributeNotFound
	} else {
		return attr, nil
	}
}

type Schema struct {
	Schemas     []string     `json:"schemas,omitempty"`
	Id          string       `json:"id,omitempty"`
	Name        string       `json:"name,omitempty"`
	Description string       `json:"description,omitempty"`
	Attributes  []*Attribute `json:"attributes,omitempty"`
}

type Attribute struct {
	Name            string       `json:"name,omitempty"`
	Type            string       `json:"type,omitempty"`
	SubAttributes   []*Attribute `json:"subAttributes,omitempty"`
	MultiValued     bool         `json:"multiValued"`
	Description     string       `json:"description,omitempty"`
	Required        bool         `json:"required"`
	CanonicalValues []string     `json:"canonicalValues,omitempty"`
	CaseExact       bool         `json:"caseExact,omitempty"`
	Mutability      string       `json:"mutability,omitempty"`
	Returned        string       `json:"returned,omitempty"`
	Uniqueness      string       `json:"uniqueness,omitempty"`
	ReferenceTypes  []string     `json:"referenceTypes,omitempty"`
	Guide			Guide 		 `json:"_guide"`
}

func (attr *Attribute) IsObject() bool {
	return !attr.MultiValued && attr.Type == TypeComplex
}

func (attr *Attribute) IsObjectArray() bool {
	return attr.MultiValued && attr.Type == TypeComplex
}

func (attr *Attribute) IsSimpleField() bool {
	return !attr.MultiValued && attr.Type != TypeComplex
}

func (attr *Attribute) IsSimpleArray() bool {
	return attr.MultiValued && attr.Type != TypeComplex
}

type Guide struct {
	// SCIM tag of the attribute, matching the Go tag from the domain model. Unique
	Tag 	string		`json:"_tag"`

	// Aliases for the current attribute. It will be used to resolve values from incoming json.
	// For instance:
	// - 'username' has the alias of 'username' and 'urn:ietf:params:scim:schemas:core:2.0:User:username'
	// - 'name.familyName' has the alias of 'familyName', 'name.familyName' and 'urn:ietf:params:scim:schemas:core:2.0:User:name.familyName'
	Aliases []string	`json:"_aliases"`
}

// Utility to parse a schema from file
func ParseSchema(filePath string) (*Schema, string, error) {
	path, err := filepath.Abs(filePath)
	if err != nil {
		return nil, "", err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, "", err
	}

	schema := &Schema{}
	err = json.Unmarshal(fileBytes, &schema)
	if err != nil {
		return nil, "", err
	}

	return schema, string(fileBytes), nil
}