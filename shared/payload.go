package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// ----------------------------------
// Patch
// ----------------------------------
type Modification struct {
	Schemas []string `json:"schemas"`
	Ops     []Patch  `json:"Operations"`
}

const (
	Add     = "add"
	Remove  = "remove"
	Replace = "replace"
)

type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func (m Modification) Validate() error {
	if len(m.Schemas) != 1 && m.Schemas[0] != PatchOpUrn {
		return Error.InvalidParam("schemas", PatchOpUrn, fmt.Sprintf("%+v", m.Schemas))
	}

	if len(m.Ops) == 0 {
		return Error.InvalidParam("Operations", "at least one patch operation", "none")
	}

	for _, patch := range m.Ops {
		switch patch.Op {
		case Add:
			if patch.Value == nil {
				return Error.InvalidParam("value of add op", "to be present", "nil")
			} else if len(patch.Path) == 0 {
				if _, ok := patch.Value.(map[string]interface{}); !ok {
					return Error.InvalidParam("value of add op", "to be complex (for implicit path)", "non-complex")
				}
			}
		case Replace:
			if patch.Value == nil {
				return Error.InvalidParam("value of replace op", "to be present", "nil")
			} else if len(patch.Path) == 0 {
				return Error.InvalidParam("path", "to be present", "empty")
			}
		case Remove:
			if patch.Value != nil {
				return Error.InvalidParam("value of remove op", "to be nil", "non-nil")
			} else if len(patch.Path) == 0 {
				return Error.InvalidParam("path", "to be present", "empty")
			}

		default:
			return Error.InvalidParam("Op", "one of [add|remove|replace]", patch.Op)
		}
	}

	return nil
}

// ----------------------------------
// List Response
// ----------------------------------
type ListResponse struct {
	Schemas      []string
	TotalResults int
	ItemsPerPage int
	StartIndex   int
	Resources    []DataProvider
}

type listResponseMarshalHelper struct {
	abstractMarshalHelper
	Data *ListResponse
}

func (h *listResponseMarshalHelper) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteString("[")
	for i, dp := range h.Data.Resources {
		if i > 0 {
			buf.WriteString(",")
		}
		b, err := MarshalJSON(dp, h.Guide, h.Attributes, h.ExcludedAttributes)
		if err != nil {
			return nil, err
		}
		buf.Write(b)
	}
	buf.WriteString("]")

	raw := json.RawMessage(buf.Bytes())
	return json.Marshal(struct {
		Schemas      []string         `json:"schemas"`
		TotalResults int              `json:"totalResults"`
		ItemsPerPage int              `json:"itemsPerPage"`
		StartIndex   int              `json:"startIndex"`
		Resources    *json.RawMessage `json:"Resources"`
	}{
		Schemas:      h.Data.Schemas,
		TotalResults: h.Data.TotalResults,
		ItemsPerPage: h.Data.ItemsPerPage,
		StartIndex:   h.Data.StartIndex,
		Resources:    &raw,
	})
}

// ----------------------------------
// Search Request
// ----------------------------------
type SearchRequest struct {
	Schemas            []string `json:"schemas"`
	Attributes         []string `json:"attributes"`
	ExcludedAttributes []string `json:"excludedAttributes"`
	Filter             string   `json:"filter"`
	SortBy             string   `json:"sortBy"`
	SortOrder          string   `json:"sortOrder"`
	StartIndex         int      `json:"startIndex"`
	Count              int      `json:"count"`
}

func (sr SearchRequest) Ascending() bool {
	switch sr.SortOrder {
	case "ascending", "":
		return true
	default:
		return false
	}
}
func (sr SearchRequest) Validate(guide AttributeSource) error {
	if len(sr.Schemas) != 1 || sr.Schemas[0] != SearchUrn {
		return Error.InvalidParam("search request", "search operation urn", "non-search urn")
	}

	if len(sr.Filter) == 0 {
		return Error.InvalidParam("search request", "query string", "empty string")
	}

	if sr.StartIndex < 1 {
		sr.StartIndex = 1
	}

	if sr.Count < 0 {
		sr.Count = 0
	}

	switch sr.SortOrder {
	case "", "ascending", "descending":
	default:
		return Error.InvalidParam("search request", "[as|des]cending or blank for sortOrder", sr.SortOrder)
	}

	if guide != nil {
		if len(sr.SortBy) > 0 {
			if corrected, err := sr.correctPathCase(sr.SortBy, guide); err != nil {
				return err
			} else {
				sr.SortBy = corrected
			}
		}

		if len(sr.Attributes) > 0 {
			updated := make([]string, 0)
			for _, each := range sr.Attributes {
				if len(each) > 0 {
					if corrected, err := sr.correctPathCase(each, guide); err != nil {
						return err
					} else {
						updated = append(updated, corrected)
					}
				}
			}
			sr.Attributes = updated
		}

		if len(sr.ExcludedAttributes) > 0 {
			updated := make([]string, 0)
			for _, each := range sr.Attributes {
				if len(each) > 0 {
					if corrected, err := sr.correctPathCase(each, guide); err != nil {
						return err
					} else {
						updated = append(updated, corrected)
					}
				}
			}
			sr.ExcludedAttributes = updated
		}
	}

	return nil
}
func (sr SearchRequest) correctPathCase(text string, guide AttributeSource) (string, error) {
	p, err := NewPath(sr.SortBy)
	if err != nil {
		return "", err
	}
	p.CorrectCase(guide, true)
	buf := new(bytes.Buffer)
	i := 0
	for p != nil {
		if i > 0 {
			buf.WriteString(".")

		}
		buf.WriteString(p.Base())
		p = p.Next()
		i++
	}
	return buf.String(), nil
}
