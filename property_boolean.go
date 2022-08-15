package scim

const (
	trueHash  uint64 = 9929646806074584996
	falseHash uint64 = 12161962213042174405
)

type booleanProperty struct {
	attr *Attribute
	vb   *bool
}

func (p *booleanProperty) Attr() *Attribute {
	return p.attr
}

func (p *booleanProperty) Value() any {
	if p.Unassigned() {
		return nil
	}
	return *p.vb
}

func (p *booleanProperty) Unassigned() bool {
	return p.vb == nil
}

func (p *booleanProperty) Add(value any) error {
	return p.Set(value)
}

func (p *booleanProperty) Set(value any) error {
	if value == nil {
		p.Delete()
		return nil
	}

	b, ok := value.(bool)
	if !ok {
		return ErrInvalidValue
	}

	p.vb = &b

	return nil
}

func (p *booleanProperty) Delete() {
	p.vb = nil
}

func (p *booleanProperty) Hash() uint64 {
	if p.Unassigned() {
		return 0
	}

	if *p.vb {
		return trueHash
	} else {
		return falseHash
	}
}

func (p *booleanProperty) Len() int {
	return 0
}

func (p *booleanProperty) ForEach(_ func(index int, child Property) error) error {
	return nil
}

func (p *booleanProperty) Find(_ func(child Property) bool) Property {
	return nil
}

func (p *booleanProperty) ByIndex(_ any) Property {
	return nil
}

func (p *booleanProperty) equalsTo(value any) bool {
	if p.Unassigned() || value == nil {
		return false
	}

	b, ok := value.(bool)
	if !ok {
		return false
	}

	return *p.vb == b
}

func (p *booleanProperty) isPresent() bool {
	return !p.Unassigned()
}
