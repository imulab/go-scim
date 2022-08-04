package scim

type Property interface {
	Value() any
	Unassigned() bool
}
