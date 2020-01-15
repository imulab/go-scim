package prop

import (
	"github.com/elvsn/scim.go/internal/annotation"
	"github.com/elvsn/scim.go/spec"
	"sync"
)

// A subscriber to the event.
type Subscriber interface {
	// Be notified and react to the a series of events. Subscriber is allowed to modify the events list in order to
	// affect downstream subscribers.
	Notify(publisher Property, events []*Event) error
}

var (
	subFactory     *subscriberFactory
	onceSubFactory sync.Once
)

type subscriberFactory struct {
	constructors map[string]func(params map[string]interface{}) Subscriber
}

// Register a subscriber constructor with the annotation.
func (f *subscriberFactory) Register(annotation string, constructor func(params map[string]interface{}) Subscriber) {
	f.constructors[annotation] = constructor
}

// Create a new subscriber associated with the annotation, given the parameters. Return the create subscriber and a
// boolean indicating whether creation is successful.
func (f *subscriberFactory) Create(annotation string, params map[string]interface{}) (subscriber Subscriber, ok bool) {
	constructor, ok := f.constructors[annotation]
	if !ok {
		return
	}
	subscriber = constructor(params)
	return
}

// Return the subscriber factory to Register and Create subscribers using annotations.
func SubscriberFactory() *subscriberFactory {
	onceSubFactory.Do(func() {
		subFactory = &subscriberFactory{constructors: map[string]func(params map[string]interface{}) Subscriber{}}
	})
	return subFactory
}

// Subscriber to maintain at most one primary attribute that is true among sub properties.
type exclusivePrimarySubscriber struct{}

func (s *exclusivePrimarySubscriber) Notify(publisher Property, events []*Event) error {
	if !s.validPublisher(publisher) {
		return nil
	}

	ev := s.findPrimaryAssignedToTrueEvent(events)
	if ev == nil {
		return nil
	}

	for i := 0; i < publisher.countChildren(); i++ {
		ppNav := Navigate(publisher).At(i).Dot(ev.source.Attribute().Name()) // points to primary property
		if ppNav.Error() == nil && ppNav.Current() != ev.source {
			dev, err := ppNav.Current().Delete()
			if err != nil {
				return err
			}
			events = append(events, dev)
		}
	}

	return nil
}

func (s *exclusivePrimarySubscriber) validPublisher(publisher Property) bool {
	return publisher.Attribute().MultiValued() && publisher.Attribute().Type() == spec.TypeComplex
}

func (s *exclusivePrimarySubscriber) findPrimaryAssignedToTrueEvent(events []*Event) *Event {
	for _, ev := range events {
		if ev.typ != EventAssigned {
			continue
		}
		if _, ok := ev.source.Attribute().Annotation(annotation.Primary); !ok {
			continue
		}
		if ev.source.Raw() != true {
			continue
		}
		return ev
	}
	return nil
}

// todo complex state subscriber

// todo auto compact subscriber

// todo schema sync subscriber

func init() {
	epSub := exclusivePrimarySubscriber{}
	SubscriberFactory().Register(annotation.ExclusivePrimary, func(params map[string]interface{}) Subscriber {
		return &epSub
	})
}
