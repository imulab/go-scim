package prop

import (
	"github.com/imulab/go-scim/pkg/v2/annotation"
	"github.com/imulab/go-scim/pkg/v2/spec"
	"sync"
)

// Subscriber attaches to Property and gets notified various state change events via Notify method.
//
// There is no explicit way of attaching to Property. All Subscriber implementations are automatically loaded onto
// Property via the annotation mechanism. A Subscriber can Register itself with the SubscriberFactory via an annotation.
// When a Property is being created, it will check if one or more of its attribute annotations are associated with
// Subscriber and ask SubscriberFactory to create an instance of Subscriber and attach the instance to itself.
//
// Implementations can choose to be either stateful or stateless. It is more memory efficient to create a single instance
// of stateless Subscriber. Yet, stateful subscribers might be more powerful as it allows user to specify parameters
// with the annotation, which gets passed in during initialization in the SubscriberFactory.
type Subscriber interface {
	// Be notified and react to the a series of events. Subscriber is allowed to modify the events list in order to
	// affect downstream subscribers.
	Notify(publisher Property, events *Events) error
}

// Return the subscriber factory to Register and Create subscribers using annotations.
func SubscriberFactory() *subscriberFactory {
	onceSubFactory.Do(func() {
		subFactory = &subscriberFactory{constructors: map[string]SubscriberFactoryFunc{}}
	})
	return subFactory
}

// Constructor function to initialize a Subscriber instance.
//
//	publisher is the property that is creating the Subscriber, and also what the Subscriber will eventually subscribe to.
//
//	params is the parameter specified with the annotation that was associated with the Subscriber type, it might be
//	useful when customizing Subscribers during initialization.
//
// Stateless Subscriber implementation may choose to return the same instance.
type SubscriberFactoryFunc func(publisher Property, params map[string]interface{}) Subscriber

var (
	subFactory     *subscriberFactory // subscriber factory singleton
	onceSubFactory sync.Once          // ensure only one subscriber factory instance
)

type subscriberFactory struct {
	constructors map[string]SubscriberFactoryFunc
}

// Register a subscriber constructor with the factory. The subscriber constructor will be associated with the annotation
// and gets invoked when Create is called with the annotation.
func (f *subscriberFactory) Register(annotation string, constructor SubscriberFactoryFunc) {
	f.constructors[annotation] = constructor
}

// Create a new subscriber associated with the annotation, given the parameters. Return the create subscriber and a
// boolean indicating whether creation is successful.
func (f *subscriberFactory) Create(annotation string, publisher Property, params map[string]interface{}) (subscriber Subscriber, ok bool) {
	constructor, ok := f.constructors[annotation]
	if !ok {
		return
	}
	subscriber = constructor(publisher, params)
	return
}

// AutoCompactSubscriber automatically compacts the multiValued property.
//
// It is mounted by @AutoCompact annotation onto a multiValued property. If the mounted property is not multiValued,
// this subscriber does nothing.
//
// The subscriber reacts to unassigned events from its elements and invokes the hidden Compact API on the multiValued
// property.
type AutoCompactSubscriber struct{}

