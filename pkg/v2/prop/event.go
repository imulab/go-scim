package prop

// A single modification event.
type Event struct {
	typ    EventType
	source Property
	pre    interface{} // property value prior to event
}

// Type returns the type of the event
func (e Event) Type() EventType {
	return e.typ
}

// Source returns the Property that emits the event
func (e Event) Source() Property {
	return e.source
}

// PreModData optionally returns the Source's data before the modification. Note PreModData
// is not always available, hence should not be relied upon.
func (e Event) PreModData() interface{} {
	return e.pre
}

// ToEvents conveniently creates an Events package that contains this single event.
func (e *Event) ToEvents() *Events {
	return &Events{events: []*Event{e}}
}

// Type of an event
type EventType int

// NewFrom constructs an Event of this type.
func (et EventType) NewFrom(source Property, pre interface{}) *Event {
	ev := Event{typ: et, source: source, pre: pre}
	return &ev
}

// Type of modification events
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

// Events is the package of one or more Event. While a single modification call emits a single Event, the emitted
// event is wrapped in the Events package and passed along the trace stack for subscriber processing. Subscriber may
// modify or append to the events in this package, hence the necessity for this structure.
type Events struct {
	events []*Event
}

// Count returns the total number of events.
func (e *Events) Count() int {
	return len(e.events)
}

// ForEachEvent invokes callback function on each Event, any error aborts the process and is returned immediately.
func (e *Events) ForEachEvent(callback func(ev *Event) error) error {
	for _, ev := range e.events {
		if err := callback(ev); err != nil {
			return err
		}
	}
	return nil
}

// FindEvent searches and returns the first Event that matches the criteria, or nil if no such match.
func (e *Events) FindEvent(criteria func(ev *Event) bool) *Event {
	for _, ev := range e.events {
		if criteria(ev) {
			return ev
		}
	}
	return nil
}

// Append adds a new Event to the Events package.
func (e *Events) Append(ev *Event) {
	e.events = append(e.events, ev)
}
