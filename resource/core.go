package resource

type Core struct {
	Schemas 	[]string 	`scim:"schemas"`
	Id 			string 		`scim:"id"`
	ExternalId 	string 		`scim:"externalId"`
	Meta 		Meta 		`scim:"meta"`
}

type Meta struct {
	ResourceType 	string 	`scim:"meta.resourceType"`
	Created 		string 	`scim:"meta.created"`
	LastModified 	string 	`scim:"meta.lastModified"`
	Location 		string 	`scim:"meta.location"`
	Version 		string 	`scim:"meta.version"`
}