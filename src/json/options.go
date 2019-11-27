package json

import "github.com/imulab/go-scim/src/core"

// Create a new empty JSON serialization option.
func Options() *options {
	return &options{}
}

// Serialization options
type options struct {
	included	[]string
	excluded	[]string
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

	if attr.Mutability() == core.MutabilityWriteOnly {
		return false
	}

	switch attr.Returned() {
	case core.ReturnedAlways:
		return true
	case core.ReturnedNever:
		return false
	case core.ReturnedRequest:
		for _, included := range opt.included {
			if attr.GoesBy(included) {
				return true
			}
		}
		return false
	case core.ReturnedDefault:
		for _, excluded := range opt.excluded {
			if property.Attribute().GoesBy(excluded) {
				return false
			}
		}
		return !property.IsUnassigned()
	default:
		panic("invalid return-ability value")
	}
}