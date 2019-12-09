package crud

import (
	"github.com/imulab/go-scim/pkg/core/expr"
	"github.com/imulab/go-scim/pkg/core/prop"
	"sort"
)

type SortOrder int

const (
	SortAsc SortOrder = iota
	SortDesc
)

type (
	Sort struct {
		By    string
		Order SortOrder
	}
	// Include or exclude attributes in the return. At most one can be specified.
	Projection struct {
		Attributes         []string
		ExcludedAttributes []string
	}
	Pagination struct {
		// 1-based start index
		StartIndex int
		Count      int
	}
)

// Sort the given list of resources according to the sort options.
func (s Sort) Sort(resources []*prop.Resource) error {
	if len(resources) <= 1 {
		return nil
	} else if len(s.By) == 0 {
		return nil
	}

	head, err := expr.CompilePath(s.By)
	if err != nil {
		return err
	}

	sort.Sort(&sortWrapper{
		resources: resources,
		by:        head,
		dir:       s.Order,
	})
	return nil
}

type sortWrapper struct {
	by        *expr.Expression
	dir       SortOrder
	resources []*prop.Resource
}

func (s *sortWrapper) Len() int {
	return len(s.resources)
}

func (s *sortWrapper) Less(i, j int) bool {
	var (
		iProp prop.Property
		jProp prop.Property
	)

	_ = traverse(s.resources[i].NewNavigator(), s.by, func(target prop.Property) error {
		iProp = target
		return nil
	})
	if iProp.IsUnassigned() {
		return false
	}

	_ = traverse(s.resources[j].NewNavigator(), s.by, func(target prop.Property) error {
		jProp = target
		return nil
	})
	if jProp.IsUnassigned() {
		return true
	}

	less, _ := iProp.LessThan(jProp.Raw())
	if s.dir == SortAsc {
		return less
	} else {
		return !less
	}
}

func (s *sortWrapper) Swap(i, j int) {
	s.resources[i], s.resources[j] = s.resources[j], s.resources[i]
}
