package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// ----------------------------------
// Bulk
// ----------------------------------
type bulkOp struct {
	Method  string `json:"method"`
	BulkId  string `json:"bulkId"`
	Version string `json:"version"`
}

type BulkReq struct {
	Schemas      []string    `json:"schemas"`
	FailOnErrors int         `json:"failOnErrors"`
	Operations   []BulkReqOp `json:"Operations"`
}

type BulkReqOp struct {
	bulkOp
	Path string          `json:"path"`
	Data json.RawMessage `json:"data"`
}

func (br BulkReq) Validate(ps PropertySource) error {
	if len(br.Schemas) != 1 || br.Schemas[0] != BulkRequestUrn {
		return Error.InvalidParam("schema", BulkRequestUrn, "other content")
	}

	for _, op := range br.Operations {
		if err := op.validate(ps); err != nil {
			return err
		}
	}

	return nil
}

func (op BulkReqOp) validate(ps PropertySource) error {
	switch strings.ToUpper(op.Method) {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return Error.InvalidParam("method", "one of ['post','put','patch','delete']", op.Method)
	}

	userUri := ps.GetString("scim.protocol.uri.user")
	groupUri := ps.GetString("scim.protocol.uri.group")
	if http.MethodPost == strings.ToUpper(op.Method) {
		switch op.Path {
		case userUri, groupUri:
		default:
			return Error.InvalidParam("path", fmt.Sprintf("one of ['%s','%s']", userUri, groupUri), op.Path)
		}
	} else {
		switch {
		case strings.HasPrefix(op.Path, userUri+"/"):
		case strings.HasPrefix(op.Path, groupUri+"/"):
		default:
			return Error.InvalidParam("path", fmt.Sprintf("one of ['%s/<id>','%s/<id>']", userUri, groupUri), op.Path)
		}
	}

	switch strings.ToUpper(op.Method) {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		if len(op.Data) == 0 {
			return Error.InvalidParam("data", "to be present", "nothing")
		}
	}

	return nil
}

type BulkResp struct {
	Schemas    []string      `json:"schemas"`
	Operations []*BulkRespOp `json:"Operations"`
}

type BulkRespOp struct {
	bulkOp
	Location string          `json:"location"`
	Response json.RawMessage `json:"response,omitempty"`
	Status   int             `json:"status"`
}

func (bro BulkRespOp) Populate(origReq BulkReqOp, resp WebResponse) {
	bro.Method = strings.ToLower(origReq.Method)
	bro.BulkId = origReq.BulkId
	bro.Version = resp.GetHeader("ETag")
	bro.Location = resp.GetHeader("Location")
	bro.Status = resp.GetStatus()
	if resp.GetStatus() > 299 {
		bro.Response = json.RawMessage(resp.GetBody())
	} else {
		bro.Response = json.RawMessage{}
	}
}

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

	for i := 0; i < len(m.Ops); i++ {
		m.Ops[i].Op = strings.ToLower(m.Ops[i].Op)
		switch m.Ops[i].Op {
		case Add:
			if m.Ops[i].Value == nil {
				return Error.InvalidParam("value of add op", "to be present", "nil")
			} else if len(m.Ops[i].Path) == 0 {
				if _, ok := m.Ops[i].Value.(map[string]interface{}); !ok {
					return Error.InvalidParam("value of add op", "to be complex (for implicit path)", "non-complex")
				}
			}
		case Replace:
			if m.Ops[i].Value == nil {
				return Error.InvalidParam("value of replace op", "to be present", "nil")
			} else if len(m.Ops[i].Path) == 0 {
				return Error.InvalidParam("path", "to be present", "empty")
			}
		case Remove:
			if m.Ops[i].Value != nil {
				return Error.InvalidParam("value of remove op", "to be nil", "non-nil")
			} else if len(m.Ops[i].Path) == 0 {
				return Error.InvalidParam("path", "to be present", "empty")
			}
		default:
			return Error.InvalidParam("Op", "one of [add|remove|replace]", m.Ops[i].Op)
		}
		m.Ops[i].Path = strings.ToLower(m.Ops[i].Path)
		if strings.HasPrefix(m.Ops[i].Path, "/") {
			m.Ops[i].Path = m.Ops[i].Path[1:]
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
	p, err := NewPath(text)
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
