package internal

import (
	"errors"
	"github.com/elvsn/scim.go/internal/annotation"
	"github.com/elvsn/scim.go/prop"
	"github.com/elvsn/scim.go/spec"
)

func init() {
	eps := ExclusivePrimarySubscriber{}
	prop.SubscriberFactory().Register(annotation.ExclusivePrimary, func(_ prop.Property, _ map[string]interface{}) prop.Subscriber {
		return &eps
	})
}

// ExclusivePrimarySubscriber automatically turns off the true-valued primary sub property when another primary sub property
// is set to true.
//
// It is mounted by @ExclusivePrimary annotation onto the a multiValued complex property whose sub property contains a
// boolean property which is annotated @Primary. This boolean property is referred to as the primary property.
//
// The subscriber reacts to assigned events from the primary property. If the event reports a primary property has a new
// value of true, this subscriber goes through all primary properties and turn off the old true value. The result is that
// at most one primary property will have the value of true.
type ExclusivePrimarySubscriber struct{}

func (s *ExclusivePrimarySubscriber) Notify(publisher prop.Property, events *prop.Events) error {
	if !s.validPublisher(publisher) {
		return nil
	}

	ev := s.findPrimaryAssignedToTrueEvent(events)
	if ev == nil {
		return nil
	}

	var (
		i   = 0
		nav = prop.Navigate(publisher)
	)
	for {
		nav.At(i) // focus on i-th element
		if err := nav.Error(); err != nil {
			if errors.Unwrap(err) == spec.ErrNoTarget {
				break
			}
			return err
		}

		nav.Dot(ev.Source().Attribute().Name()) // focus on the primary attribute of i-th element
		if nav.Error() == nil && nav.Current() != ev.Source() {
			// Here we want to submit events to the same event propagation flow that triggered this subscriber, hence,
			// we don't use nav.Delete() which triggers a new propagation flow. Rather, we append the individual
			// delete event to the events in this exact flow.
			dev, err := nav.Current().Delete()
			if err != nil {
				return err
			}
			events.Append(dev)
		}

		nav.Retract() // retract the Dot
		nav.Retract() // retract the At

		i++ // goto next element
	}

	return nil
}

func (s *ExclusivePrimarySubscriber) validPublisher(publisher prop.Property) bool {
	return publisher.Attribute().MultiValued() && publisher.Attribute().Type() == spec.TypeComplex
}

func (s *ExclusivePrimarySubscriber) findPrimaryAssignedToTrueEvent(events *prop.Events) *prop.Event {
	return events.FindEvent(func(ev *prop.Event) bool {
		if ev.Type() != prop.EventAssigned {
			return false
		}
		if _, ok := ev.Source().Attribute().Annotation(annotation.Primary); !ok {
			return false
		}
		return ev.Source().Raw() == true
	})
}
