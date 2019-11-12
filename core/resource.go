package core

type Resource struct {
	rt   *ResourceType
	base *complexProperty
}

func (r *Resource) Get(step *Step) (interface{}, error) {
	return r.base.Get(step)
}

func (r *Resource) Replace(step *Step, value interface{}) error {
	return r.base.Replace(step, value)
}

// Visit the properties contained in this resource in a DFS manner while considering the opinions
// of the provided visitor.
func (r *Resource) Visit(visitor Visitor) error {
	visitor.BeginComplex()
	for _, subProp := range r.base.subProps {
		if !visitor.ShouldVisit(subProp) {
			continue
		}

		err := subProp.(Visitable).VisitedBy(visitor)
		if err != nil {
			return err
		}
	}
	visitor.EndComplex()

	return nil
}