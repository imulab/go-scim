package core

const (
	// SCIM uniqueness attribute defined in RFC7643
	UniquenessNone Uniqueness = iota
	UniquenessServer
	UniquenessGlobal
)
// SCIM uniqueness definition
type Uniqueness int

// Return a standard definition of this uniqueness.
func (u Uniqueness) String() string {
	switch u {
	case UniquenessNone:
		return "none"
	case UniquenessServer:
		return "server"
	case UniquenessGlobal:
		return "global"
	default:
		panic("invalid uniqueness")
	}
}