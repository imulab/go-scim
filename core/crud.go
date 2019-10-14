package core

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// An extension to the Property interface to enable create-read-update-delete operations. The new value
// must be compatible with the property's attribute for the operation to be successful. Mutability and required
// rules are not enforced here. For simple properties, the step must be nil; for complex properties, the step
// may entail a nested path.
type Crud interface {
	Property

	// Get the property at the specified step. If the property does not exist the specified step, returns error.
	Get(step *step) (interface{}, error)

	// Add / create a value to this property at the specified step. For simple properties,
	// the operation is equivalent to a Replace operation.
	Add(step *step, value interface{}) error

	// Replace a value to this property at the specified step.
	Replace(step *step, value interface{}) error

	// Delete a value to this property at the specified step.
	Delete(step *step) error
}

func (s *stringProperty) Get(step *step) (interface{}, error) {
	if step != nil {
		return nil, s.attr.errNoTarget(step)
	}
	return s.Raw(), nil
}

func (s *stringProperty) Add(step *step, value interface{}) error {
	if value == nil {
		return s.Delete(step)
	}
	return s.Replace(step, value)
}

func (s *stringProperty) Replace(step *step, value interface{}) error {
	if value == nil {
		return s.Delete(step)
	}

	if step != nil {
		return s.attr.errNoTarget(step)
	}

	if str, ok := value.(string); !ok {
		return s.attr.errInvalidValue()
	} else {
		s.v = &str
		return nil
	}
}

func (s *stringProperty) Delete(step *step) error {
	if step != nil {
		return s.attr.errNoTarget(step)
	}
	s.v = nil
	return nil
}

func (i *integerProperty) Get(step *step) (interface{}, error) {
	if step != nil {
		return nil, i.attr.errNoTarget(step)
	}
	return i.Raw(), nil
}

func (i *integerProperty) Add(step *step, value interface{}) error {
	if value == nil {
		return i.Delete(step)
	}
	return i.Replace(step, value)
}

func (i *integerProperty) Replace(step *step, value interface{}) error {
	if value == nil {
		return i.Delete(step)
	}

	if step != nil {
		return i.attr.errNoTarget(step)
	}

	var v int64
	{
		switch value.(type) {
		case int:
			v = int64(value.(int))
		case int32:
			v = int64(value.(int32))
		case int64:
			v = value.(int64)
		default:
			return i.attr.errInvalidValue()
		}
	}

	i.v = &v
	return nil
}

func (i *integerProperty) Delete(step *step) error {
	if step != nil {
		return i.attr.errNoTarget(step)
	}
	i.v = nil
	return nil
}

func (d *decimalProperty) Get(step *step) (interface{}, error) {
	if step != nil {
		return nil, d.attr.errNoTarget(step)
	}
	return d.Raw(), nil
}

func (d *decimalProperty) Add(step *step, value interface{}) error {
	if value == nil {
		return d.Delete(step)
	}
	return d.Replace(step, value)
}

func (d *decimalProperty) Replace(step *step, value interface{}) error {
	if value == nil {
		return d.Delete(step)
	}

	if step != nil {
		return d.attr.errNoTarget(step)
	}

	var v float64
	{
		switch value.(type) {
		case float32:
			v = float64(value.(float32))
		case float64:
			v = value.(float64)
		default:
			return d.attr.errInvalidValue()
		}
	}

	d.v = &v
	return nil
}

func (d *decimalProperty) Delete(step *step) error {
	if step != nil {
		return d.attr.errNoTarget(step)
	}
	d.v = nil
	return nil
}

func (b *booleanProperty) Get(step *step) (interface{}, error) {
	if step != nil {
		return nil, b.attr.errNoTarget(step)
	}
	return b.Raw(), nil
}

func (b *booleanProperty) Add(step *step, value interface{}) error {
	if value == nil {
		return b.Delete(step)
	}
	return b.Replace(step, value)
}

func (b *booleanProperty) Replace(step *step, value interface{}) error {
	if value == nil {
		return b.Delete(step)
	}

	if step != nil {
		return b.attr.errNoTarget(step)
	}

	if v, ok := value.(bool); !ok {
		return b.attr.errInvalidValue()
	} else {
		b.v = &v
		return nil
	}
}

func (b *booleanProperty) Delete(step *step) error {
	if step != nil {
		return b.attr.errNoTarget(step)
	}
	b.v = nil
	return nil
}

