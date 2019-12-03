package core

const (
	_ EventType = iota
	EventNewValue
	EventDelete
)

type (
	EventType int
	Event     struct {
		typ       EventType
		target    Property
		propagate bool
	}
	// A subscriber to the event.
	Subscriber interface {
		// Be notified and react to the event. Implementations of this method
		// is allowed to return a non-nil error, which will be regarded as caller's
		// own error. Implementations are not supposed to block the Goroutine and
		// should keep the reaction quick.
		Notify(event *Event) error
	}
)

// Create a new event. By default, the event will propagate.
func NewEvent(eventType EventType, target Property) *Event {
	return &Event{
		typ:       eventType,
		target:    target,
		propagate: true,
	}
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
