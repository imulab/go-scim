package crud

import (
	"github.com/imulab/go-scim/pkg/v2/crud/expr"
	"github.com/imulab/go-scim/pkg/v2/prop"
	"sort"
)

// Order for sorting
type SortOrder string

// Sort order defined in specification
const (
	SortDefault SortOrder = ""
	SortAsc     SortOrder = "ascending"
	SortDesc    SortOrder = "descending"
)

type (
	// Option to sort
	Sort struct {
		By    string
		Order SortOrder
	}
	// Option to include or exclude attributes in the return. At most one can be specified.
	Projection struct {
		Attributes         []string
		ExcludedAttributes []string
	}
	// Option to paginate.
	Pagination struct {
		StartIndex int // 1-based start index
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
		a   prop.Property
		b   prop.Property
		err error
	)

	a, err = SeekSortTarget(s.resources[i], s.by)
	if err != nil {
		switch s.dir {
		case SortDefault, SortAsc:
			return false
		case SortDesc:
			return true
		default:
			panic("invalid sortOrder")
		}
	}

	b, err = SeekSortTarget(s.resources[j], s.by)
	if err != nil {
		switch s.dir {
		case SortDefault, SortAsc:
			return true
		case SortDesc:
			return false
		default:
			panic("invalid sortOrder")
		}
	}

	if ltCapable, ok := a.(prop.LtCapable); !ok {
		return false
	} else {
		switch s.dir {
		case SortDefault, SortAsc:
			return ltCapable.LessThan(b.Raw())
		case SortDesc:
			return !ltCapable.LessThan(b.Raw())
		default:
			panic("invalid sortOrder")
		}
	}
}

func (s *sortWrapper) Swap(i, j int) {
	s.resources[i], s.resources[j] = s.resources[j], s.resources[i]
}