func (d *dateTimeProperty) Get(step *step) (interface{}, error) {
	if step != nil {
		return nil, d.attr.errNoTarget(step)
	}
	return d.Raw(), nil
}

func (d *dateTimeProperty) Add(step *step, value interface{}) error {
	if value == nil {
		return d.Delete(step)
	}
	return d.Replace(step, value)
}

func (d *dateTimeProperty) Replace(step *step, value interface{}) error {
	if value == nil {
		return d.Delete(step)
	}

	if step != nil {
		return d.attr.errNoTarget(step)
	}

	if v, ok := value.(string); !ok {
		return d.attr.errInvalidValue()
	} else if _, err := time.Parse(ISO8601, v); err != nil {
		return d.attr.errInvalidValue()
	} else {
		d.v = &v
		return nil
	}
}

func (d *dateTimeProperty) Delete(step *step) error {
	if step != nil {
		return d.attr.errNoTarget(step)
	}
	d.v = nil
	return nil
}

func (b *binaryProperty) Get(step *step) (interface{}, error) {
	if step != nil {
		return nil, b.attr.errNoTarget(step)
	}
	return b.Raw(), nil
}

func (b *binaryProperty) Add(step *step, value interface{}) error {
	if value == nil {
		return b.Delete(step)
	}
	return b.Replace(step, value)
}

func (b *binaryProperty) Replace(step *step, value interface{}) error {
	if value == nil {
		return b.Delete(step)
	}

	if step != nil {
		return b.attr.errNoTarget(step)
	}

	if v, ok := value.(string); !ok {
		return b.attr.errInvalidValue()
	} else if _, err := base64.StdEncoding.DecodeString(v); err != nil {
		return b.attr.errInvalidValue()
	} else {
		b.v = &v
		return nil
	}
}

func (b *binaryProperty) Delete(step *step) error {
	if step != nil {
		return b.attr.errNoTarget(step)
	}
	b.v = nil
	return nil
}

func (r *referenceProperty) Get(step *step) (interface{}, error) {
	if step != nil {
		return nil, r.attr.errNoTarget(step)
	}
	return r.Raw(), nil
}

func (r *referenceProperty) Add(step *step, value interface{}) error {
	if value == nil {
		return r.Delete(step)
	}
	return r.Replace(step, value)
}

func (r *referenceProperty) Replace(step *step, value interface{}) error {
	if value == nil {
		return r.Delete(step)
	}

	if step != nil {
		return r.attr.errNoTarget(step)
	}

	if v, ok := value.(string); !ok {
		return r.attr.errInvalidValue()
	} else {
		r.v = &v
		return nil
	}
}

func (r *referenceProperty) Delete(step *step) error {
	if step != nil {
		return r.attr.errNoTarget(step)
	}
	r.v = nil
	return nil
}

func (c *complexProperty) Get(step *step) (interface{}, error) {
	if step == nil {
		return c.Raw(), nil
	}

	if step.Typ != stepPath {
		return nil, c.attr.errNoTarget(step)
	}

	idx, ok := c.index[strings.ToLower(step.Token)]
	if !ok {
		return nil, c.attr.errNoTarget(step)
	}

	return c.props[idx].(Crud).Get(step.Next)
}

func (c *complexProperty) Add(step *step, value interface{}) error {
	if step == nil {
		return c.selfAdd(value)
	}

	if step.Typ != stepPath {
		return c.attr.errNoTarget(step)
	}

	idx, ok := c.index[strings.ToLower(step.Token)]
	if !ok {
		return c.attr.errNoTarget(step)
	}

	return c.props[idx].(Crud).Add(step.Next, value)
}

func (c *complexProperty) selfAdd(value interface{}) error {
	if value == nil {
		return c.attr.errInvalidValue()
	}

	m, ok := value.(map[string]interface{})
	if !ok {
		return c.attr.errInvalidValue()
	}

	for k, v := range m {
		idx, ok := c.index[strings.ToLower(k)]
		if !ok {
			return c.attr.errNoTarget(Steps.NewPath(c.attr.DisplayName() + "." + k))
		}

		if err := c.props[idx].(Crud).Add(nil, v); err != nil {
			return err
		}
	}

	return nil
}

func (c *complexProperty) Replace(step *step, value interface{}) error {
	if step == nil {
		return c.selfReplace(value)
	}

	if step.Typ != stepPath {
		return c.attr.errNoTarget(step)
	}

	idx, ok := c.index[strings.ToLower(step.Token)]
	if !ok {
		return c.attr.errNoTarget(step)
	}

	return c.props[idx].(Crud).Replace(step.Next, value)
}

