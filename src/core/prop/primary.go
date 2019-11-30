package prop

import (
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
)

func NewPrimaryMonitor(resource *Resource) *PrimaryMonitor {
	monitor := new(PrimaryMonitor)
	monitor.init(resource)
	return monitor
}

type PrimaryMonitor struct {
	// all multiValued complex containers that has
	// primary boolean sub properties.
	containerProps []core.Container
	// A map of every attribute id of the container properties to
	// the list of boolean primary sub properties whose value is
	// set to true. The allowed size of this list if 0 or 1 after
	// adjusting.
	state map[string][]*booleanProperty
}

func (m *PrimaryMonitor) init(resource *Resource) {
	_ = resource.Visit(m)
}

func (m *PrimaryMonitor) Scan() (err error) {
	newState := m.refresh()

	if m.state == nil {
		err = m.noOldState(newState)
	} else {
		for k := range newState {
			switch len(m.state[k]) {
			case 0:
				err = m.noPrimaryInOldState(k, newState)
			case 1:
				err = m.onePrimaryInOldState(k, newState)
			default:
				panic("invalid state")
			}
			if err != nil {
				break
			}
		}
	}
	if err != nil {
		return
	}

	m.state = newState
	return
}

func (m *PrimaryMonitor) noOldState(newState map[string][]*booleanProperty) error {
	for k, v := range newState {
		if len(v) > 1 {
			return m.errViolation(k)
		}
	}
	return nil
}

func (m *PrimaryMonitor) noPrimaryInOldState(k string, newState map[string][]*booleanProperty) error {
	switch len(newState[k]) {
	case 0, 1:
		return nil
	default:
		return m.errViolation(k)
	}
}

func (m *PrimaryMonitor) onePrimaryInOldState(k string, newState map[string][]*booleanProperty) error {
	switch len(newState[k]) {
	case 0, 1:
		return nil
	case 2:
		diff := 0
		for _, b := range newState[k] {
			if b == m.state[k][0] {
				_, err := b.Delete()
				if err != nil {
					return err
				}
			} else {
				diff++
			}
		}
		if diff > 1 {
			return m.errViolation(k)
		} else {
			return nil
		}
	default:
		return m.errViolation(k)
	}
}

func (m *PrimaryMonitor) refresh() map[string][]*booleanProperty {
	newState := make(map[string][]*booleanProperty)

	for _, container := range m.containerProps {
		newState[container.Attribute().ID()] = []*booleanProperty{}
		_ = container.ForEachChild(func(index int, complexProp core.Property) error {
			_ = complexProp.(core.Container).ForEachChild(func(index int, subProp core.Property) error {
				// Perform this direct check is OK because we made sure only one boolean sub property
				// can be marked as primary. See Attribute#MustValidate.
				if b, ok := subProp.(*booleanProperty); ok && subProp.Attribute().IsPrimary() {
					if true == subProp.Raw() {
						list := newState[container.Attribute().ID()]
						list = append(list, b)
						newState[container.Attribute().ID()] = list
					}
				}
				return nil
			})
			return nil
		})
	}
	return newState
}

func (m *PrimaryMonitor) ShouldVisit(property core.Property) bool {
	attr := property.Attribute()
	return attr.MultiValued() && attr.Type() == core.TypeComplex && attr.HasPrimarySubAttribute()
}

func (m *PrimaryMonitor) Visit(property core.Property) error {
	m.containerProps = append(m.containerProps, property.(core.Container))
	return nil
}

func (m *PrimaryMonitor) BeginChildren(container core.Container) {}

func (m *PrimaryMonitor) EndChildren(container core.Container) {}

func (m *PrimaryMonitor) errViolation(attrId string) error {
	return errors.InvalidValue("'%s' violates the rule that at most one primary boolean sub property may be true at all times.", attrId)
}
