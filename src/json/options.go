package json

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