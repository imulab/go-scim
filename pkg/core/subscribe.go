package core

// type of an event
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

// Create a new event of this type. By default, the event will propagate.
func (t EventType) NewFrom(target Property) *Event {
	return &Event{
		typ:       t,
		target:    target,
		propagate: true,
	}
}

// Event payload
type Event struct {
	typ       EventType
	target    Property // property that emitted the event
	propagate bool     // if true, then propagate to parent
}

// Return the type of the event
func (e *Event) Type() EventType {
	return e.typ
}

// Return the target property of the event
func (e *Event) Target() Property {
	return e.target
}

func (e *Event) WillPropagate() bool {
	return e.propagate
}

// Set the propagate flag of this event to false, thus preventing
// this event to be further propagated upstream.
func (e *Event) StopPropagation() {
	e.propagate = false
}

// A subscriber to the event.
type Subscriber interface {
	// Be notified and react to the event. Implementations of this method
	// is allowed to return a non-nil error, which will be regarded as caller's
	// own error. Implementations are not supposed to block the Goroutine and
	// should keep the reaction quick.
	Notify(publisher Property, event *Event) error
}
