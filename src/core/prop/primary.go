package prop

import (
	"github.com/imulab/go-scim/src/core"
)

func NewExclusivePrimaryPropertySubscriber() core.Subscriber {
	return &exclusivePrimarySubscriber{}
}

type exclusivePrimarySubscriber struct {}

func (s *exclusivePrimarySubscriber) supports(event *core.Event) bool {
	return event.Type() == core.EventNewValue &&
		event.Target().Attribute().IsPrimary() &&
		event.Target().Raw() == true
}

func (s *exclusivePrimarySubscriber) Notify(publisher core.Property, event *core.Event) error {
	if !s.supports(event) {
		return nil
	} else if publisher.Attribute().SingleValued() || publisher.Attribute().Type() != core.TypeComplex {
		// this subscriber should only be hooked on complex multiValued containers
		return nil
	}

	target := event.Target()
	nav := NewNavigator(publisher)

	for i := 0; i < publisher.(core.Container).CountChildren(); i++ {
		_, err := nav.FocusIndex(i)
		if err != nil {
			return err
		}
		{
			_, err := nav.FocusName(target.Attribute().Name())
			if err != nil {
				return err
			}
			if nav.Current() != target {
				if err := nav.Current().Delete(); err != nil {
					return err
				}
			}
			nav.Retract()
		}
		nav.Retract()
	}

	return nil
}
