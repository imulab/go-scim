package prop

// EqCapable defines the capability to perform 'eq' operations, and by logic, 'ne' operations. It should be implemented
// by capable Property implementations.
type EqCapable interface {
	// EqualsTo return true if the property's value is equal to the given value.
	// If the given value is nil, always return false.
	EqualsTo(value interface{}) bool
}

// SwCapable defines capability to perform 'sw' operations. It should be implemented by capable Property implementations.
type SwCapable interface {
	// StartsWith return true if the property's value starts with the given value.
	// If the given value is nil, always return false.
	StartsWith(value string) bool
}

// EwCapable defines capability to perform 'ew' operations. It should be implemented by capable Property implementations.
type EwCapable interface {
	// EndsWith return true if the property's value ends with the given value.
	// If the given value is nil, always return false.
	EndsWith(value string) bool
}

// CoCapable defines capability to perform 'co' operations. It should be implemented by capable Property implementations.
type CoCapable interface {
	// Contains return true if the property's value contains the given value.
	// If the given value is nil, always return false.
	Contains(value string) bool
}

// GtCapable defines capability to perform 'gt' operations. It should be implemented by capable Property implementations.
type GtCapable interface {
	// GreaterThan return true if the property's value is greater than the given value.
	// If the given value is nil, always return false.
	GreaterThan(value interface{}) bool
}

// GeCapable defines capability to perform 'ge' operations. It should be implemented by capable Property implementations.
type GeCapable interface {
	// GreaterThanOrEqualTo return true if the property's value is greater than or equal to the given value.
	// If the given value is nil, always return false.
	GreaterThanOrEqualTo(value interface{}) bool
}

// LtCapable defines capability to perform 'lt' operations. It should be implemented by capable Property implementations.
type LtCapable interface {
	// LessThan return true if the property's value is greater than the given value.
	// If the given value is nil, always return false.
	LessThan(value interface{}) bool
}

// LeCapable defines capability to perform 'le' operations. It should be implemented by capable Property implementations.
type LeCapable interface {
	// LessThanOrEqualTo return true if the property's value is greater than or equal to the given value.
	// If the given value is nil, always return false.
	LessThanOrEqualTo(value interface{}) bool
}

// PrCapable defines capability to perform 'pr' operations. It should be implemented by capable Property implementations.
type PrCapable interface {
	// Present return true if the property's value is present. Presence is defined to be non-nil and non-empty.
	Present() bool
}
