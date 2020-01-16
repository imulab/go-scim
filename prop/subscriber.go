package prop

import (
	"sync"
)

// A subscriber to the event.
type Subscriber interface {
	// Be notified and react to the a series of events. Subscriber is allowed to modify the events list in order to
	// affect downstream subscribers.
	Notify(publisher Property, events *Events) error
}

var (
	subFactory     *subscriberFactory
	onceSubFactory sync.Once
)

type SubscriberFactoryFunc func(publisher Property, params map[string]interface{}) Subscriber

type subscriberFactory struct {
	constructors map[string]SubscriberFactoryFunc
}

// Register a subscriber constructor with the annotation.
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

// Return the subscriber factory to Register and Create subscribers using annotations.
func SubscriberFactory() *subscriberFactory {
	onceSubFactory.Do(func() {
		subFactory = &subscriberFactory{constructors: map[string]SubscriberFactoryFunc{}}
	})
	return subFactory
}
