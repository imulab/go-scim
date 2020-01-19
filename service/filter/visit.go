package filter

import (
	"context"
	"github.com/elvsn/scim.go/prop"
	"github.com/elvsn/scim.go/spec"
)

func Visit(ctx context.Context, resource *prop.Resource, filter ByProperty) error {
	n := flexNavigator{stack: []prop.Property{resource.RootProperty()}}
	v := syncVisitor{
		resourceNav: &n,
		visitFunc: func(resourceNav prop.Navigator, referenceNav prop.Navigator) error {
			if !filter.Supports(resourceNav.Current().Attribute()) {
				return nil
			}
			return filter.Filter(ctx, resource.ResourceType(), resourceNav)
		},
	}
	return resource.Visit(&v)
}

func VisitWithRef(ctx context.Context, resource *prop.Resource, ref *prop.Resource, filter ByProperty) error {
	n := flexNavigator{stack: []prop.Property{resource.RootProperty()}}
	f := flexNavigator{stack: []prop.Property{ref.RootProperty()}}
	v := syncVisitor{
		resourceNav:  &n,
		referenceNav: &f,
		visitFunc: func(resourceNav prop.Navigator, referenceNav prop.Navigator) error {
			if !filter.Supports(resourceNav.Current().Attribute()) {
				return nil
			}
			return filter.FilterRef(ctx, resource.ResourceType(), resourceNav, referenceNav)
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
