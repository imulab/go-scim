package shared

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
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
	Schemas     []string     `json:"schemas,omitempty"`
	Id          string       `json:"id,omitempty"`
	Name        string       `json:"name,omitempty"`
	Description string       `json:"description,omitempty"`
	Attributes  []*Attribute `json:"attributes,omitempty"`
}

func (s *Schema) ToAttribute() *Attribute {
	return &Attribute{
		Type:          TypeComplex,
		MultiValued:   false,
		Mutability:    ReadWrite,
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
	Assist          *Assist      `json:"_assist"`
}

func (a *Attribute) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
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
	}{
		a.Name,
		a.Type,
		a.SubAttributes,
		a.MultiValued,
		a.Description,
		a.Required,
		a.CanonicalValues,
		a.CaseExact,
		a.Mutability,
		a.Returned,
		a.Uniqueness,
		a.ReferenceTypes,
	})
}

func (a *Attribute) EqualsToPath(p Path) bool {
	switch strings.ToLower(p.CollectValue()) {
	case strings.ToLower(a.Assist.FullPath):
		return true
	case strings.ToLower(a.Assist.Path):
		return true
	default:
		return false
	}
}

func (a *Attribute) Assigned(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}

	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String, reflect.Map, reflect.Array, reflect.Slice:
		return v.Len() > 0
	default:
		return true
	}
}

func (a *Attribute) Clone() *Attribute {
	return &Attribute{
		Name:            a.Name,
		Type:            a.Type,
		SubAttributes:   a.SubAttributes,
		MultiValued:     a.MultiValued,
		Description:     a.Description,
		Required:        a.Required,
		CanonicalValues: a.CanonicalValues,
		CaseExact:       a.CaseExact,
		Mutability:      a.Mutability,
		Returned:        a.Returned,
		Uniqueness:      a.Uniqueness,
		ReferenceTypes:  a.ReferenceTypes,
		Assist:          a.Assist,
	}
}

func (a *Attribute) TypeExpectation() string {
	var expects string = ""
	switch a.Type {
	case TypeString, TypeBinary, TypeReference, TypeDateTime:
		expects = "string"
	case TypeInteger:
		expects = "integer"
	case TypeDecimal:
		expects = "decimal"
	case TypeBoolean:
		expects = "boolean"
	case TypeComplex:
		expects = "complex"
	}
	if a.MultiValued {
		expects += " array"
	}
	return expects
}

func (a *Attribute) ExpectsString() bool {
	switch a.Type {
	case TypeString, TypeDateTime, TypeReference, TypeBinary:
		return !a.MultiValued
	default:
		return false
	}
}

func (a *Attribute) ExpectsStringArray() bool {
	switch a.Type {
	case TypeString, TypeDateTime, TypeReference, TypeBinary:
		return a.MultiValued
	default:
		return false
	}
}

func (a *Attribute) ExpectsInteger() bool {
	return !a.MultiValued && a.Type == TypeInteger
}

func (a *Attribute) ExpectsFloat() bool {
	return !a.MultiValued && a.Type == TypeDecimal
}

func (a *Attribute) ExpectsBool() bool {
	return !a.MultiValued && a.Type == TypeBoolean
}

func (a *Attribute) ExpectsBinary() bool {
	return !a.MultiValued && a.Type == TypeBinary
}

func (a *Attribute) ExpectsComplex() bool {
	return !a.MultiValued && a.Type == TypeComplex
}

func (a *Attribute) ExpectsComplexArray() bool {
	return a.MultiValued && a.Type == TypeComplex
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
	ResourceTypeUrn = "urn:ietf:params:scim:schemas:core:2.0:resourceType"
	SPConfigUrn     = "urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"
	SchemaUrn       = "urn:ietf:params:scim:schemas:core:2.0:Schema"
	ErrorUrn        = "urn:ietf:params:scim:api:messages:2.0:Error"
	ListResponseUrn = "urn:ietf:params:scim:api:messages:2.0:ListResponse"
	PatchOpUrn      = "urn:ietf:params:scim:api:messages:2.0:PatchOp"
	SearchUrn       = "urn:ietf:params:scim:api:messages:2.0:SearchRequest"
	BulkRequestUrn  = "urn:ietf:params:scim:api:messages:2.0:BulkRequest"
	BulkResponseUrn = "urn:ietf:params:scim:api:messages:2.0:BulkResponse"

	UserEnterpriseUrn = "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"

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

	UserResourceType                  = "User"
	GroupResourceType                 = "Group"
	SchemaResourceType                = "Schema"
	ResourceTypeResourceType          = "ResourceType"
	ServiceProviderConfigResourceType = "ServiceProviderConfig"
)