func (c *complexProperty) selfReplace(value interface{}) error {
	if value == nil {
		return c.Delete(nil)
	}

	m, ok := value.(map[string]interface{})
	if !ok {
		return c.attr.errInvalidValue()
	}

	m0 := make(map[string]interface{})
	for k, v := range m {
		m0[strings.ToLower(k)] = v
	}

	for _, sub := range c.props {
		if nv, ok := m0[strings.ToLower(sub.Attribute().Name)]; ok && nv != nil {
			if err := sub.(Crud).Replace(nil, nv); err != nil {
				return err
			}
		} else {
			if err := sub.(Crud).Delete(nil); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *complexProperty) Delete(step *step) error {
	if step == nil {
		return c.selfDelete()
	}

	if step.Typ != stepPath {
		return c.attr.errNoTarget(step)
	}

	idx, ok := c.index[strings.ToLower(step.Token)]
	if !ok {
		return c.attr.errNoTarget(step)
	}

	return c.props[idx].(Crud).Delete(step.Next)
}

func (c *complexProperty) selfDelete() error {
	for _, sub := range c.props {
		if err := sub.(Crud).Delete(nil); err != nil {
			return err
		}
	}

	if !c.IsUnassigned() {
		return Errors.internal(
			fmt.Sprintf("invalid state: %s should be unassigned after deletion", c.attr.DisplayName()),
		)
	}

	return nil
}

func (m *multiValuedProperty) Get(step *step) (interface{}, error) {
	if step == nil {
		return m, nil
	}

	var results []interface{}
	{
		switch {
		case step.IsPath():
			results = make([]interface{}, 0)
			for i, elem := range m.props {
				if elemV, err := elem.(Crud).Get(step); err != nil {
					return nil, Errors.noTarget(err.Error() + fmt.Sprintf(" (hint: idx=%d)", i))
				} else if elemV != nil {
					results = append(results, elemV)
				}
			}

		case step.IsOperator():
			results = make([]interface{}, 0)
			for i, elem := range m.props {
				hint := fmt.Sprintf("(hint: idx=%d)", i)

				evalElem, ok := elem.(Evaluation)
				if !ok {
					return nil, m.attr.errNoTarget(step)
				}

				if b, err := evalElem.Evaluate(step); err != nil {
					return nil, ErrAppendHint(err, hint)
				} else if !b {
					continue
				}

				if elemV, err := elem.(Crud).Get(step.Next); err != nil {
					return nil, ErrAppendHint(err, hint)
				} else if elemV != nil {
					results = append(results, elemV)
				}
			}

		default:
			return nil, m.attr.errNoTarget(step)
		}
	}

	return results, nil
}

func (m *multiValuedProperty) Add(step *step, value interface{}) error {
	if step == nil {
		return m.selfAdd(value)
	}

	switch {
	case step.IsPath():
		for i, elem := range m.props {
			if err := elem.(Crud).Add(step, value); err != nil {
				return ErrAppendHint(err, fmt.Sprintf("(hint: idx=%d)", i))
			}
		}

	case step.IsOperator():
		for i, elem := range m.props {
			hint := fmt.Sprintf("(hint: idx=%d)", i)

			evalElem, ok := elem.(Evaluation)
			if !ok {
				return m.attr.errNoTarget(step)
			}

			if b, err := evalElem.Evaluate(step); err != nil {
				return ErrAppendHint(err, hint)
			} else if !b {
				continue
			}

			if err := elem.(Crud).Add(step.Next, value); err != nil {
				return ErrAppendHint(err, hint)
			}
		}

	default:
		return m.attr.errNoTarget(step)
	}

	m.compact()
	m.updateExclusive()
	return nil
}

func (m *multiValuedProperty) selfAdd(value interface{}) error {
	if value == nil {
		return m.attr.errInvalidValue()
	}

	elem := Properties.New(m.attr.ToSingleValued())
	if err := elem.(Crud).Replace(nil, value); err != nil {
		return m.attr.errInvalidValue()
	}

	m.props = append(m.props, elem)
	return nil
}

func (m *multiValuedProperty) Replace(step *step, value interface{}) error {
	if step == nil {
		return m.selfReplace(value)
	}

	switch {
	case step.IsPath():
		for i, elem := range m.props {
			if err := elem.(Crud).Replace(step, value); err != nil {
				return ErrAppendHint(err, fmt.Sprintf("(hint: idx=%d)", i))
			}
		}

	case step.IsOperator():
		for i, elem := range m.props {
			hint := fmt.Sprintf("(hint: idx=%d)", i)

			evalElem, ok := elem.(Evaluation)
			if !ok {
				return m.attr.errNoTarget(step)
			}

			if b, err := evalElem.Evaluate(step); err != nil {
				return ErrAppendHint(err, hint)
			} else if !b {
				continue
			}

			if err := elem.(Crud).Replace(step.Next, value); err != nil {
				return ErrAppendHint(err, hint)
			}
		}

	default:
		return m.attr.errNoTarget(step)
	}

	m.compact()
	m.updateExclusive()
	return nil
}

func (m *multiValuedProperty) selfReplace(value interface{}) error {
	if value == nil {
		return m.Delete(nil)
	}

	array, ok := value.([]interface{})
	if !ok {
		return m.attr.errInvalidValue()
	}

	m.props = make([]Property, 0)
	for _, v := range array {
		elem := Properties.New(m.attr.ToSingleValued())
		if err := elem.(Crud).Replace(nil, v); err != nil {
			return m.attr.errInvalidValue()
		}
		m.props = append(m.props, elem)
	}

	return nil
}

func (m *multiValuedProperty) Delete(step *step) error {
	if step == nil {
		m.props = make([]Property, 0)
		return nil
	}

	switch {
	case step.IsPath():
		for i, elem := range m.props {
			if err := elem.(Crud).Delete(step); err != nil {
				return ErrAppendHint(err, fmt.Sprintf("(hint: idx=%d)", i))
			}
		}

	case step.IsOperator():
		for i, elem := range m.props {
			hint := fmt.Sprintf("(hint: idx=%d)", i)

			evalElem, ok := elem.(Evaluation)
			if !ok {
				return m.attr.errNoTarget(step)
			}

			if b, err := evalElem.Evaluate(step); err != nil {
				return ErrAppendHint(err, hint)
			} else if !b {
				continue
			}

			if err := elem.(Crud).Delete(step.Next); err != nil {
				return ErrAppendHint(err, hint)
			}
		}

	default:
		return m.attr.errNoTarget(step)
	}

	m.compact()
	m.updateExclusive()
	return nil
}

func (m *multiValuedProperty) compact() {
	if len(m.props) == 0 {
		return
	}

	var i int
	for i = len(m.props) - 1; i >= 0; i-- {
		if m.props[i].IsUnassigned() {
			if i == len(m.props)-1 {
				m.props = m.props[:i]
			} else if i == 0 {
				m.props = m.props[i+1:]
			} else {
				m.props = append(m.props[:i], m.props[i+1:]...)
			}
		}
	}
}

// Handles the exclusivity rule that at most one 'exclusive' boolean sub attribute can have the value of true.
// This method needs to be called after any potential state modification for multiValued properties.
func (m *multiValuedProperty) updateExclusive() {
	// short circuit not a complex attribute
	if m.attr.Type != TypeComplex {
		return
	}

	// short circuit if the complex attribute has no exclusive boolean sub attribute
	hasExcl := false
	for _, subAttr := range m.attr.SubAttributes {
		if subAttr.Type == TypeBoolean && subAttr.Metadata != nil && subAttr.Metadata.IsExclusive {
			hasExcl = true
			break
		}
	}
	if !hasExcl {
		return
	}

	// ranging through the elements, if having a new exclusive boolean property, switch off the old exclusive
	// boolean property; if no exclusive boolean property at all, clear the cache.
	hasExcl = false
	for _, elem := range m.props {
		if excl := elem.(*complexProperty).getExclusiveTrue(); excl != nil {
			hasExcl = true
			if m.excl != nil && m.excl != excl {
				_ = m.excl.Replace(nil, false)
			}
			m.excl = excl
		}
	}
	if !hasExcl {
		m.excl = nil
	}
}

var (
	// implementation checks
	_ Crud = (*stringProperty)(nil)
	_ Crud = (*integerProperty)(nil)
	_ Crud = (*decimalProperty)(nil)
	_ Crud = (*booleanProperty)(nil)
	_ Crud = (*dateTimeProperty)(nil)
	_ Crud = (*binaryProperty)(nil)
	_ Crud = (*referenceProperty)(nil)
	_ Crud = (*complexProperty)(nil)
)
