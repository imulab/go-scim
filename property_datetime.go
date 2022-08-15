package scim

import "time"

const (
	formatISO8601 = "2006-01-02T15:04:05"
)

type dateTimeProperty struct {
	*stringProperty
}

func (p *dateTimeProperty) Set(value any) error {
	s, ok := value.(string)
	if !ok {
		return ErrInvalidValue
	}

	if _, err := time.Parse(formatISO8601, s); err != nil {
		return ErrInvalidValue
	}

	return p.stringProperty.Set(s)
}

func (p *dateTimeProperty) equalsTo(value any) bool {
	r, ok := p.compare(value)
	return ok && r == 0
}

func (p *dateTimeProperty) greaterThan(value any) bool {
	r, ok := p.compare(value)
	return ok && r > 0
}

func (p *dateTimeProperty) lessThan(value any) bool {
	r, ok := p.compare(value)
	return ok && r < 0
}

func (p *dateTimeProperty) compare(value any) (int, bool) {
	if p.Unassigned() || value == nil {
		return 0, false
	}

	s, ok := value.(string)
	if !ok {
		return 0, false
	}

	if this, err := time.Parse(formatISO8601, *p.vs); err != nil {
		return 0, false
	} else if that, err := time.Parse(formatISO8601, s); err != nil {
		return 0, false
	} else {
		switch {
		case this.Before(that):
			return -1, true
		case this.After(that):
			return 1, true
		case this.Equal(that):
			return 0, true
		default:
			panic("impossible case")
		}
	}
}
