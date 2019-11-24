package core

import "strings"

type Resource struct {
	rt   *ResourceType
	base *complexProperty
}

// Convenience method to get the id property value from the resource
func (r *Resource) GetID() (string, error) {
	v, err := r.base.Get(Steps.NewPath("id"))
	if err != nil {
		return "", err
	}

	id, ok := v.(string)
	if !ok {
		return "", Errors.InvalidValue("id has wrong type")
	}

	return id, nil
}

func (r *Resource) GetResourceType() *ResourceType {
	return r.rt
}

func (r *Resource) Get(step *Step) (interface{}, error) {
	return r.base.Get(r.skipResourceUrn(step))
}

func (r *Resource) Add(step *Step, value interface{}) error {
	return r.base.Add(r.skipResourceUrn(step), value)
}

func (r *Resource) Replace(step *Step, value interface{}) error {
	return r.base.Replace(r.skipResourceUrn(step), value)
}

func (r *Resource) Evaluate(queryRoot *Step) (bool, error) {
	return r.base.Evaluate(queryRoot)
}

func (r *Resource) skipResourceUrn(step *Step) *Step {
	if step == nil {
		return nil
	}

	if strings.ToLower(r.rt.Id) == strings.ToLower(step.Token) {
		return step.Next
	}

	return step
}

// Visit the properties contained in this resource in a DFS manner while considering the opinions
// of the provided visitor.
func (r *Resource) Visit(visitor Visitor) error {
	visitor.BeginComplex(r.base)
	for _, subProp := range r.base.orderedSubProperties() {
		if !visitor.ShouldVisit(subProp) {
			continue
		}

		err := subProp.(Visitable).VisitedBy(visitor)
		if err != nil {
			return err
		}
	}
	visitor.EndComplex(r.base)

	return nil
}