func (s *AutoCompactSubscriber) Notify(publisher Property, events *Events) error {
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

func (s *AutoCompactSubscriber) validPublisher(publisher Property) bool {
	return publisher.Attribute().MultiValued()
}

func (s *AutoCompactSubscriber) hasUnassignedEventFromElementAttribute(publisher Property, events *Events) bool {
	return events.FindEvent(func(ev *Event) bool {
		return ev.Type() == EventUnassigned && ev.Source().Attribute().IsElementAttributeOf(publisher.Attribute())
	}) != nil
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

func (s *ExclusivePrimarySubscriber) Notify(publisher Property, events *Events) error {
	if !s.validPublisher(publisher) {
		return nil
	}

	ev := s.findPrimaryAssignedToTrueEvent(events)
	if ev == nil {
		return nil
	}

	nav := Navigate(publisher)
	return nav.ForEachChild(func(index int, child Property) error {
		defer func() {
			for nav.Current() != publisher {
				nav.Retract()
			}
		}()

		nav.At(index).Dot(ev.Source().Attribute().Name())
		if nav.HasError() {
			return nil
		}

		if nav.Current() == ev.Source() {
			return nil
		}

		// Here we want to submit events to the same event propagation flow that triggered this subscriber, hence,
		// we don't use nav.Delete() which triggers a new propagation flow. Rather, we append the individual
		// delete event to the events in this exact flow.
		dev, err := nav.Current().Delete()
		if err != nil {
			return err
		}
		events.Append(dev)

		return nil
	})
}

func (s *ExclusivePrimarySubscriber) validPublisher(publisher Property) bool {
	return publisher.Attribute().MultiValued() && publisher.Attribute().Type() == spec.TypeComplex
}

func (s *ExclusivePrimarySubscriber) findPrimaryAssignedToTrueEvent(events *Events) *Event {
	return events.FindEvent(func(ev *Event) bool {
		if ev.Type() != EventAssigned {
			return false
		}
		if _, ok := ev.Source().Attribute().Annotation(annotation.Primary); !ok {
			return false
		}
		return ev.Source().Raw() == true
	})
}

// SchemaSyncSubscriber automatically synchronizes the schema property with respect to data changes in the resource.
//
// It is mounted by @SyncSchema annotation onto the root of the property whose attribute is annotated with @Root. If
// not mounted onto @Root, this subscriber does nothing.
//
// The subscriber reacts to state change events from the root property of schema extension attributes. In other words,
// it reacts to events whose source attribute is annotated with @StateSummary and @SchemaExtensionRoot. It adds the
// schema extension id to schemas property on assigned events; and removes the schema extension id from schemas property
// on unassigned events.
//
// This subscriber does not attempt to compact the schemas property after element removal. This should be handled by
// AutoCompactSubscriber.
type SchemaSyncSubscriber struct{}

func (s *SchemaSyncSubscriber) Notify(publisher Property, events *Events) error {
	if !s.validPublisher(publisher) {
		return nil
	}

	return events.ForEachEvent(func(ev *Event) error {
		if !s.isStateSummaryOnExtensionRoot(ev) {
			return nil
		}

		nav := Navigate(publisher).Dot("schemas")
		if nav.HasError() {
			return nav.Error()
		}

		schemaId := ev.Source().Attribute().Name()
		switch ev.Type() {
		case EventAssigned:
			if nav.Add(schemaId).HasError() {
				return nav.Error()
			}
		case EventUnassigned:
			mark := nav.Current() // ensure every callback retracts to this position
			if err := nav.ForEachChild(func(index int, child Property) error {
				if child.Raw() != schemaId {
					return nil
				}

				defer func() {
					for nav.Current() != mark {
						nav.Retract()
					}
				}()

				if nav.At(index).Delete().HasError() {
					return nav.Error()
				}

				return nil
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *SchemaSyncSubscriber) validPublisher(publisher Property) bool {
	_, ok := publisher.Attribute().Annotation(annotation.Root)
	return ok
}

func (s *SchemaSyncSubscriber) isStateSummaryOnExtensionRoot(event *Event) bool {
	if _, ok := event.Source().Attribute().Annotation(annotation.SchemaExtensionRoot); !ok {
		return false
	}
	if _, ok := event.Source().Attribute().Annotation(annotation.StateSummary); !ok {
		return false
	}
	return true
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

func (s *ComplexStateSummarySubscriber) Notify(publisher Property, events *Events) error {
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
	if events.FindEvent(func(ev *Event) bool {
		return ev.Source() == publisher
	}) != nil {
		return nil
	}

	if wasAssigned && !s.assigned {
		events.Append(EventUnassigned.NewFrom(publisher, nil))
	} else if !wasAssigned && s.assigned {
		events.Append(EventAssigned.NewFrom(publisher, nil))
	}

	return nil
}

func (s *ComplexStateSummarySubscriber) validPublisher(publisher Property) bool {
	return !publisher.Attribute().MultiValued() && publisher.Attribute().Type() == spec.TypeComplex
}

func init() {
	acs := AutoCompactSubscriber{}
	SubscriberFactory().Register(annotation.AutoCompact, func(_ Property, _ map[string]interface{}) Subscriber {
		return &acs
	})

	eps := ExclusivePrimarySubscriber{}
	SubscriberFactory().Register(annotation.ExclusivePrimary, func(_ Property, _ map[string]interface{}) Subscriber {
		return &eps
	})

	s3 := SchemaSyncSubscriber{}
	SubscriberFactory().Register(annotation.SyncSchema, func(_ Property, _ map[string]interface{}) Subscriber {
		return &s3
	})

	SubscriberFactory().Register(annotation.StateSummary, func(publisher Property, _ map[string]interface{}) Subscriber {
		return &ComplexStateSummarySubscriber{assigned: !publisher.IsUnassigned()}
	})
}
