package filter

import (
	"context"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"github.com/imulab/go-scim/pkg/v2/spec"
)

// Visit performs a DFS visit on the resource and sequentially invokes the ByProperty filters on each visited property
// in the resource. Any visit or filtering error is returned.
func Visit(ctx context.Context, resource *prop.Resource, filters ...ByProperty) error {
	n := flexNavigator{stack: []prop.Property{resource.RootProperty()}}
	v := syncVisitor{
		resourceNav: &n,
		visitFunc: func(resourceNav prop.Navigator, referenceNav prop.Navigator) error {
			for _, filter := range filters {
				if !filter.Supports(resourceNav.Current().Attribute()) {
					continue
				}
				if err := filter.Filter(ctx, resource.ResourceType(), resourceNav); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return resource.Visit(&v)
}

// VisitWithRef performs a DFS visit on the resource and sequentially invokes the ByProperty filters on each visited
// property in the resource, along with the synchronized reference property. The synchronization is carried out with
// best effort, which means the reference property may be out of sync. Out of sync can happen when the resource has a
// property value that the reference resource does not have (i.e. Add) Caller need to test if
//	ref == nil || ref == outOfSync
// to determine if the reference is out of sync.
// Any visit or filtering error is returned.
func VisitWithRef(ctx context.Context, resource *prop.Resource, ref *prop.Resource, filters ...ByProperty) error {
	n := flexNavigator{stack: []prop.Property{resource.RootProperty()}}
	f := flexNavigator{stack: []prop.Property{ref.RootProperty()}}
	v := syncVisitor{
		resourceNav:  &n,
		referenceNav: &f,
		visitFunc: func(resourceNav prop.Navigator, referenceNav prop.Navigator) error {
			for _, filter := range filters {
				if !filter.Supports(resourceNav.Current().Attribute()) {
					continue
				}
				if err := filter.FilterRef(ctx, resource.ResourceType(), resourceNav, referenceNav); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return resource.Visit(&v)
}

type syncVisitor struct {
	resourceNav  *flexNavigator // flex navigator to be used in active mode
	referenceNav *flexNavigator // flex navigator to be used in passive (follow-along) mode
	visitFunc    func(resourceNav prop.Navigator, referenceNav prop.Navigator) error
}

func (v *syncVisitor) ShouldVisit(_ prop.Property) bool {
	return true
}

func (v *syncVisitor) Visit(property prop.Property) error {
	v.resourceNav.Push(property)
	if v.referenceNav != nil {
		if container := v.resourceNav.Last(); container != nil {
			if container.Attribute().MultiValued() {
				v.referenceNav.Where(func(child prop.Property) bool {
					return child.Matches(v.resourceNav.Current())
				})
			} else {
				v.referenceNav.Dot(property.Attribute().Name())
			}
		}
	}

	// Simple properties retract at the end of Visit, container properties retract in EndChildren
	if !property.Attribute().MultiValued() && property.Attribute().Type() != spec.TypeComplex {
		defer func() {
			v.resourceNav.Retract()
			if v.referenceNav != nil {
				v.referenceNav.Retract()
			}
		}()
	}

	return v.visitFunc(v.resourceNav, v.referenceNav)
}

func (v *syncVisitor) BeginChildren(_ prop.Property) {}

func (v *syncVisitor) EndChildren(_ prop.Property) {
	v.resourceNav.Retract()
	if v.referenceNav != nil {
		v.referenceNav.Retract()
	}
}
