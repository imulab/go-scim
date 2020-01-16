package internal

import (
	"errors"
	"github.com/elvsn/scim.go/internal/annotation"
	"github.com/elvsn/scim.go/prop"
	"github.com/elvsn/scim.go/spec"
)

func init() {
	s3 := SchemaSyncSubscriber{}
	prop.SubscriberFactory().Register(annotation.SyncSchema, func(_ prop.Property, _ map[string]interface{}) prop.Subscriber {
		return &s3
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

func (s *SchemaSyncSubscriber) Notify(publisher prop.Property, events *prop.Events) error {
	if !s.validPublisher(publisher) {
		return nil
	}

	return events.ForEachEvent(func(ev *prop.Event) error {
		if !s.isStateSummaryOnExtensionRoot(ev) {
			return nil
		}

		nav := prop.Navigate(publisher).Dot("schemas")
		if err := nav.Error(); err != nil {
			return err
		}

		schemaId := ev.Source().Attribute().Name()
		switch ev.Type() {
		case prop.EventAssigned:
			if err := nav.Add(schemaId); err != nil {
				return err
			}
		case prop.EventUnassigned:
			i := 0
			for {
				nav.At(i) // focus on i-th schema
				if err := nav.Error(); err != nil {
					if errors.Unwrap(err) == spec.ErrNoTarget {
						break
					}
					return err
				}

				if nav.Current().Raw() == schemaId {
					if err := nav.Delete(); err != nil {
						return err
					}
					break
				}

				nav.Retract() // retract At
				i++           // goto next schema
			}
		}

		return nil
	})
}

func (s *SchemaSyncSubscriber) validPublisher(publisher prop.Property) bool {
	_, ok := publisher.Attribute().Annotation(annotation.Root)
	return ok
}

func (s *SchemaSyncSubscriber) isStateSummaryOnExtensionRoot(event *prop.Event) bool {
	if _, ok := event.Source().Attribute().Annotation(annotation.SchemaExtensionRoot); !ok {
		return false
	}
	if _, ok := event.Source().Attribute().Annotation(annotation.StateSummary); !ok {
		return false
	}
	return true
}
