package internal

// adapter to marshal the attribute
type AttributeMarshaler struct {
	Name            string                `json:"name"`
	Description     string                `json:"description,omitempty"`
	Type            string                `json:"type"`
	SubAttributes   []*AttributeMarshaler `json:"subAttributes,omitempty"`
	CanonicalValues []string              `json:"canonicalValues,omitempty"`
	MultiValued     bool                  `json:"multiValued"`
	Required        bool                  `json:"required"`
	CaseExact       bool                  `json:"caseExact"`
	Mutability      string                `json:"mutability"`
	Returned        string                `json:"returned"`
	Uniqueness      string                `json:"uniqueness"`
	ReferenceTypes  []string              `json:"referenceTypes,omitempty"`
}

// adapter to unmarshal the attribute
type AttributeUnmarshaler struct {
	ID              string                            `json:"id"`
	Name            string                            `json:"name"`
	Description     string                            `json:"description"`
	Type            string                            `json:"type"`
	SubAttributes   []*AttributeUnmarshaler           `json:"subAttributes"`
	CanonicalValues []string                          `json:"canonicalValues"`
	MultiValued     bool                              `json:"multiValued"`
	Required        bool                              `json:"required"`
	CaseExact       bool                              `json:"caseExact"`
	Mutability      string                            `json:"mutability"`
	Returned        string                            `json:"returned"`
	Uniqueness      string                            `json:"uniqueness"`
	ReferenceTypes  []string                          `json:"referenceTypes"`
	Index           int                               `json:"_index"`
	Path            string                            `json:"_path"`
	Annotations     map[string]map[string]interface{} `json:"_annotations"`
}
