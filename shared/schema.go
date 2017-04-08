package shared

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

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

type AttributeSource interface {
	GetAttribute(p Path, recursive bool) *Attribute
}

type Schema struct {
	Schemas     []string     `json:"schemas"`
	Id          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Attributes  []*Attribute `json:"attributes"`
}

func (s *Schema) ToAttribute() *Attribute {
	return &Attribute{
		Type:          TypeComplex,
		MultiValued:   false,
		SubAttributes: s.Attributes,
		Assist:        &Assist{JSONName: "", Path: "", FullPath: ""},
	}
}

func (s *Schema) GetAttribute(p Path, recursive bool) *Attribute {
	for _, attr := range s.Attributes {
		if strings.ToLower(attr.Name) == strings.ToLower(p.Base()) {
			if recursive {
				return attr.GetAttribute(p.Next(), recursive)
			} else {
				return attr
			}
		} else {
			switch attr.Name {
			case "schemas", "id", "externalId", "meta":
			default:
				if strings.ToLower(fmt.Sprintf("%s:%s", s.Id, attr.Name)) == strings.ToLower(p.Base()) {
					if recursive {
						return attr.GetAttribute(p.Next(), recursive)
					} else {
						return attr
					}
				}
			}
		}
	}
	return nil
}

type Attribute struct {
	Name            string       `json:"name"`
	Type            string       `json:"type"`
	SubAttributes   []*Attribute `json:"subAttributes"`
	MultiValued     bool         `json:"multiValued"`
	Description     string       `json:"description"`
	Required        bool         `json:"required"`
	CanonicalValues []string     `json:"canonicalValues"`
	CaseExact       bool         `json:"caseExact"`
	Mutability      string       `json:"mutability"`
	Returned        string       `json:"returned"`
	Uniqueness      string       `json:"uniqueness"`
	ReferenceTypes  []string     `json:"referenceTypes"`
	Assist          *Assist      `json:"_assist"`
}

func (a *Attribute) GetAttribute(p Path, recursive bool) *Attribute {
	if p == nil {
		return a
	}

	for _, subAttr := range a.SubAttributes {
		if strings.ToLower(subAttr.Name) == strings.ToLower(p.Base()) {
			if recursive {
				return subAttr.GetAttribute(p.Next(), recursive)
			} else {
				return subAttr
			}

		}
	}

	return nil
}

type Assist struct {
	JSONName      string   `json:"_jsonName"`      // JSON field name used to render this field
	Path          string   `json:"_path"`          // period delimited field names, useful to retrieve nested fields
	FullPath      string   `json:"_full_path"`     // Path prefixed with the URN of this resource
	ArrayIndexKey []string `json:"_arrayIndexKey"` // the field names of the multiValued complex fields that can be used as a search index
}

const (
	UserUrn         = "urn:ietf:params:scim:schemas:core:2.0:User"
	GroupUrn        = "urn:ietf:params:scim:schemas:core:2.0:Group"
	ResourceTypeUrn = "urn:ietf:params:scim:schemas:core:2.0:ResourceType"
	SPConfigUrn     = "urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"
	SchemaUrn       = "urn:ietf:params:scim:schemas:core:2.0:Schema"
	ErrorUrn        = "urn:ietf:params:scim:api:messages:2.0:Error"
	ListResponseUrn = "urn:ietf:params:scim:api:messages:2.0:ListResponse"
	PathOpUrn       = "urn:ietf:params:scim:api:messages:2.0:PatchOp"
	SearchUrn       = "urn:ietf:params:scim:api:messages:2.0:SearchRequest"
	BulkRequestUrn  = "urn:ietf:params:scim:api:messages:2.0:BulkRequest"
	BulkResponseUrn = "urn:ietf:params:scim:api:messages:2.0:BulkResponse"

	TypeString    = "string"
	TypeBoolean   = "boolean"
	TypeBinary    = "binary"
	TypeDecimal   = "decimal"
	TypeInteger   = "integer"
	TypeDateTime  = "datetime"
	TypeReference = "reference"
	TypeComplex   = "complex"

	ReadOnly  = "readOnly"
	ReadWrite = "readWrite"
	Immutable = "immutable"
	WriteOnly = "writeOnly"

	Always  = "always"
	Never   = "never"
	Default = "default"
	Request = "request"

	None   = "none"
	Server = "server"
	Global = "global"

	External = "external"
	Uri      = "uri"
)
