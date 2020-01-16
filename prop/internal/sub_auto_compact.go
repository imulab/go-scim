package internal

import (
	"github.com/elvsn/scim.go/internal/annotation"
	"github.com/elvsn/scim.go/prop"
)

func init() {
	acs := AutoCompactSubscriber{}
	prop.SubscriberFactory().Register(annotation.AutoCompact, func(_ prop.Property, _ map[string]interface{}) prop.Subscriber {
		return &acs
	})
}

// AutoCompactSubscriber automatically compacts the multiValued property.
//
// It is mounted by @AutoCompact annotation onto a multiValued property. If the mounted property is not multiValued,
// this subscriber does nothing.
//
// The subscriber reacts to unassigned events from its elements and invokes the hidden Compact API on the multiValued
// property.
type AutoCompactSubscriber struct{}

func (s *AutoCompactSubscriber) Notify(publisher prop.Property, events *prop.Events) error {
	if !s.validPublisher(publisher) {
		return nil
	}

	if !s.hasUnassignedEventFromElementAttribute(publisher, events) {
		return nil
	}

	if c, ok := publisher.(interface {
		Compact()
	}); ok {
		c.Compact()
	}

	return nil
}

func (s *AutoCompactSubscriber) validPublisher(publisher prop.Property) bool {
	return publisher.Attribute().MultiValued()
}

func (s *AutoCompactSubscriber) hasUnassignedEventFromElementAttribute(publisher prop.Property, events *prop.Events) bool {
	return events.FindEvent(func(ev *prop.Event) bool {
		return ev.Type() == prop.EventUnassigned && ev.Source().Attribute().IsElementAttributeOf(publisher.Attribute())
	}) != nil
}
