package internal

import "github.com/elvsn/scim.go/prop"

// Internal implementation of Subscriber used in tests.
type recordingSubscriber struct {
	events *prop.Events
}

func (s *recordingSubscriber) Notify(_ prop.Property, events *prop.Events) error {
	s.events = events
	return nil
}
