package prop

import (
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/annotations"
)

// Create a subscriber that handles the exclusive primary problem. The subscriber listens on the EventAssigned event
// emitted by a property whose sub attribute is marked as primary. If the value of that property becomes "true", the
// subscriber will go through all elements of this property (assuming being a multiValued complex property) and delete
// all other sub properties that went by the same name of the event target. The end result would be that only one
// primary sub property will have the value "true".
func NewExclusivePrimarySubscriber() core.Subscriber {
	return &exclusivePrimarySubscriber{}
}

type exclusivePrimarySubscriber struct{}

func (s *exclusivePrimarySubscriber) supports(event *core.Event) bool {
	return event.Type() == core.EventAssigned &&
		event.Target().Attribute().IsPrimary() &&
		event.Target().Raw() == true
}

func (s *exclusivePrimarySubscriber) Notify(publisher core.Property, event *core.Event) error {
	if !s.supports(event) {
		return nil
	} else if publisher.Attribute().SingleValued() || publisher.Attribute().Type() != core.TypeComplex {
		return nil // this subscriber should only be hooked on complex multiValued containers
	}

	var (
		target    = event.Target()
		container = publisher.(core.Container)
	)
	for i := 0; i < container.CountChildren(); i++ {
		primaryProperty := container.ChildAtIndex(i).(core.Container).ChildAtIndex(target.Attribute().Name())
		if primaryProperty != target {
			if err := primaryProperty.Delete(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Create a subscriber that watches the state of a complex property. The subscriber listens to EventAssigned and
// EventUnassigned events from the sub properties of the complex property it subscribes to, and re-emit EventAssigned
// and EventUnassigned events on the complex property. For every EventAssigned event from sub properties, it re-emits
// EventAssigned events on the complex property; for every EventUnassigned event from sub properties, it checks if
// the complex property (container) has become unassigned and only emits EventUnassigned event when it became unassigned.
//
// This subscriber is useful when upstream subscribers need a summary of the container state as a whole.
func NewComplexStateSubscriber() core.Subscriber {
	return &complexStateSubscriber{}
}

type complexStateSubscriber struct{}

func (s *complexStateSubscriber) Notify(publisher core.Property, event *core.Event) (err error) {
	if publisher == event.Target() {
		// We will invoke Propagate on publisher later, avoid
		// creating a pub-sub dead loop.
		return
	}

	container, ok := publisher.(core.Container)
	if !ok {
		return
	}

	switch event.Type() {
	case core.EventAssigned:
		err = container.Propagate(core.EventAssigned.NewFrom(publisher))
	case core.EventUnassigned:
		if container.IsUnassigned() {
			err = container.Propagate(core.EventUnassigned.NewFrom(publisher))
		}
	}

	return
}

// Creates a subscriber that automatically compacts multiValued properties. This subscriber listens for EventUnassigned
// events from its elements (special attributes whose attribute id suffixes "$elem" on its multiValued container). If
// any element was unassigned, it calls the Compact method on the multiValued container in order to compact the collection.
func NewAutoCompactSubscriber() core.Subscriber {
	return &autoCompactSubscriber{}
}

type autoCompactSubscriber struct{}

func (s *autoCompactSubscriber) Notify(publisher core.Property, event *core.Event) error {
	if !event.Target().Attribute().IsElementAttributeOf(publisher.Attribute()) {
		return nil
	} else if event.Type() != core.EventUnassigned {
		return nil
	}

	container, ok := publisher.(core.Container)
	if !ok {
		return nil
	}

	container.Compact()
	return nil
}

// Create a new schema sync subscriber. The subscriber is designed to listen on the top level complex property for any
// emitted EventAssigned and EventUnassigned events from a complex sub property which serves as a container for
// attributes from one of the schema extensions in the resource type. Such property would have been annotated with
// "@StateSummary".
//
// When EventAssigned is emitted, it ensures the schema extension URN is present in the "schemas" attribute by adding
// the URN to it (note multiValued properties do not actually add a new item unless it wasn't part of the collection).
// When EventUnassigned is emitted, it ensures the schema extension URN is absent from the "schemas" attribute by
// removing it. The removal does not automatically compact the multiValued "schema" property. As a result, it will
// have an unassigned element in its collection. The compacting can be done by triggering the AutoCompactSubscriber.
func NewSchemaSyncSubscriber() core.Subscriber {
	return &schemaSyncSubscriber{}
}

type schemaSyncSubscriber struct{}

func (s *schemaSyncSubscriber) Notify(publisher core.Property, event *core.Event) (err error) {
	if !event.Target().Attribute().HasAnnotation(annotations.StateSummary) ||
		!event.Target().Attribute().HasAnnotation(annotations.SchemaExtensionRoot) {
		return
	}

	var schemas core.Property
	{
		container, ok := publisher.(core.Container)
		if !ok {
			return
		}
		schemas = container.ChildAtIndex("schemas")
		if schemas == nil {
			return
		} else if schemas.Attribute().ID() != "schemas" {
			return
		}
	}

	switch event.Type() {
	case core.EventAssigned:
		err = schemas.Add(event.Target().Attribute().Name())
	case core.EventUnassigned:
		err = schemas.(core.Container).ForEachChild(func(index int, child core.Property) error {
			if child.Raw() == event.Target().Attribute().Name() {
				return child.Delete()
			}
			return nil
		})
	}
	return
}

var (
	subscriberFactory = map[string]func() core.Subscriber{
		annotations.AutoCompact:      func() core.Subscriber { return NewAutoCompactSubscriber() },
		annotations.ExclusivePrimary: func() core.Subscriber { return NewExclusivePrimarySubscriber() },
		annotations.StateSummary:     func() core.Subscriber { return NewComplexStateSubscriber() },
		annotations.SyncSchema:       func() core.Subscriber { return NewSchemaSyncSubscriber() },
	}
)

// Add a subscriber factory function that corresponds to a given annotation.
func AddEventFactory(annotation string, factory func() core.Subscriber) {
	subscriberFactory[annotation] = factory
}

// Go through the registered annotations of a property attribute, and subscribe all corresponding subscribers.
func subscribeWithAnnotation(property core.Property) {
	property.Attribute().ForEachAnnotation(func(annotation string) {
		if f, ok := subscriberFactory[annotation]; ok {
			property.Subscribe(f())
		}
	})
}
