package prop

// A single event
type Event struct {
	typ    EventType
	source Property
	pre    interface{} // property value prior to event
}

// Type of an event
type EventType int

const (
	_ EventType = iota
	// Event that new value has been assigned to the property.
	// The event entails that the property is now in an "assigned" state, such that
	// calling IsUnassigned shall return false. It is only emitted when value introduced
	// to the property is different than the previous value, otherwise, no event would be
	// emitted.
	EventAssigned
	// Event that property value was deleted and now is an "unassigned" state. Calling
	// IsUnassigned shall return true. It is only emitted that value was deleted from property
	// when it was in an "assigned" state, otherwise, no event would be emitted.
	EventUnassigned
)

// A subscriber to the event.
type Subscriber interface {
	// Be notified and react to the a series of events. Subscriber is allowed to modify the events list in order to
	// affect downstream subscribers.
	Notify(publisher Property, events []*Event) error
}
