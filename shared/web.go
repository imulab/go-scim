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
)
