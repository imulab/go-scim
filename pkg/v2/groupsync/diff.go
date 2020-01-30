package groupsync

import (
	"github.com/imulab/go-scim/pkg/v2/prop"
)

const (
	fieldMembers = "members"
	fieldValue   = "value"
)

// Compare compares the two snapshots of two group resources before and after the
// modification and reports their differences in membership. At least one
// of before and after should be non-nil. When before is nil, all members
// of the after resource are considered to have just joined; when after
// is nil, all members of the before resource are considered to have just left.
func Compare(before *prop.Resource, after *prop.Resource) *Diff {
	if before == nil && after == nil {
		panic("at least one of before and after should be non-nil")
	}

	var (
		beforeIds = map[string]struct{}{}
		afterIds  = map[string]struct{}{}
	)
	for _, t := range []struct {
		resource  *prop.Resource
		collector map[string]struct{}
	}{
		{resource: before, collector: beforeIds},
		{resource: after, collector: afterIds},
	} {
		if t.resource == nil {
			continue
		}

		members, _ := t.resource.RootProperty().ChildAtIndex(fieldMembers)
		_ = members.ForEachChild(func(index int, child prop.Property) error {
			value, _ := child.ChildAtIndex(fieldValue)
			if value != nil && !value.IsUnassigned() {
				t.collector[value.Raw().(string)] = struct{}{}
			}
			return nil
		})
	}

	diff := new(Diff)
	for k := range beforeIds {
		if _, ok := afterIds[k]; !ok {
			diff.addLeft(k)
		}
	}
	for k := range afterIds {
		if _, ok := beforeIds[k]; !ok {
			diff.addJoined(k)
		}
	}
	return diff
}

// Diff reports the difference between the members of two group resources.
type Diff struct {
	joined map[string]struct{}
	left   map[string]struct{}
}

func (d *Diff) addJoined(id string) {
	if d.joined == nil {
		d.joined = map[string]struct{}{}
	}
	d.joined[id] = struct{}{}
}

func (d *Diff) addLeft(id string) {
	if d.left == nil {
		d.left = map[string]struct{}{}
	}
	d.left[id] = struct{}{}
}

// ForEachJoined iterates all member ids that joined the group and invoke the callback.
func (d *Diff) ForEachJoined(callback func(id string)) {
	for k := range d.joined {
		callback(k)
	}
}

// ForEachLeft iterates all member ids that left the group and invoke the callback.
func (d *Diff) ForEachLeft(callback func(id string)) {
	for k := range d.left {
		callback(k)
	}
}

// CountJoined returns the total number of new members that joined the group.
func (d *Diff) CountJoined() int {
	return len(d.joined)
}

// CountLeft returns the total number of members that just left the group.
func (d *Diff) CountLeft() int {
	return len(d.left)
}
