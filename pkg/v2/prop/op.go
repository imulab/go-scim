package prop

// Capability to perform 'eq' operations, and by logic, 'ne' operations.
type EqCapable interface {
	// EqualsTo return true if the property's value is equal to the given value.
	// If the given value is nil, always return false.
	EqualsTo(value interface{}) bool
}

// Capability to perform 'sw' operations.
type SwCapable interface {
	// StartsWith return true if the property's value starts with the given value.
	// If the given value is nil, always return false.
	StartsWith(value string) bool
}

// Capability to perform 'ew' operations.
type EwCapable interface {
	// EndsWith return true if the property's value ends with the given value.
	// If the given value is nil, always return false.
	EndsWith(value string) bool
}

// Capability to perform 'co' operations.
type CoCapable interface {
	// Contains return true if the property's value contains the given value.
	// If the given value is nil, always return false.
	Contains(value string) bool
}

// Capability to perform 'gt' operations.
type GtCapable interface {
	// GreaterThan return true if the property's value is greater than the given value.
	// If the given value is nil, always return false.
	GreaterThan(value interface{}) bool
}

// Capability to perform 'ge' operations.
type GeCapable interface {
	// GreaterThanOrEqualTo return true if the property's value is greater than or equal to the given value.
	// If the given value is nil, always return false.
	GreaterThanOrEqualTo(value interface{}) bool
}

// Capability to perform 'lt' operations.
type LtCapable interface {
	// LessThan return true if the property's value is greater than the given value.
	// If the given value is nil, always return false.
	LessThan(value interface{}) bool
}

// Capability to perform 'le' operations.
type LeCapable interface {
	// LessThanOrEqualTo return true if the property's value is greater than or equal to the given value.
	// If the given value is nil, always return false.
	LessThanOrEqualTo(value interface{}) bool
}

// Capability to perform 'pr' operations.
type PrCapable interface {
	// Present return true if the property's value is present. Presence is defined to be non-nil and non-empty.
	Present() bool
}
