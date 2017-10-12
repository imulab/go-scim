package resource

// Domain design rule of thumb
// - Use struct for nested objects
// - Use slice of pointers to struct for arrays

type User struct {
	Core
	Username 		string 			`scim:"urn:ietf:params:scim:schemas:core:2.0:User:userName"`
	Name 			Name 			`scim:"urn:ietf:params:scim:schemas:core:2.0:User:name"`
	DisplayName 	string 			`scim:"urn:ietf:params:scim:schemas:core:2.0:User:displayName"`
	NickName 		string 			`scim:"urn:ietf:params:scim:schemas:core:2.0:User:nickName"`
	ProfileUrl 		string 			`scim:"urn:ietf:params:scim:schemas:core:2.0:User:profileUrl"`
	Title 			string 			`scim:"urn:ietf:params:scim:schemas:core:2.0:User:title"`
	UserType 		string 			`scim:"urn:ietf:params:scim:schemas:core:2.0:User:userType"`
	Language 		string 			`scim:"urn:ietf:params:scim:schemas:core:2.0:User:preferredLanguage"`
	Locale 			string 			`scim:"urn:ietf:params:scim:schemas:core:2.0:User:locale"`
	Timezone 		string 			`scim:"urn:ietf:params:scim:schemas:core:2.0:User:timezone"`
	Active 			bool 			`scim:"urn:ietf:params:scim:schemas:core:2.0:User:active"`
	Password 		string 			`scim:"urn:ietf:params:scim:schemas:core:2.0:User:password"`
	Emails 			[]*Email		`scim:"urn:ietf:params:scim:schemas:core:2.0:User:emails"`
	PhoneNumbers	[]*PhoneNumber	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:phoneNumbers"`
	Ims 			[]*Ims			`scim:"urn:ietf:params:scim:schemas:core:2.0:User:ims"`
	Photos 			[]*Photo 		`scim:"urn:ietf:params:scim:schemas:core:2.0:User:photos"`
	Addresses 		[]*Address 		`scim:"urn:ietf:params:scim:schemas:core:2.0:User:addresses"`
	Groups 			[]*GroupRef 	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:groups"`
	Entitlement 	[]*Entitlement 	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:entitlements"`
	Role 			[]*Role 		`scim:"urn:ietf:params:scim:schemas:core:2.0:User:roles"`
	X509 			[]*X509 		`scim:"urn:ietf:params:scim:schemas:core:2.0:User:x509Certificates"`
}

type Name struct {
	Formatted 	string 	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:name.formatted"`
	FamilyName 	string 	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:name.familyName"`
	GivenName 	string 	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:name.givenName"`
	MiddleName 	string 	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:name.middleName"`
	Prefix 		string 	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:name.honorificPrefix"`
	Suffix 		string 	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:name.honorificSuffix"`
}

type Email struct {
	Value 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:emails.value"`
	Display 	string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:emails.display"`
	Type 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:emails.type"`
	Primary 	bool	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:emails.primary"`
}

type PhoneNumber struct {
	Value 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:phoneNumbers.value"`
	Display 	string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:phoneNumbers.display"`
	Type 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:phoneNumbers.type"`
	Primary 	bool	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:phoneNumbers.primary"`
}

type Ims struct {
	Value 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:ims.value"`
	Display 	string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:ims.display"`
	Type 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:ims.type"`
	Primary 	bool	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:ims.primary"`
}

type Photo struct {
	Value 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:photos.value"`
	Display 	string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:photos.display"`
	Type 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:photos.type"`
	Primary 	bool	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:photos.primary"`
}

type Address struct {
	Formatted 		string 	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:addresses.formatted"`
	StreetAddress 	string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:addresses.streetAddress"`
	Locality 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:addresses.locality"`
	Region 			string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:addresses.region"`
	PostalCode 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:addresses.postalCode"`
	Country 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:addresses.country"`
	Type 			string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:addresses.type"`
	Primary 		bool	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:addresses.primary"`
}

type GroupRef struct {
	Value 	string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:groups.value"`
	Ref 	string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:groups.$ref"`
	Display string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:groups.display"`
	Type 	string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:groups.type"`
}

type Entitlement struct {
	Value 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:entitlements.value"`
	Display 	string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:entitlements.display"`
	Type 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:entitlements.type"`
	Primary 	bool	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:entitlements.primary"`
}

type Role struct {
	Value 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:roles.value"`
	Display 	string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:roles.display"`
	Type 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:roles.type"`
	Primary 	bool	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:roles.primary"`
}

type X509 struct {
	Value 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:x509Certificates.value"`
	Display 	string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:x509Certificates.display"`
	Type 		string	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:x509Certificates.type"`
	Primary 	bool	`scim:"urn:ietf:params:scim:schemas:core:2.0:User:x509Certificates.primary"`
}