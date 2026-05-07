package internal

type ResourceTypeJsonAdapter struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Endpoint    string             `json:"endpoint"`
	Schema      string             `json:"schema"`
	Extensions  []*SchemaExtension `json:"schemaExtensions,omitempty"`
}

type SchemaExtension struct {
	Schema   string `json:"schema"`
	Required bool   `json:"required"`
}
