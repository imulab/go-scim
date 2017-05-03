package shared

type RequestId struct{}
type ResourceId struct{}
type RequestTimestamp struct{}
type RequestType struct{}

const (
	_ = iota
	GetUserById
	CreateUser
	ReplaceUser
	PatchUser
	QueryUser
	DeleteUser
	GetGroupById
	CreateGroup
	ReplaceGroup
	PatchGroup
	QueryGroup
	DeleteGroup
	RootQuery
	BulkOp
	GetSchemaById
	GetAllSchema
	GetSPConfig
	GetAllResourceType
)

type WebRequest interface {
	Target() string
	Method() string
	Header(name string) string
	Param(name string) string
	Body() ([]byte, error)
}

type WebResponse interface {
	GetStatus() int
	GetHeader(name string) string
	GetBody() []byte
}
