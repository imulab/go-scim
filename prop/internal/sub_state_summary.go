package internal

import (
	"github.com/elvsn/scim.go/internal/annotation"
	"github.com/elvsn/scim.go/prop"
	"github.com/elvsn/scim.go/spec"
)

func init() {
	prop.SubscriberFactory().Register(annotation.StateSummary, func(publisher prop.Property, _ map[string]interface{}) prop.Subscriber {
		return &ComplexStateSummarySubscriber{assigned: !publisher.IsUnassigned()}
	})
}

// ComplexStateSummarySubscriber summarizes the state changes of the sub properties of a complex property and generate
// new event to describe the inferred state change on the complex property.
//
// It is mounted by @StateSummary annotation onto a complex property. If the mounted property is not complex, this
// subscriber does nothing.
//
// The subscriber reacts to state change events from its sub properties and computes a new assigned state for the complex
// property. The new state is compared to an old cached assigned state in order to determine state change for the complex
// property. It also reacts to direct state changes on the complex property itself, in which case, the subscriber simply
// updates the state without generating new events.
type ComplexStateSummarySubscriber struct {
	assigned bool
}

func (s *ComplexStateSummarySubscriber) Notify(publisher prop.Property, events *prop.Events) error {
	if !s.validPublisher(publisher) {
		return nil
	}

	// update and record pre and post state for any events because
	// they originate either from the publisher property itself or
	// sub properties.
	wasAssigned := s.assigned
	s.assigned = !publisher.IsUnassigned()

	// retire early if we have an event from the publisher itself.
	// since we already have the mod event and state has already been updated,
	// there is no need to generate a new event.
	if events.FindEvent(func(ev *prop.Event) bool {
		return ev.Source() == publisher
	}) != nil {
		return nil
	}

	if wasAssigned && !s.assigned {
		events.Append(prop.EventUnassigned.NewFrom(publisher, nil))
	} else if !wasAssigned && s.assigned {
		events.Append(prop.EventAssigned.NewFrom(publisher, nil))
	}

	return nil
}

func (s *ComplexStateSummarySubscriber) validPublisher(publisher prop.Property) bool {
	return !publisher.Attribute().MultiValued() && publisher.Attribute().Type() == spec.TypeComplex
}
