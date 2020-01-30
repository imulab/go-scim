package spec

// Service provider config
type ServiceProviderConfig struct {
	Schemas []string `json:"schemas"`
	DocURI  string   `json:"documentationUri"`
	Patch   struct {
		Supported bool `json:"supported"`
	} `json:"patch"`
	Bulk struct {
		Supported  bool `json:"supported"`
		MaxOp      int  `json:"maxOperations"`
		MaxPayload int  `json:"maxPayloadSize"`
	} `json:"bulk"`
	Filter struct {
		Supported  bool `json:"supported"`
		MaxResults int  `json:"maxResults"`
	} `json:"filter"`
	ChangePassword struct {
		Supported bool `json:"supported"`
	} `json:"changePassword"`
	Sort struct {
		Supported bool `json:"supported"`
	} `json:"sort"`
	ETag struct {
		Supported bool `json:"supported"`
	} `json:"etag"`
	AuthSchemes []struct {
		Type        string `json:"type"`
		Name        string `json:"name"`
		Description string `json:"description"`
		SpecURI     string `json:"specUri"`
		DocURI      string `json:"documentationUri"`
	} `json:"authenticationSchemes"`
}
