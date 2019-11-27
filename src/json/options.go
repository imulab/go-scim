package json

import "github.com/imulab/go-scim/src/core"

// Create a new empty JSON serialization option.
func Options() *options {
	return &options{}
}

// Serialization options
type options struct {
	included []string
	excluded []string
}

// Specify included attributes to the options.
func (opt *options) Include(fields ...string) *options {
	if opt.included == nil {
		opt.included = []string{}
	}
	opt.included = append(opt.included, fields...)
	return opt
}

// Specify excluded attributes to the options.
func (opt *options) Exclude(fields ...string) *options {
	if opt.excluded == nil {
		opt.excluded = []string{}
	}
	opt.excluded = append(opt.excluded, fields...)
	return opt
}

// Returns true if the property should be serialized.
func (opt *options) shouldSerialize(property core.Property) bool {
	attr := property.Attribute()

	// Write only properties are never returned. It is usually coupled
	// with returned=never, but we will check it to make sure.
	if attr.Mutability() == core.MutabilityWriteOnly {
		return false
	}

	if attr.Returned() == core.ReturnedAlways {
		return true
	}

	if attr.Returned() == core.ReturnedNever {
		return false
	}

	// Request-returned attributes are only returned when
	// it is not excluded and it is indeed included.
	if attr.Returned() == core.ReturnedRequest {
		for _, excluded := range opt.excluded {
			if property.Attribute().GoesBy(excluded) {
				return false
			}
		}
		for _, included := range opt.included {
			if attr.GoesBy(included) {
				return true
			}
		}
		return false
	}

	// Default-returned attributes are returned only when
	// it is not excluded and there's no included attributes (otherwise, only
	// those are returned), and the property itself is not unassigned.
	if attr.Returned() == core.ReturnedDefault {
		for _, excluded := range opt.excluded {
			if property.Attribute().GoesBy(excluded) {
				return false
			}
		}
		for _, included := range opt.included {
			if attr.GoesBy(included) {
				return true
			}
		}
		return len(opt.included) == 0 && !property.IsUnassigned()
	}

	return false // impossible to reach here
}
