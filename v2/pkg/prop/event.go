package prop

// A single event
type Event struct {
	typ    EventType
	source Property
	pre    interface{} // property value prior to event
}

func (e Event) Type() EventType {
	return e.typ
}

func (e Event) Source() Property {
	return e.source
}

func (e Event) PreModData() interface{} {
	return e.pre
}

func (e *Event) ToEvents() *Events {
	return &Events{events: []*Event{e}}
}

// Type of an event
type EventType int

func (et EventType) NewFrom(source Property, pre interface{}) *Event {
	ev := Event{typ: et, source: source, pre: pre}
	return &ev
}

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

type Events struct {
	events []*Event
}

func (e *Events) Count() int {
	return len(e.events)
}

func (e *Events) ForEachEvent(callback func(ev *Event) error) error {
	for _, ev := range e.events {
		if err := callback(ev); err != nil {
			return err
		}
	}
	return nil
}

func (e *Events) FindEvent(criteria func(ev *Event) bool) *Event {
	for _, ev := range e.events {
		if criteria(ev) {
			return ev
		}
	}
	return nil
}

func (e *Events) Append(ev *Event) {
	e.events = append(e.events, ev)
}
